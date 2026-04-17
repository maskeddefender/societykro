package handler

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/complaint-service/internal/model"
	"github.com/societykro/complaint-service/internal/service"
)

// ComplaintHandler handles all complaint HTTP endpoints.
type ComplaintHandler struct {
	svc *service.ComplaintService
}

// NewComplaintHandler creates a new ComplaintHandler.
func NewComplaintHandler(svc *service.ComplaintService) *ComplaintHandler {
	return &ComplaintHandler{svc: svc}
}

// Create raises a new complaint.
// POST /api/v1/complaints
func (h *ComplaintHandler) Create(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.CreateComplaintRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Category == "" || req.Title == "" || req.Description == "" {
		return response.BadRequest(c, "Required: category, title, description")
	}
	if req.ImageURLs == nil {
		req.ImageURLs = []string{}
	}

	comp, err := h.svc.Create(c.Context(), req, userID, societyID, nil)
	if err != nil {
		return response.InternalError(c, "Failed to create complaint")
	}

	return response.Created(c, comp)
}

// GetByID returns complaint details.
// GET /api/v1/complaints/:id
func (h *ComplaintHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid complaint ID")
	}

	comp, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrComplaintNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch complaint")
	}

	return response.OK(c, comp)
}

// List returns filtered complaints for the user's society.
// GET /api/v1/complaints?status=open&category=water&limit=20&cursor=uuid
func (h *ComplaintHandler) List(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	filter := model.ComplaintListFilter{
		SocietyID: societyID,
		Limit:     c.QueryInt("limit", 20),
	}

	if s := c.Query("status"); s != "" {
		filter.Status = &s
	}
	if cat := c.Query("category"); cat != "" {
		filter.Category = &cat
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

	complaints, err := h.svc.List(c.Context(), filter)
	if err != nil {
		return response.InternalError(c, "Failed to list complaints")
	}

	var nextCursor string
	hasMore := len(complaints) == filter.Limit
	if hasMore && len(complaints) > 0 {
		nextCursor = complaints[len(complaints)-1].ID.String()
	}

	return response.Paginated(c, complaints, nextCursor, hasMore, 0)
}

// UpdateStatus changes complaint status.
// PUT /api/v1/complaints/:id/status
func (h *ComplaintHandler) UpdateStatus(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid complaint ID")
	}

	var req model.UpdateStatusRequest
	if err := c.BodyParser(&req); err != nil || req.Status == "" {
		return response.BadRequest(c, "status is required")
	}

	if err := h.svc.UpdateStatus(c.Context(), id, req.Status, userID); err != nil {
		if errors.Is(err, service.ErrInvalidStatus) {
			return response.BadRequest(c, err.Error())
		}
		if errors.Is(err, service.ErrComplaintNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to update status")
	}

	return response.OKMessage(c, "Status updated to "+req.Status)
}

// AssignVendor assigns a vendor to a complaint.
// PUT /api/v1/complaints/:id/assign
func (h *ComplaintHandler) AssignVendor(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid complaint ID")
	}

	var req model.AssignVendorRequest
	if err := c.BodyParser(&req); err != nil || req.VendorID == "" {
		return response.BadRequest(c, "vendor_id is required")
	}

	if err := h.svc.AssignVendor(c.Context(), id, req.VendorID, userID); err != nil {
		return response.InternalError(c, "Failed to assign vendor")
	}

	return response.OKMessage(c, "Vendor assigned")
}

// Rate records a resolution rating.
// POST /api/v1/complaints/:id/rate
func (h *ComplaintHandler) Rate(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid complaint ID")
	}

	var req model.RateResolutionRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.svc.Rate(c.Context(), id, req.Rating, req.Feedback); err != nil {
		if errors.Is(err, service.ErrInvalidRating) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalError(c, "Failed to rate")
	}

	return response.OKMessage(c, "Rating recorded")
}

// AddComment adds a comment to a complaint.
// POST /api/v1/complaints/:id/comments
func (h *ComplaintHandler) AddComment(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid complaint ID")
	}

	var req model.AddCommentRequest
	if err := c.BodyParser(&req); err != nil || req.Comment == "" {
		return response.BadRequest(c, "comment is required")
	}

	comment, err := h.svc.AddComment(c.Context(), id, userID, req)
	if err != nil {
		return response.InternalError(c, "Failed to add comment")
	}

	return response.Created(c, comment)
}

// ListComments returns all comments for a complaint.
// GET /api/v1/complaints/:id/comments
func (h *ComplaintHandler) ListComments(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid complaint ID")
	}

	role, _ := middleware.GetRole(c)
	isAdmin := role == "admin" || role == "secretary" || role == "president"

	comments, err := h.svc.ListComments(c.Context(), id, isAdmin)
	if err != nil {
		return response.InternalError(c, "Failed to list comments")
	}
	if comments == nil {
		comments = []model.Comment{}
	}

	return response.OK(c, comments)
}

// GetStats returns complaint analytics for the society.
// GET /api/v1/complaints/analytics
func (h *ComplaintHandler) GetStats(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	stats, err := h.svc.GetStats(c.Context(), societyID)
	if err != nil {
		return response.InternalError(c, "Failed to fetch analytics")
	}

	return response.OK(c, stats)
}
