package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/notice-service/internal/model"
	"github.com/societykro/notice-service/internal/service"
)

// NoticeHandler handles notice HTTP endpoints.
type NoticeHandler struct {
	svc *service.NoticeService
}

// NewNoticeHandler creates a new NoticeHandler.
func NewNoticeHandler(svc *service.NoticeService) *NoticeHandler {
	return &NoticeHandler{svc: svc}
}

// Create posts a new notice.
// POST /api/v1/notices
func (h *NoticeHandler) Create(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.CreateNoticeRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Title == "" || req.Body == "" {
		return response.BadRequest(c, "Required: title, body")
	}

	notice, err := h.svc.Create(c.Context(), req, userID, societyID)
	if err != nil {
		return response.InternalError(c, "Failed to create notice")
	}

	return response.Created(c, notice)
}

// GetByID returns notice detail with read count.
// GET /api/v1/notices/:id
func (h *NoticeHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid notice ID")
	}

	// Auto-mark as read for the requesting user
	userID, _ := middleware.GetUserID(c)
	_ = h.svc.MarkRead(c.Context(), id, userID)

	notice, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNoticeNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch notice")
	}

	return response.OK(c, notice)
}

// List returns paginated notices (pinned first).
// GET /api/v1/notices?limit=20&cursor=uuid
func (h *NoticeHandler) List(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	limit := c.QueryInt("limit", 20)
	var cursor *uuid.UUID
	if cur := c.Query("cursor"); cur != "" {
		if id, err := uuid.Parse(cur); err == nil {
			cursor = &id
		}
	}

	notices, err := h.svc.List(c.Context(), societyID, cursor, limit)
	if err != nil {
		return response.InternalError(c, "Failed to list notices")
	}

	var nextCursor string
	hasMore := len(notices) == limit
	if hasMore && len(notices) > 0 {
		nextCursor = notices[len(notices)-1].ID.String()
	}

	return response.Paginated(c, notices, nextCursor, hasMore, 0)
}

// MarkRead marks a notice as read by the current user.
// POST /api/v1/notices/:id/read
func (h *NoticeHandler) MarkRead(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid notice ID")
	}

	if err := h.svc.MarkRead(c.Context(), id, userID); err != nil {
		return response.InternalError(c, "Failed to mark as read")
	}

	return response.OKMessage(c, "Marked as read")
}

// Delete removes a notice (admin only).
// DELETE /api/v1/notices/:id
func (h *NoticeHandler) Delete(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid notice ID")
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete notice")
	}

	return response.OKMessage(c, "Notice deleted")
}
