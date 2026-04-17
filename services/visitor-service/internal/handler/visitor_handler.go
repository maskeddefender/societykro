package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/societykro/go-common/logger"
	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/visitor-service/internal/model"
	"github.com/societykro/visitor-service/internal/service"
)

// VisitorHandler handles all visitor HTTP endpoints.
type VisitorHandler struct {
	svc *service.VisitorService
}

// NewVisitorHandler creates a new VisitorHandler.
func NewVisitorHandler(svc *service.VisitorService) *VisitorHandler {
	return &VisitorHandler{svc: svc}
}

// LogVisitor creates a new visitor entry at the gate.
// POST /api/v1/visitors/log
func (h *VisitorHandler) LogVisitor(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.CreateVisitorRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Name == "" || req.FlatID == "" || req.Purpose == "" {
		return response.BadRequest(c, "Required: name, flat_id, purpose")
	}

	vis, err := h.svc.LogVisitor(c.Context(), req, userID, societyID)
	if err != nil {
		return response.InternalError(c, "Failed to log visitor")
	}

	return response.Created(c, vis)
}

// PreApprove generates an OTP for a pre-approved visitor.
// POST /api/v1/visitors/pre-approve
func (h *VisitorHandler) PreApprove(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.PreApproveRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Name == "" || req.Purpose == "" {
		return response.BadRequest(c, "Required: name, purpose")
	}

	// Flat ID: from request or look up from membership
	var flatID uuid.UUID
	if req.FlatID != "" {
		flatID, _ = uuid.Parse(req.FlatID)
	}
	vis, otp, err := h.svc.PreApprove(c.Context(), req, userID, societyID, flatID)
	if err != nil {
		logger.Log.Error().Err(err).Msg("PreApprove failed")
		return response.InternalError(c, err.Error())
	}

	return response.Created(c, fiber.Map{
		"visitor": vis,
		"otp":     otp,
	})
}

// VerifyOTP checks an OTP at the gate.
// POST /api/v1/visitors/verify-otp
func (h *VisitorHandler) VerifyOTP(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.VerifyOTPRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.OTPCode == "" {
		return response.BadRequest(c, "otp_code is required")
	}

	vis, err := h.svc.VerifyOTP(c.Context(), societyID, req.OTPCode)
	if err != nil {
		if errors.Is(err, service.ErrVisitorNotFound) {
			return response.NotFound(c, "Invalid OTP")
		}
		if errors.Is(err, service.ErrOTPExpired) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalError(c, "Failed to verify OTP")
	}

	return response.OK(c, vis)
}

// Approve approves a pending visitor.
// PUT /api/v1/visitors/:id/approve
func (h *VisitorHandler) Approve(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid visitor ID")
	}

	var req model.ApproveRequest
	_ = c.BodyParser(&req)

	if err := h.svc.Approve(c.Context(), id, userID, req.Via); err != nil {
		if errors.Is(err, service.ErrVisitorNotFound) {
			return response.NotFound(c, err.Error())
		}
		if errors.Is(err, service.ErrAlreadyProcessed) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalError(c, "Failed to approve visitor")
	}

	return response.OKMessage(c, "Visitor approved")
}

// Deny denies a pending visitor.
// PUT /api/v1/visitors/:id/deny
func (h *VisitorHandler) Deny(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid visitor ID")
	}

	var req model.DenyRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.svc.Deny(c.Context(), id, userID, req.Reason); err != nil {
		if errors.Is(err, service.ErrVisitorNotFound) {
			return response.NotFound(c, err.Error())
		}
		if errors.Is(err, service.ErrAlreadyProcessed) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalError(c, "Failed to deny visitor")
	}

	return response.OKMessage(c, "Visitor denied")
}

// Checkout marks a visitor as checked out.
// PUT /api/v1/visitors/:id/checkout
func (h *VisitorHandler) Checkout(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid visitor ID")
	}

	if err := h.svc.Checkout(c.Context(), id); err != nil {
		if errors.Is(err, service.ErrVisitorNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to checkout visitor")
	}

	return response.OKMessage(c, "Visitor checked out")
}

// List returns filtered visitors for the user's society.
// GET /api/v1/visitors?status=pending&flat_id=uuid&limit=20&cursor=uuid
func (h *VisitorHandler) List(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	filter := model.VisitorListFilter{
		SocietyID: societyID,
		Limit:     c.QueryInt("limit", 20),
	}

	if s := c.Query("status"); s != "" {
		filter.Status = &s
	}
	if fid := c.Query("flat_id"); fid != "" {
		if id, err := uuid.Parse(fid); err == nil {
			filter.FlatID = &id
		}
	}
	if cur := c.Query("cursor"); cur != "" {
		if id, err := uuid.Parse(cur); err == nil {
			filter.Cursor = &id
		}
	}

	visitors, err := h.svc.List(c.Context(), filter)
	if err != nil {
		return response.InternalError(c, "Failed to list visitors")
	}

	var nextCursor string
	hasMore := len(visitors) == filter.Limit
	if hasMore && len(visitors) > 0 {
		nextCursor = visitors[len(visitors)-1].ID.String()
	}

	return response.Paginated(c, visitors, nextCursor, hasMore, 0)
}

// ListActive returns currently checked-in visitors.
// GET /api/v1/visitors/active
func (h *VisitorHandler) ListActive(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	visitors, err := h.svc.ListActive(c.Context(), societyID)
	if err != nil {
		return response.InternalError(c, "Failed to list active visitors")
	}

	return response.OK(c, visitors)
}

// GetByID returns visitor details.
// GET /api/v1/visitors/:id
func (h *VisitorHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid visitor ID")
	}

	vis, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrVisitorNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch visitor")
	}

	return response.OK(c, vis)
}

// CreatePass creates a recurring visitor pass.
// POST /api/v1/visitors/passes
func (h *VisitorHandler) CreatePass(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.CreatePassRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.VisitorName == "" || req.Purpose == "" || req.ValidFrom == "" || req.ValidUntil == "" {
		return response.BadRequest(c, "Required: visitor_name, purpose, valid_from, valid_until")
	}

	pass, err := h.svc.CreatePass(c.Context(), req, userID, societyID, userID)
	if err != nil {
		return response.InternalError(c, "Failed to create pass")
	}

	return response.Created(c, pass)
}

// ListPasses returns active passes for the user's flat.
// GET /api/v1/visitors/passes
func (h *VisitorHandler) ListPasses(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)

	// Use userID as flatID lookup key; in production, resolve flat from user context
	passes, err := h.svc.ListPasses(c.Context(), userID)
	if err != nil {
		return response.InternalError(c, "Failed to list passes")
	}

	return response.OK(c, passes)
}

// DeletePass removes a visitor pass.
// DELETE /api/v1/visitors/passes/:id
func (h *VisitorHandler) DeletePass(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid pass ID")
	}

	if err := h.svc.DeletePass(c.Context(), id); err != nil {
		return response.InternalError(c, "Failed to delete pass")
	}

	return response.OKMessage(c, "Pass deleted")
}
