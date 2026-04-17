package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/complaint-service/internal/model"
	"github.com/societykro/complaint-service/internal/repository"
)

var (
	ErrComplaintNotFound = errors.New("complaint not found")
	ErrInvalidStatus     = errors.New("invalid status transition")
	ErrInvalidRating     = errors.New("rating must be between 1 and 5")
)

// ComplaintService handles all complaint business logic.
type ComplaintService struct {
	repo *repository.ComplaintRepository
	bus  *events.Bus
}

// NewComplaintService creates a new ComplaintService.
func NewComplaintService(repo *repository.ComplaintRepository, bus *events.Bus) *ComplaintService {
	return &ComplaintService{repo: repo, bus: bus}
}

// Create raises a new complaint and publishes an event.
func (s *ComplaintService) Create(ctx context.Context, req model.CreateComplaintRequest, userID, societyID uuid.UUID, flatID *uuid.UUID) (*model.Complaint, error) {
	comp := &model.Complaint{
		SocietyID:    societyID,
		FlatID:       flatID,
		RaisedBy:     userID,
		Category:     req.Category,
		Title:        req.Title,
		Description:  req.Description,
		ImageURLs:    req.ImageURLs,
		Priority:     req.Priority,
		IsEmergency:  req.IsEmergency,
		IsCommonArea: req.IsCommonArea,
		Source:       req.Source,
	}

	created, err := s.repo.Create(ctx, comp)
	if err != nil {
		return nil, fmt.Errorf("create complaint: %w", err)
	}

	// Publish event
	subject := events.SubjectComplaintCreated
	if created.IsEmergency {
		subject = "complaint.emergency"
	}
	if err := s.bus.Publish(subject, "complaint.created", created); err != nil {
		logger.Log.Error().Err(err).Str("ticket", created.TicketNumber).Msg("Failed to publish complaint event")
	}

	return created, nil
}

// GetByID returns a complaint with full details.
func (s *ComplaintService) GetByID(ctx context.Context, id uuid.UUID) (*model.Complaint, error) {
	comp, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comp == nil {
		return nil, ErrComplaintNotFound
	}
	return comp, nil
}

// List returns filtered and paginated complaints.
func (s *ComplaintService) List(ctx context.Context, filter model.ComplaintListFilter) ([]model.Complaint, error) {
	complaints, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	if complaints == nil {
		complaints = []model.Complaint{}
	}
	return complaints, nil
}

// UpdateStatus transitions a complaint to a new status.
func (s *ComplaintService) UpdateStatus(ctx context.Context, complaintID uuid.UUID, status string, userID uuid.UUID) error {
	valid := map[string]bool{"in_progress": true, "resolved": true, "closed": true, "reopened": true}
	if !valid[status] {
		return ErrInvalidStatus
	}

	comp, err := s.repo.FindByID(ctx, complaintID)
	if err != nil || comp == nil {
		return ErrComplaintNotFound
	}

	if err := s.repo.UpdateStatus(ctx, complaintID, status, userID); err != nil {
		return err
	}

	// Auto-add status change comment
	s.repo.AddComment(ctx, &model.Comment{
		ComplaintID:    complaintID,
		UserID:         userID,
		Comment:        fmt.Sprintf("Status changed to %s", status),
		IsStatusChange: true,
		OldStatus:      &comp.Status,
		NewStatus:      &status,
	})

	// Publish event
	eventSubject := fmt.Sprintf("complaint.%s", status)
	s.bus.Publish(eventSubject, fmt.Sprintf("complaint.%s", status), map[string]interface{}{
		"complaint_id": complaintID,
		"society_id":   comp.SocietyID,
		"status":       status,
		"ticket":       comp.TicketNumber,
	})

	return nil
}

// AssignVendor assigns a vendor to a complaint and transitions to in_progress.
func (s *ComplaintService) AssignVendor(ctx context.Context, complaintID uuid.UUID, vendorIDStr string, assignedBy uuid.UUID) error {
	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		return fmt.Errorf("invalid vendor ID")
	}

	if err := s.repo.AssignVendor(ctx, complaintID, vendorID, assignedBy); err != nil {
		return err
	}

	s.bus.Publish(events.SubjectComplaintAssigned, "complaint.assigned", map[string]interface{}{
		"complaint_id": complaintID,
		"vendor_id":    vendorID,
		"assigned_by":  assignedBy,
	})

	return nil
}

// Rate records a resolution rating for a complaint.
func (s *ComplaintService) Rate(ctx context.Context, complaintID uuid.UUID, rating int, feedback *string) error {
	if rating < 1 || rating > 5 {
		return ErrInvalidRating
	}
	return s.repo.Rate(ctx, complaintID, rating, feedback)
}

// AddComment adds a comment to a complaint thread.
func (s *ComplaintService) AddComment(ctx context.Context, complaintID, userID uuid.UUID, req model.AddCommentRequest) (*model.Comment, error) {
	comment := &model.Comment{
		ComplaintID: complaintID,
		UserID:      userID,
		Comment:     req.Comment,
		ImageURL:    req.ImageURL,
		IsInternal:  req.IsInternal,
	}
	return s.repo.AddComment(ctx, comment)
}

// ListComments returns all comments for a complaint.
func (s *ComplaintService) ListComments(ctx context.Context, complaintID uuid.UUID, isAdmin bool) ([]model.Comment, error) {
	return s.repo.ListComments(ctx, complaintID, isAdmin)
}

// GetStats returns complaint analytics for a society.
func (s *ComplaintService) GetStats(ctx context.Context, societyID uuid.UUID) (map[string]interface{}, error) {
	counts, err := s.repo.GetStats(ctx, societyID)
	if err != nil {
		return nil, err
	}
	avgTime, _ := s.repo.GetAvgResolutionTime(ctx, societyID)

	return map[string]interface{}{
		"counts":                    counts,
		"avg_resolution_time_hours": fmt.Sprintf("%.1f", avgTime),
	}, nil
}
