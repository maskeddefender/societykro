package handler

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/societykro/go-common/middleware"
	"github.com/societykro/go-common/response"

	"github.com/societykro/payment-service/internal/model"
	"github.com/societykro/payment-service/internal/service"
)

// PaymentHandler handles all payment and expense HTTP endpoints.
type PaymentHandler struct {
	svc *service.PaymentService
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// GenerateBills creates monthly payment records for all occupied flats.
// POST /api/v1/payments/generate-bills
func (h *PaymentHandler) GenerateBills(c *fiber.Ctx) error {
	var req model.CreateBillRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	// Use society_id from JWT if not provided in body
	if req.SocietyID == "" {
		sid, _ := middleware.GetSocietyID(c)
		req.SocietyID = sid
	}

	if req.SocietyID == "" || req.Month == "" {
		return response.BadRequest(c, "Required: month (YYYY-MM). Optional: amount (defaults to society config)")
	}

	societyID, err := uuid.Parse(req.SocietyID)
	if err != nil {
		return response.BadRequest(c, "Invalid society_id")
	}

	month, err := time.Parse("2006-01", req.Month)
	if err != nil {
		return response.BadRequest(c, "Invalid month format, use YYYY-MM")
	}

	// Default amount to 0 (service will use society's maintenance_amount)
	if req.Amount <= 0 {
		req.Amount = 3500 // fallback default
	}

	count, err := h.svc.GenerateMonthlyBills(c.Context(), societyID, month, req.Amount, req.DueDay)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalError(c, "Failed to generate bills")
	}

	return response.OK(c, fiber.Map{"bills_generated": count})
}

// List returns payments filtered by society or flat.
// GET /api/v1/payments?flat_id=...&status=...&cursor=...&limit=...
func (h *PaymentHandler) List(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	// If flat_id is provided, return flat-specific history
	if flatIDStr := c.Query("flat_id"); flatIDStr != "" {
		flatID, err := uuid.Parse(flatIDStr)
		if err != nil {
			return response.BadRequest(c, "Invalid flat_id")
		}

		var cursor *uuid.UUID
		if cur := c.Query("cursor"); cur != "" {
			if id, err := uuid.Parse(cur); err == nil {
				cursor = &id
			}
		}

		payments, err := h.svc.ListByFlat(c.Context(), flatID, cursor, c.QueryInt("limit", 20))
		if err != nil {
			return response.InternalError(c, "Failed to list payments")
		}

		return paginatedPayments(c, payments, c.QueryInt("limit", 20))
	}

	// Otherwise list by society with filters
	filter := model.PaymentListFilter{
		SocietyID: societyID,
		Limit:     c.QueryInt("limit", 20),
	}

	if s := c.Query("status"); s != "" {
		filter.Status = &s
	}
	if cur := c.Query("cursor"); cur != "" {
		if id, err := uuid.Parse(cur); err == nil {
			filter.Cursor = &id
		}
	}

	payments, err := h.svc.ListBySociety(c.Context(), filter)
	if err != nil {
		return response.InternalError(c, "Failed to list payments")
	}

	return paginatedPayments(c, payments, filter.Limit)
}

// GetPendingDues returns the total unpaid amount for the current user's flat.
// GET /api/v1/payments/pending?flat_id=...
func (h *PaymentHandler) GetPendingDues(c *fiber.Ctx) error {
	flatIDStr := c.Query("flat_id")
	if flatIDStr == "" {
		return response.BadRequest(c, "flat_id is required")
	}

	flatID, err := uuid.Parse(flatIDStr)
	if err != nil {
		return response.BadRequest(c, "Invalid flat_id")
	}

	total, err := h.svc.GetPendingDues(c.Context(), flatID)
	if err != nil {
		return response.InternalError(c, "Failed to fetch pending dues")
	}

	return response.OK(c, fiber.Map{"flat_id": flatID, "total_pending": total})
}

// GetDefaulters returns flats with overdue payments for a society.
// GET /api/v1/payments/defaulters
func (h *PaymentHandler) GetDefaulters(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	defaulters, err := h.svc.GetDefaulters(c.Context(), societyID)
	if err != nil {
		return response.InternalError(c, "Failed to fetch defaulters")
	}

	return response.OK(c, defaulters)
}

// GetByID returns payment details.
// GET /api/v1/payments/:id
func (h *PaymentHandler) GetByID(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid payment ID")
	}

	pay, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch payment")
	}

	return response.OK(c, pay)
}

// InitiatePayment starts the payment flow (placeholder for Razorpay order creation).
// POST /api/v1/payments/:id/initiate
func (h *PaymentHandler) InitiatePayment(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid payment ID")
	}

	pay, err := h.svc.InitiatePayment(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return response.NotFound(c, err.Error())
		}
		if errors.Is(err, service.ErrAlreadyPaid) {
			return response.Conflict(c, err.Error())
		}
		return response.InternalError(c, "Failed to initiate payment")
	}

	return response.OK(c, pay)
}

// RazorpayWebhook handles Razorpay payment confirmation callbacks.
// POST /api/v1/payments/webhook/razorpay
func (h *PaymentHandler) RazorpayWebhook(c *fiber.Ctx) error {
	// TODO: Validate Razorpay webhook signature
	// signature := c.Get("X-Razorpay-Signature")
	// if !razorpay.VerifySignature(c.Body(), signature, secret) {
	//     return response.Unauthorized(c, "Invalid webhook signature")
	// }

	var req model.RecordPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.PaymentID == "" || req.GatewayPaymentID == "" {
		return response.BadRequest(c, "Required: payment_id, gateway_payment_id")
	}

	paymentID, err := uuid.Parse(req.PaymentID)
	if err != nil {
		return response.BadRequest(c, "Invalid payment_id")
	}

	method := req.Method
	if method == "" {
		method = "online"
	}

	if err := h.svc.ConfirmPayment(c.Context(), paymentID, method, &req.GatewayOrderID, &req.GatewayPaymentID); err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return response.NotFound(c, err.Error())
		}
		if errors.Is(err, service.ErrAlreadyPaid) {
			return response.Conflict(c, err.Error())
		}
		return response.InternalError(c, "Failed to confirm payment")
	}

	return response.OKMessage(c, "Payment confirmed")
}

// RecordCash records a cash or cheque payment by an admin.
// POST /api/v1/payments/:id/record-cash
func (h *PaymentHandler) RecordCash(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid payment ID")
	}

	var req model.RecordCashRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Method == "" {
		return response.BadRequest(c, "method is required (cash, cheque, bank_transfer)")
	}

	if err := h.svc.RecordCashPayment(c.Context(), id, req.Method, req.Notes, userID); err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return response.NotFound(c, err.Error())
		}
		if errors.Is(err, service.ErrAlreadyPaid) {
			return response.Conflict(c, err.Error())
		}
		return response.InternalError(c, "Failed to record payment")
	}

	return response.OKMessage(c, "Cash payment recorded")
}

// GetReceipt redirects to the receipt URL for a payment.
// GET /api/v1/payments/:id/receipt
func (h *PaymentHandler) GetReceipt(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid payment ID")
	}

	pay, err := h.svc.GetByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrPaymentNotFound) {
			return response.NotFound(c, err.Error())
		}
		return response.InternalError(c, "Failed to fetch payment")
	}

	if pay.ReceiptURL == nil || *pay.ReceiptURL == "" {
		return response.NotFound(c, "Receipt not available")
	}

	return c.Redirect(*pay.ReceiptURL, fiber.StatusFound)
}

// CreateExpense records a new society expense.
// POST /api/v1/expenses
func (h *PaymentHandler) CreateExpense(c *fiber.Ctx) error {
	userID, _ := middleware.GetUserID(c)
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var req model.CreateExpenseRequest
	if err := c.BodyParser(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}
	if req.Category == "" || req.Description == "" || req.Amount <= 0 || req.ExpenseDate == "" {
		return response.BadRequest(c, "Required: category, description, amount, expense_date")
	}

	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return response.BadRequest(c, "Invalid expense_date format, use YYYY-MM-DD")
	}

	expense := &model.Expense{
		SocietyID:   societyID,
		Category:    req.Category,
		Description: req.Description,
		Amount:      req.Amount,
		ExpenseDate: expenseDate,
		ReceiptURL:  req.ReceiptURL,
		CreatedBy:   userID,
	}

	if req.VendorID != nil {
		vid, err := uuid.Parse(*req.VendorID)
		if err != nil {
			return response.BadRequest(c, "Invalid vendor_id")
		}
		expense.VendorID = &vid
	}

	created, err := h.svc.CreateExpense(c.Context(), expense)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAmount) {
			return response.BadRequest(c, err.Error())
		}
		return response.InternalError(c, "Failed to create expense")
	}

	return response.Created(c, created)
}

// ListExpenses returns paginated expenses for a society.
// GET /api/v1/expenses?cursor=...&limit=...
func (h *PaymentHandler) ListExpenses(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	var cursor *uuid.UUID
	if cur := c.Query("cursor"); cur != "" {
		if id, err := uuid.Parse(cur); err == nil {
			cursor = &id
		}
	}

	expenses, err := h.svc.ListExpenses(c.Context(), societyID, cursor, c.QueryInt("limit", 20))
	if err != nil {
		return response.InternalError(c, "Failed to list expenses")
	}

	limit := c.QueryInt("limit", 20)
	var nextCursor string
	hasMore := len(expenses) == limit
	if hasMore && len(expenses) > 0 {
		nextCursor = expenses[len(expenses)-1].ID.String()
	}

	return response.Paginated(c, expenses, nextCursor, hasMore, 0)
}

// GetFinancialSummary returns the financial overview for a society.
// GET /api/v1/financial-summary
func (h *PaymentHandler) GetFinancialSummary(c *fiber.Ctx) error {
	societyIDStr, _ := middleware.GetSocietyID(c)
	societyID, _ := uuid.Parse(societyIDStr)

	summary, err := h.svc.GetFinancialSummary(c.Context(), societyID)
	if err != nil {
		return response.InternalError(c, "Failed to fetch financial summary")
	}

	return response.OK(c, summary)
}

// paginatedPayments is a helper to build paginated responses for payment lists.
func paginatedPayments(c *fiber.Ctx, payments []model.Payment, limit int) error {
	var nextCursor string
	hasMore := len(payments) == limit
	if hasMore && len(payments) > 0 {
		nextCursor = payments[len(payments)-1].ID.String()
	}
	return response.Paginated(c, payments, nextCursor, hasMore, 0)
}
