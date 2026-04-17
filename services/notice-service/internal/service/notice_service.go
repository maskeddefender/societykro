package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/notice-service/internal/model"
	"github.com/societykro/notice-service/internal/repository"
)

var ErrNoticeNotFound = errors.New("notice not found")

// NoticeService handles notice business logic.
type NoticeService struct {
	repo *repository.NoticeRepository
	bus  *events.Bus
}

// NewNoticeService creates a new NoticeService.
func NewNoticeService(repo *repository.NoticeRepository, bus *events.Bus) *NoticeService {
	return &NoticeService{repo: repo, bus: bus}
}

// Create posts a new notice and publishes an event for broadcasting.
func (s *NoticeService) Create(ctx context.Context, req model.CreateNoticeRequest, userID, societyID uuid.UUID) (*model.Notice, error) {
	n := &model.Notice{
		SocietyID:         societyID,
		CreatedBy:         userID,
		Title:             req.Title,
		Body:              req.Body,
		Category:          req.Category,
		IsPinned:          req.IsPinned,
		BroadcastWhatsapp: req.BroadcastWhatsapp,
		BroadcastTelegram: req.BroadcastTelegram,
		AttachmentURLs:    req.AttachmentURLs,
	}
	if n.AttachmentURLs == nil {
		n.AttachmentURLs = []string{}
	}

	created, err := s.repo.Create(ctx, n)
	if err != nil {
		return nil, fmt.Errorf("create notice: %w", err)
	}

	if err := s.bus.Publish(events.SubjectNoticePosted, "notice.posted", created); err != nil {
		logger.Log.Error().Err(err).Str("notice_id", created.ID.String()).Msg("Failed to publish notice event")
	}

	return created, nil
}

// GetByID returns a notice with read statistics.
func (s *NoticeService) GetByID(ctx context.Context, id uuid.UUID) (*model.Notice, error) {
	n, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, ErrNoticeNotFound
	}
	return n, nil
}

// List returns paginated notices for a society.
func (s *NoticeService) List(ctx context.Context, societyID uuid.UUID, cursor *uuid.UUID, limit int) ([]model.Notice, error) {
	notices, err := s.repo.List(ctx, societyID, cursor, limit)
	if err != nil {
		return nil, err
	}
	if notices == nil {
		notices = []model.Notice{}
	}
	return notices, nil
}

// MarkRead records that a user has read a notice.
func (s *NoticeService) MarkRead(ctx context.Context, noticeID, userID uuid.UUID) error {
	return s.repo.MarkRead(ctx, noticeID, userID, "app")
}

// Delete removes a notice. Only admins should call this.
func (s *NoticeService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
