package model

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a maintenance payment record for a flat.
type Payment struct {
	ID               uuid.UUID  `json:"id"`
	SocietyID        uuid.UUID  `json:"society_id"`
	FlatID           uuid.UUID  `json:"flat_id"`
	FlatNumber       string     `json:"flat_number,omitempty"`
	InvoiceNumber    string     `json:"invoice_number"`
	BillMonth        time.Time  `json:"bill_month"`
	BaseAmount       float64    `json:"base_amount"`
	LateFee          float64    `json:"late_fee"`
	Discount         float64    `json:"discount"`
	TotalDue         float64    `json:"total_due"`
	AmountPaid       float64    `json:"amount_paid"`
	Status           string     `json:"status"` // pending, paid, overdue, partially_paid
	PaymentMethod    *string    `json:"payment_method,omitempty"`
	GatewayOrderID   *string    `json:"gateway_order_id,omitempty"`
	GatewayPaymentID *string    `json:"gateway_payment_id,omitempty"`
	PaidAt           *time.Time `json:"paid_at,omitempty"`
	ReceiptURL       *string    `json:"receipt_url,omitempty"`
	DueDate          time.Time  `json:"due_date"`
	ReminderCount    int        `json:"reminder_count"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Expense represents a society expense record.
type Expense struct {
	ID          uuid.UUID  `json:"id"`
	SocietyID   uuid.UUID  `json:"society_id"`
	Category    string     `json:"category"`
	Description string     `json:"description"`
	Amount      float64    `json:"amount"`
	ExpenseDate time.Time  `json:"expense_date"`
	ReceiptURL  *string    `json:"receipt_url,omitempty"`
	VendorID    *uuid.UUID `json:"vendor_id,omitempty"`
	CreatedBy   uuid.UUID  `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// FinancialSummary provides an overview of a society's finances.
type FinancialSummary struct {
	TotalCollected float64 `json:"total_collected"`
	TotalPending   float64 `json:"total_pending"`
	TotalExpenses  float64 `json:"total_expenses"`
	NetBalance     float64 `json:"net_balance"`
}

// --- Request DTOs ---

// CreateBillRequest is the input for generating monthly bills.
type CreateBillRequest struct {
	SocietyID string `json:"society_id"`
	Month     string `json:"month"` // YYYY-MM format
	Amount    float64 `json:"amount"`
	DueDay    int     `json:"due_day"`
}

// RecordPaymentRequest is the input for recording an online payment.
type RecordPaymentRequest struct {
	PaymentID        string `json:"payment_id"`
	Method           string `json:"method"`
	GatewayOrderID   string `json:"gateway_order_id"`
	GatewayPaymentID string `json:"gateway_payment_id"`
}

// RecordCashRequest is the input for recording a cash or cheque payment.
type RecordCashRequest struct {
	PaymentID string `json:"payment_id"`
	Method    string `json:"method"` // cash, cheque, bank_transfer
	Notes     string `json:"notes,omitempty"`
}

// CreateExpenseRequest is the input for creating a new expense record.
type CreateExpenseRequest struct {
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	ExpenseDate string  `json:"expense_date"` // YYYY-MM-DD format
	ReceiptURL  *string `json:"receipt_url,omitempty"`
	VendorID    *string `json:"vendor_id,omitempty"`
}

// PaymentListFilter specifies filters for listing payments.
type PaymentListFilter struct {
	SocietyID uuid.UUID
	FlatID    *uuid.UUID
	Status    *string
	Cursor    *uuid.UUID
	Limit     int
}
