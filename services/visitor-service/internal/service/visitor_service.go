package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/visitor-service/internal/model"
	"github.com/societykro/visitor-service/internal/repository"
)

var (
	// ErrVisitorNotFound is returned when a visitor record does not exist.
	ErrVisitorNotFound = errors.New("visitor not found")
	// ErrOTPExpired is returned when the OTP has expired.
	ErrOTPExpired = errors.New("otp expired")
	// ErrAlreadyProcessed is returned when a visitor has already been approved or denied.
	ErrAlreadyProcessed = errors.New("visitor already processed")
)

// VisitorService handles all visitor business logic.
type VisitorService struct {
	repo *repository.VisitorRepository
	bus  *events.Bus
}

// NewVisitorService creates a new VisitorService.
func NewVisitorService(repo *repository.VisitorRepository, bus *events.Bus) *VisitorService {
	return &VisitorService{repo: repo, bus: bus}
}

// LogVisitor creates a new visitor entry at the gate.
func (s *VisitorService) LogVisitor(ctx context.Context, req model.CreateVisitorRequest, loggedBy, societyID uuid.UUID) (*model.Visitor, error) {
	flatID, err := uuid.Parse(req.FlatID)
	if err != nil {
		return nil, fmt.Errorf("invalid flat_id: %w", err)
	}

	vis := &model.Visitor{
		SocietyID: societyID,
		FlatID:    flatID,
		Name:      req.Name,
		Purpose:   req.Purpose,
		LoggedBy:  &loggedBy,
	}
	if req.VehicleNumber != "" {
		vis.VehicleNumber = &req.VehicleNumber
	}
	if req.Phone != "" {
		vis.Phone = &req.Phone
	}

	created, err := s.repo.Create(ctx, vis)
	if err != nil {
		return nil, fmt.Errorf("log visitor: %w", err)
	}

	if err := s.bus.Publish(events.SubjectVisitorLogged, "visitor.logged", created); err != nil {
		logger.Log.Error().Err(err).Str("visitor_id", created.ID.String()).Msg("Failed to publish visitor.logged event")
	}

	return created, nil
}

// PreApprove generates a 6-digit OTP for a visitor and stores it.
func (s *VisitorService) PreApprove(ctx context.Context, req model.PreApproveRequest, userID, societyID uuid.UUID, flatID uuid.UUID) (*model.Visitor, string, error) {
	targetFlatID := flatID
	if req.FlatID != "" {
		parsed, err := uuid.Parse(req.FlatID)
		if err != nil {
			return nil, "", fmt.Errorf("invalid flat_id: %w", err)
		}
		targetFlatID = parsed
	}

	// If no flat ID provided, look up from user's membership
	if targetFlatID == uuid.Nil {
		var fID *uuid.UUID
		s.repo.Pool().QueryRow(ctx,
			`SELECT flat_id FROM user_society_membership WHERE user_id = $1 AND society_id = $2 AND is_active = true LIMIT 1`,
			userID, societyID).Scan(&fID)
		if fID != nil {
			targetFlatID = *fID
		} else {
			return nil, "", fmt.Errorf("no flat assigned to your membership, provide flat_id")
		}
	}

	vis := &model.Visitor{
		SocietyID: societyID,
		FlatID:    targetFlatID,
		Name:      req.Name,
		Purpose:   req.Purpose,
	}
	if req.VehicleNumber != "" {
		vis.VehicleNumber = &req.VehicleNumber
	}
	if req.Phone != "" {
		vis.Phone = &req.Phone
	}

	created, err := s.repo.Create(ctx, vis)
	if err != nil {
		return nil, "", fmt.Errorf("create visitor for pre-approve: %w", err)
	}

	otp, err := generateOTP()
	if err != nil {
		return nil, "", fmt.Errorf("generate otp: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	if err := s.repo.SetOTP(ctx, created.ID, otp, expiresAt); err != nil {
		return nil, "", fmt.Errorf("set otp: %w", err)
	}

	created.Status = "pending" // pre-approved visitors stay pending until gate check-in
	created.OTPCode = &otp
	created.OTPExpiresAt = &expiresAt

	return created, otp, nil
}

// Approve approves a visitor entry and publishes an event.
func (s *VisitorService) Approve(ctx context.Context, visitorID, approvedBy uuid.UUID, via string) error {
	vis, err := s.repo.FindByID(ctx, visitorID)
	if err != nil {
		return err
	}
	if vis == nil {
		return ErrVisitorNotFound
	}
	if vis.Status != "pending" && vis.Status != "approved" {
		return ErrAlreadyProcessed
	}

	if via == "" {
		via = "app"
	}
	if err := s.repo.UpdateStatus(ctx, visitorID, "checked_in", &approvedBy, &via); err != nil {
		return err
	}

	if err := s.bus.Publish(events.SubjectVisitorApproved, "visitor.approved", map[string]interface{}{
		"visitor_id": visitorID,
		"society_id": vis.SocietyID,
		"flat_id":    vis.FlatID,
		"name":       vis.Name,
	}); err != nil {
		logger.Log.Error().Err(err).Str("visitor_id", visitorID.String()).Msg("Failed to publish visitor.approved event")
	}

	return nil
}

// Deny denies a visitor entry and publishes an event.
func (s *VisitorService) Deny(ctx context.Context, visitorID, deniedBy uuid.UUID, reason string) error {
	vis, err := s.repo.FindByID(ctx, visitorID)
	if err != nil {
		return err
	}
	if vis == nil {
		return ErrVisitorNotFound
	}
	if vis.Status != "pending" && vis.Status != "approved" {
		return ErrAlreadyProcessed
	}

	if err := s.repo.UpdateStatus(ctx, visitorID, "denied", &deniedBy, nil); err != nil {
		return err
	}
	if reason != "" {
		if err := s.repo.SetDenyReason(ctx, visitorID, reason); err != nil {
			logger.Log.Error().Err(err).Msg("Failed to set deny reason")
		}
	}

	if err := s.bus.Publish(events.SubjectVisitorDenied, "visitor.denied", map[string]interface{}{
		"visitor_id": visitorID,
		"society_id": vis.SocietyID,
		"flat_id":    vis.FlatID,
		"name":       vis.Name,
		"reason":     reason,
	}); err != nil {
		logger.Log.Error().Err(err).Str("visitor_id", visitorID.String()).Msg("Failed to publish visitor.denied event")
	}

	return nil
}

// Checkout marks a visitor as checked out.
func (s *VisitorService) Checkout(ctx context.Context, visitorID uuid.UUID) error {
	vis, err := s.repo.FindByID(ctx, visitorID)
	if err != nil {
		return err
	}
	if vis == nil {
		return ErrVisitorNotFound
	}

	return s.repo.UpdateStatus(ctx, visitorID, "checked_out", nil, nil)
}

// VerifyOTP checks an OTP at the gate and auto-approves the visitor.
func (s *VisitorService) VerifyOTP(ctx context.Context, societyID uuid.UUID, otpCode string) (*model.Visitor, error) {
	vis, err := s.repo.VerifyOTP(ctx, societyID, otpCode)
	if err != nil {
		return nil, err
	}
	if vis == nil {
		return nil, ErrVisitorNotFound
	}

	if vis.OTPExpiresAt != nil && vis.OTPExpiresAt.Before(time.Now()) {
		return nil, ErrOTPExpired
	}

	via := "otp"
	if err := s.repo.UpdateStatus(ctx, vis.ID, "checked_in", nil, &via); err != nil {
		return nil, err
	}
	vis.Status = "checked_in"

	return vis, nil
}

// GetByID returns a visitor with full details.
func (s *VisitorService) GetByID(ctx context.Context, id uuid.UUID) (*model.Visitor, error) {
	vis, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if vis == nil {
		return nil, ErrVisitorNotFound
	}
	return vis, nil
}

// List returns filtered and paginated visitors.
func (s *VisitorService) List(ctx context.Context, filter model.VisitorListFilter) ([]model.Visitor, error) {
	visitors, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if visitors == nil {
		visitors = []model.Visitor{}
	}
	return visitors, nil
}

// CreatePass creates a recurring visitor pass.
func (s *VisitorService) CreatePass(ctx context.Context, req model.CreatePassRequest, userID, societyID, flatID uuid.UUID) (*model.VisitorPass, error) {
	validFrom, err := time.Parse(time.RFC3339, req.ValidFrom)
	if err != nil {
		return nil, fmt.Errorf("invalid valid_from: %w", err)
	}
	validUntil, err := time.Parse(time.RFC3339, req.ValidUntil)
	if err != nil {
		return nil, fmt.Errorf("invalid valid_until: %w", err)
	}

	pass := &model.VisitorPass{
		FlatID:      flatID,
		SocietyID:   societyID,
		VisitorName: req.VisitorName,
		Purpose:     req.Purpose,
		ValidFrom:   validFrom,
		ValidUntil:  validUntil,
		CreatedBy:   userID,
	}
	if req.Phone != "" {
		pass.Phone = &req.Phone
	}
	if req.VehicleNumber != "" {
		pass.VehicleNumber = &req.VehicleNumber
	}

	return s.repo.CreatePass(ctx, pass)
}

// ListPasses returns active passes for a flat.
func (s *VisitorService) ListPasses(ctx context.Context, flatID uuid.UUID) ([]model.VisitorPass, error) {
	passes, err := s.repo.ListPasses(ctx, flatID)
	if err != nil {
		return nil, err
	}
	if passes == nil {
		passes = []model.VisitorPass{}
	}
	return passes, nil
}

// DeletePass removes a visitor pass.
func (s *VisitorService) DeletePass(ctx context.Context, passID uuid.UUID) error {
	return s.repo.DeletePass(ctx, passID)
}

// ListActive returns currently checked-in visitors for a society.
func (s *VisitorService) ListActive(ctx context.Context, societyID uuid.UUID) ([]model.Visitor, error) {
	visitors, err := s.repo.ListActive(ctx, societyID)
	if err != nil {
		return nil, err
	}
	if visitors == nil {
		visitors = []model.Visitor{}
	}
	return visitors, nil
}

// generateOTP produces a cryptographically random 6-digit code.
func generateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(900000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()+100000), nil
}
