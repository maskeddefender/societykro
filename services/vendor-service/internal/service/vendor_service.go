package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/vendor-service/internal/model"
	"github.com/societykro/vendor-service/internal/repository"
)

// Sentinel errors for vendor service operations.
var (
	ErrVendorNotFound       = errors.New("vendor not found")
	ErrDomesticHelpNotFound = errors.New("domestic help not found")
	ErrAttendanceNotFound   = errors.New("attendance record not found")
)

// VendorService handles all vendor and domestic help business logic.
type VendorService struct {
	repo *repository.VendorRepository
	bus  *events.Bus
}

// NewVendorService creates a new VendorService.
func NewVendorService(repo *repository.VendorRepository, bus *events.Bus) *VendorService {
	return &VendorService{repo: repo, bus: bus}
}

// --------------- Vendor Operations ---------------

// CreateVendor registers a new vendor for a society.
func (s *VendorService) CreateVendor(ctx context.Context, req model.CreateVendorRequest, societyID uuid.UUID) (*model.Vendor, error) {
	v := &model.Vendor{
		SocietyID:     societyID,
		Name:          req.Name,
		CompanyName:   req.CompanyName,
		Phone:         req.Phone,
		WhatsappPhone: req.WhatsappPhone,
		Category:      req.Category,
		SubCategory:   req.SubCategory,
		Address:       req.Address,
	}

	created, err := s.repo.CreateVendor(ctx, v)
	if err != nil {
		return nil, fmt.Errorf("create vendor: %w", err)
	}

	if err := s.bus.Publish("vendor.created", "vendor.created", map[string]interface{}{
		"vendor_id":  created.ID,
		"society_id": created.SocietyID,
		"category":   created.Category,
	}); err != nil {
		logger.Log.Error().Err(err).Str("vendor_id", created.ID.String()).Msg("Failed to publish vendor.created event")
	}

	return created, nil
}

// GetVendorByID returns a vendor by its ID.
func (s *VendorService) GetVendorByID(ctx context.Context, id uuid.UUID) (*model.Vendor, error) {
	v, err := s.repo.FindVendorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, ErrVendorNotFound
	}
	return v, nil
}

// ListVendors returns filtered and paginated vendors for a society.
func (s *VendorService) ListVendors(ctx context.Context, filter model.VendorListFilter) ([]model.Vendor, error) {
	vendors, err := s.repo.ListVendors(ctx, filter)
	if err != nil {
		return nil, err
	}
	if vendors == nil {
		vendors = []model.Vendor{}
	}
	return vendors, nil
}

// UpdateVendor updates vendor fields.
func (s *VendorService) UpdateVendor(ctx context.Context, id uuid.UUID, req model.UpdateVendorRequest) error {
	v, err := s.repo.FindVendorByID(ctx, id)
	if err != nil {
		return err
	}
	if v == nil {
		return ErrVendorNotFound
	}
	return s.repo.UpdateVendor(ctx, id, req)
}

// DeleteVendor performs a soft delete on a vendor.
func (s *VendorService) DeleteVendor(ctx context.Context, id uuid.UUID) error {
	v, err := s.repo.FindVendorByID(ctx, id)
	if err != nil {
		return err
	}
	if v == nil {
		return ErrVendorNotFound
	}
	return s.repo.DeleteVendor(ctx, id)
}

// UpdateVendorStats increments job count and updates average rating.
func (s *VendorService) UpdateVendorStats(ctx context.Context, id uuid.UUID, rating float64, completed bool) error {
	return s.repo.UpdateVendorStats(ctx, id, rating, completed)
}

// --------------- Domestic Help Operations ---------------

// CreateDomesticHelp registers a new domestic helper for a society.
func (s *VendorService) CreateDomesticHelp(ctx context.Context, req model.CreateDomesticHelpRequest, societyID uuid.UUID) (*model.DomesticHelp, error) {
	h := &model.DomesticHelp{
		SocietyID:   societyID,
		Name:        req.Name,
		Phone:       req.Phone,
		PhotoURL:    req.PhotoURL,
		Role:        req.Role,
		EntryMethod: req.EntryMethod,
	}

	created, err := s.repo.CreateDomesticHelp(ctx, h)
	if err != nil {
		return nil, fmt.Errorf("create domestic help: %w", err)
	}
	return created, nil
}

// GetDomesticHelpByID returns a domestic help record by ID.
func (s *VendorService) GetDomesticHelpByID(ctx context.Context, id uuid.UUID) (*model.DomesticHelp, error) {
	h, err := s.repo.FindDomesticHelpByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, ErrDomesticHelpNotFound
	}
	return h, nil
}

// ListDomesticHelp returns domestic help filtered by society or flat.
func (s *VendorService) ListDomesticHelp(ctx context.Context, societyID uuid.UUID, flatID *uuid.UUID) ([]model.DomesticHelp, error) {
	if flatID != nil {
		helpers, err := s.repo.ListDomesticHelpByFlat(ctx, *flatID)
		if err != nil {
			return nil, err
		}
		if helpers == nil {
			helpers = []model.DomesticHelp{}
		}
		return helpers, nil
	}

	helpers, err := s.repo.ListDomesticHelpBySociety(ctx, societyID)
	if err != nil {
		return nil, err
	}
	if helpers == nil {
		helpers = []model.DomesticHelp{}
	}
	return helpers, nil
}

// LinkHelpToFlat links a domestic helper to a flat.
func (s *VendorService) LinkHelpToFlat(ctx context.Context, helpID uuid.UUID, req model.LinkFlatRequest) (*model.DomesticHelpFlat, error) {
	h, err := s.repo.FindDomesticHelpByID(ctx, helpID)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, ErrDomesticHelpNotFound
	}

	link := &model.DomesticHelpFlat{
		DomesticHelpID: helpID,
		FlatID:         req.FlatID,
		MonthlyPay:     req.MonthlyPay,
	}
	return s.repo.LinkHelpToFlat(ctx, link)
}

// LogAttendance records entry for domestic help at a flat.
func (s *VendorService) LogAttendance(ctx context.Context, helpID, societyID uuid.UUID, req model.LogAttendanceRequest) (*model.DomesticHelpAttendance, error) {
	h, err := s.repo.FindDomesticHelpByID(ctx, helpID)
	if err != nil {
		return nil, err
	}
	if h == nil {
		return nil, ErrDomesticHelpNotFound
	}

	att := &model.DomesticHelpAttendance{
		DomesticHelpID: helpID,
		SocietyID:      societyID,
		FlatID:         req.FlatID,
	}
	return s.repo.LogAttendance(ctx, att)
}

// LogExit records exit time for an attendance record.
func (s *VendorService) LogExit(ctx context.Context, attendanceID uuid.UUID) error {
	return s.repo.LogExit(ctx, attendanceID)
}

// GetAttendance returns attendance records for a domestic helper in a given month.
func (s *VendorService) GetAttendance(ctx context.Context, helpID uuid.UUID, year int, month time.Month) ([]model.DomesticHelpAttendance, error) {
	records, err := s.repo.GetAttendance(ctx, helpID, year, month)
	if err != nil {
		return nil, err
	}
	if records == nil {
		records = []model.DomesticHelpAttendance{}
	}
	return records, nil
}
