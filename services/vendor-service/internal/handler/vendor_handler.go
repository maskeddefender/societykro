package handler

import (
	"errors"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/vendor-service/internal/model"
	"github.com/societykro/vendor-service/internal/service"
)

// VendorHandler handles all vendor and domestic help HTTP endpoints.
type VendorHandler struct {
	svc *service.VendorService
}

// NewVendorHandler creates a new VendorHandler.
func NewVendorHandler(svc *service.VendorService) *VendorHandler {
	return &VendorHandler{svc: svc}
}

// --------------- Vendor Endpoints ---------------

// CreateVendor registers a new vendor.
// POST /api/v1/vendors
func (h *VendorHandler) CreateVendor(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.CreateVendorRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Name == "" || req.Phone == "" || req.Category == "" {
		return response.BadRequest(c, "Required: name, phone, category")
	}

	vendor, err := h.svc.CreateVendor(c.Context(), req, societyID)
	if err != nil {
		return response.InternalError(c, "Failed to create vendor")
	}

	return response.Created(c, vendor)
}

// ListVendors returns vendors for the society with optional filters.
// GET /api/v1/vendors?category=plumber&limit=20&cursor=uuid
func (h *VendorHandler) ListVendors(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	filter := model.VendorListFilter{
		SocietyID: societyID,
		Limit:     c.QueryInt("limit", 20),
	}

	if cat := c.Query("category"); cat != "" {
		filter.Category = &cat
	}
	if sub := c.Query("sub_category"); sub != "" {
		filter.SubCategory = &sub
	}
	if cur := c.Query("cursor"); cur != "" {
		if id, err := uuid.Parse(cur); err == nil {
			filter.Cursor = &id
		}
	}

	vendors, err := h.svc.ListVendors(c.Context(), filter)
	if err != nil {
		return response.InternalError(c, "Failed to list vendors")
	}

	var nextCursor string
	hasMore := len(vendors) == filter.Limit
	if hasMore && len(vendors) > 0 {
		nextCursor = vendors[len(vendors)-1].ID.String()
	}

	return response.Paginated(c, vendors, nextCursor, hasMore, 0)
}

// GetVendor returns a vendor by ID.
// GET /api/v1/vendors/:id
func (h *VendorHandler) GetVendor(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid vendor ID")
	}

	vendor, err := h.svc.GetVendorByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrVendorNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch vendor")
	}

	return response.OK(c, vendor)
}

// UpdateVendor updates vendor details.
// PUT /api/v1/vendors/:id
func (h *VendorHandler) UpdateVendor(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid vendor ID")
	}

	var req model.UpdateVendorRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	if err := h.svc.UpdateVendor(c.Context(), id, req); err != nil {
		if errors.Is(err, service.ErrVendorNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to update vendor")
	}

	return response.OKMessage(c, "Vendor updated")
}

// DeleteVendor soft-deletes a vendor.
// DELETE /api/v1/vendors/:id
func (h *VendorHandler) DeleteVendor(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid vendor ID")
	}

	if err := h.svc.DeleteVendor(c.Context(), id); err != nil {
		if errors.Is(err, service.ErrVendorNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to delete vendor")
	}

	return response.OKMessage(c, "Vendor deleted")
}

// --------------- Domestic Help Endpoints ---------------

// CreateDomesticHelp registers a new domestic helper.
// POST /api/v1/domestic-help
func (h *VendorHandler) CreateDomesticHelp(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.CreateDomesticHelpRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Name == "" || req.Phone == "" || req.Role == "" {
		return response.BadRequest(c, "Required: name, phone, role")
	}

	help, err := h.svc.CreateDomesticHelp(c.Context(), req, societyID)
	if err != nil {
		return response.InternalError(c, "Failed to create domestic help")
	}

	return response.Created(c, help)
}

// ListDomesticHelp returns domestic help by society or flat.
// GET /api/v1/domestic-help?flat_id=uuid
func (h *VendorHandler) ListDomesticHelp(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var flatID *uuid.UUID
	if fid := c.Query("flat_id"); fid != "" {
		if id, err := uuid.Parse(fid); err == nil {
			flatID = &id
		}
	}

	helpers, err := h.svc.ListDomesticHelp(c.Context(), societyID, flatID)
	if err != nil {
		return response.InternalError(c, "Failed to list domestic help")
	}

	return response.OK(c, helpers)
}

// GetDomesticHelp returns a domestic help record by ID.
// GET /api/v1/domestic-help/:id
func (h *VendorHandler) GetDomesticHelp(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid domestic help ID")
	}

	help, err := h.svc.GetDomesticHelpByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrDomesticHelpNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch domestic help")
	}

	return response.OK(c, help)
}

// LinkToFlat links a domestic helper to a flat with pay details.
// POST /api/v1/domestic-help/:id/link-flat
func (h *VendorHandler) LinkToFlat(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid domestic help ID")
	}

	var req model.LinkFlatRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.FlatID == uuid.Nil {
		return response.BadRequest(c, "flat_id is required")
	}

	link, err := h.svc.LinkHelpToFlat(c.Context(), id, req)
	if err != nil {
		if errors.Is(err, service.ErrDomesticHelpNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to link domestic help to flat")
	}

	return response.Created(c, link)
}

// LogAttendance records entry for domestic help.
// POST /api/v1/domestic-help/:id/attendance
func (h *VendorHandler) LogAttendance(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid domestic help ID")
	}

	var req model.LogAttendanceRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.FlatID == uuid.Nil {
		return response.BadRequest(c, "flat_id is required")
	}

	att, err := h.svc.LogAttendance(c.Context(), id, societyID, req)
	if err != nil {
		if errors.Is(err, service.ErrDomesticHelpNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to log attendance")
	}

	return response.Created(c, att)
}

// LogExit records exit time for an attendance record.
// PUT /api/v1/domestic-help/attendance/:id/exit
func (h *VendorHandler) LogExit(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid attendance ID")
	}

	if err := h.svc.LogExit(c.Context(), id); err != nil {
		return response.InternalError(c, "Failed to log exit")
	}

	return response.OKMessage(c, "Exit recorded")
}

// GetAttendance returns attendance records for a domestic helper in a given month.
// GET /api/v1/domestic-help/:id/attendance?month=2026-04
func (h *VendorHandler) GetAttendance(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid domestic help ID")
	}

	// Default to current month
	now := time.Now()
	year := now.Year()
	month := now.Month()

	if m := c.Query("month"); m != "" {
		t, err := time.Parse("2006-01", m)
		if err != nil {
			return response.BadRequest(c, "Invalid month format, use YYYY-MM")
		}
		year = t.Year()
		month = t.Month()
	} else {
		// Also support year/month as separate params
		if y := c.Query("year"); y != "" {
			if v, err := strconv.Atoi(y); err == nil {
				year = v
			}
		}
		if m := c.Query("m"); m != "" {
			if v, err := strconv.Atoi(m); err == nil && v >= 1 && v <= 12 {
				month = time.Month(v)
			}
		}
	}

	records, err := h.svc.GetAttendance(c.Context(), id, year, month)
	if err != nil {
		return response.InternalError(c, "Failed to fetch attendance")
	}

	return response.OK(c, records)
}
