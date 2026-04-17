package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/auth-service/internal/model"
	"github.com/societykro/auth-service/internal/service"
)

// SocietyHandler handles society and flat management HTTP endpoints.
type SocietyHandler struct {
	svc *service.AuthService
}

// NewSocietyHandler creates a new SocietyHandler.
func NewSocietyHandler(svc *service.AuthService) *SocietyHandler {
	return &SocietyHandler{svc: svc}
}

// CreateSociety creates a new society and makes the caller an admin.
// POST /api/v1/societies
func (h *SocietyHandler) CreateSociety(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Not authenticated")
	}

	var req model.CreateSocietyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Name == "" || req.Address == "" || req.City == "" || req.State == "" || req.Pincode == "" || req.TotalFlats <= 0 {
		return response.BadRequest(c, "Required: name, address, city, state, pincode, total_flats (>0)")
	}

	society, err := h.svc.CreateSociety(c.Context(), req, userID)
	if err != nil {
		return response.InternalError(c, "Failed to create society")
	}

	return response.Created(c, society)
}

// GetSociety returns society details by UUID.
// GET /api/v1/societies/:id
func (h *SocietyHandler) GetSociety(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid society ID")
	}

	society, err := h.svc.GetSociety(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrSocietyNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch society")
	}

	return response.OK(c, society)
}

// JoinSociety adds the authenticated user to a society by invite code.
// POST /api/v1/societies/join
func (h *SocietyHandler) JoinSociety(c *fiber.Ctx) error {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return response.Unauthorized(c, "Not authenticated")
	}

	var req model.JoinSocietyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Code == "" {
		return response.BadRequest(c, "Society code is required")
	}

	membership, err := h.svc.JoinSociety(c.Context(), userID, req.Code, req.FlatNumber)
	if err != nil {
		if errors.Is(err, service.ErrSocietyNotFound) {
			return response.NotFound(c, "Society not found with this code")
		}
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, membership)
}

// ListFlats returns all flats for a society.
// GET /api/v1/societies/:id/flats
func (h *SocietyHandler) ListFlats(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid society ID")
	}

	flats, err := h.svc.ListFlats(c.Context(), id)
	if err != nil {
		return response.InternalError(c, "Failed to fetch flats")
	}
	if flats == nil {
		flats = []model.Flat{}
	}

	return response.OK(c, flats)
}
