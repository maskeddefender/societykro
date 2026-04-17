package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/societykro/go-common/events"
	"github.com/societykro/go-common/logger"

	"github.com/societykro/payment-service/internal/model"
	"github.com/societykro/payment-service/internal/repository"
)

var (
	// ErrPaymentNotFound is returned when a payment record does not exist.
	ErrPaymentNotFound = errors.New("payment not found")
	// ErrAlreadyPaid is returned when attempting to pay an already-paid bill.
	ErrAlreadyPaid = errors.New("payment already completed")
	// ErrInvalidAmount is returned when a bill amount is invalid.
	ErrInvalidAmount = errors.New("invalid amount")
)

// PaymentService handles all payment and expense business logic.
type PaymentService struct {
	repo *repository.PaymentRepository
	bus  *events.Bus
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(repo *repository.PaymentRepository, bus *events.Bus) *PaymentService {
	return &PaymentService{repo: repo, bus: bus}
}

// GenerateMonthlyBills creates payment records for all occupied flats in a society.
func (s *PaymentService) GenerateMonthlyBills(ctx context.Context, societyID uuid.UUID, month time.Time, amount float64, dueDay int) (int, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}
	if dueDay < 1 || dueDay > 28 {
		dueDay = 10
	}

	count, err := s.repo.GenerateBills(ctx, societyID, month, amount, dueDay)
	if err != nil {
		return 0, fmt.Errorf("generate monthly bills: %w", err)
	}

	if err := s.bus.Publish(events.SubjectPaymentGenerated, "payment.generated", map[string]interface{}{
		"society_id": societyID,
		"month":      month.Format("2006-01"),
		"count":      count,
		"amount":     amount,
	}); err != nil {
		logger.Log.Error().Err(err).Msg("Failed to publish payment.generated event")
	}

	return count, nil
}

// GetByID returns a payment with full details.
func (s *PaymentService) GetByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	pay, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if pay == nil {
		return nil, ErrPaymentNotFound
	}
	return pay, nil
}

// ListByFlat returns payment history for a specific flat.
func (s *PaymentService) ListByFlat(ctx context.Context, flatID uuid.UUID, cursor *uuid.UUID, limit int) ([]model.Payment, error) {
	payments, err := s.repo.ListByFlat(ctx, flatID, cursor, limit)
	if err != nil {
		return nil, err
	}
	if payments == nil {
		payments = []model.Payment{}
	}
	return payments, nil
}

// ListBySociety returns filtered and paginated payments for a society.
func (s *PaymentService) ListBySociety(ctx context.Context, filter model.PaymentListFilter) ([]model.Payment, error) {
	payments, err := s.repo.ListBySociety(ctx, filter)
	if err != nil {
		return nil, err
	}
	if payments == nil {
		payments = []model.Payment{}
	}
	return payments, nil
}

// InitiatePayment is a placeholder for Razorpay order creation.
// It returns the payment record with its current status for the client to proceed.
func (s *PaymentService) InitiatePayment(ctx context.Context, paymentID uuid.UUID) (*model.Payment, error) {
	pay, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if pay == nil {
		return nil, ErrPaymentNotFound
	}
	if pay.Status == "paid" {
		return nil, ErrAlreadyPaid
	}

	// TODO: Create Razorpay order via Razorpay SDK
	// orderID := razorpay.CreateOrder(pay.TotalDue)
	// Update payment with gateway_order_id

	return pay, nil
}

// ConfirmPayment is called by the Razorpay webhook to mark a payment as paid.
func (s *PaymentService) ConfirmPayment(ctx context.Context, paymentID uuid.UUID, method string, gatewayOrderID, gatewayPaymentID *string) error {
	pay, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}
	if pay == nil {
		return ErrPaymentNotFound
	}
	if pay.Status == "paid" {
		return ErrAlreadyPaid
	}

	now := time.Now().UTC()
	if err := s.repo.UpdatePayment(ctx, paymentID, "paid", method, gatewayOrderID, gatewayPaymentID, now); err != nil {
		return fmt.Errorf("confirm payment: %w", err)
	}

	if err := s.bus.Publish(events.SubjectPaymentReceived, "payment.received", map[string]interface{}{
		"payment_id": paymentID,
		"society_id": pay.SocietyID,
		"flat_id":    pay.FlatID,
		"amount":     pay.TotalDue,
		"method":     method,
	}); err != nil {
		logger.Log.Error().Err(err).Str("payment_id", paymentID.String()).Msg("Failed to publish payment.received event")
	}

	return nil
}

// RecordCashPayment allows an admin to manually record a cash, cheque, or bank transfer payment.
func (s *PaymentService) RecordCashPayment(ctx context.Context, paymentID uuid.UUID, method, notes string, recordedBy uuid.UUID) error {
	pay, err := s.repo.FindByID(ctx, paymentID)
	if err != nil {
		return err
	}
	if pay == nil {
		return ErrPaymentNotFound
	}
	if pay.Status == "paid" {
		return ErrAlreadyPaid
	}

	if err := s.repo.RecordCash(ctx, paymentID, method, notes, recordedBy); err != nil {
		return fmt.Errorf("record cash payment: %w", err)
	}

	if err := s.bus.Publish(events.SubjectPaymentReceived, "payment.received", map[string]interface{}{
		"payment_id":  paymentID,
		"society_id":  pay.SocietyID,
		"flat_id":     pay.FlatID,
		"amount":      pay.TotalDue,
		"method":      method,
		"recorded_by": recordedBy,
	}); err != nil {
		logger.Log.Error().Err(err).Str("payment_id", paymentID.String()).Msg("Failed to publish payment.received event")
	}

	return nil
}

// GetPendingDues returns the total unpaid amount for a flat.
func (s *PaymentService) GetPendingDues(ctx context.Context, flatID uuid.UUID) (float64, error) {
	return s.repo.GetPendingDues(ctx, flatID)
}

// GetDefaulters returns a list of flats with overdue payments for a society.
func (s *PaymentService) GetDefaulters(ctx context.Context, societyID uuid.UUID) ([]repository.Defaulter, error) {
	defaulters, err := s.repo.GetDefaulters(ctx, societyID)
	if err != nil {
		return nil, err
	}
	if defaulters == nil {
		defaulters = []repository.Defaulter{}
	}
	return defaulters, nil
}

// CreateExpense records a new society expense.
func (s *PaymentService) CreateExpense(ctx context.Context, expense *model.Expense) (*model.Expense, error) {
	if expense.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	return s.repo.CreateExpense(ctx, expense)
}

// ListExpenses returns paginated expenses for a society.
func (s *PaymentService) ListExpenses(ctx context.Context, societyID uuid.UUID, cursor *uuid.UUID, limit int) ([]model.Expense, error) {
	expenses, err := s.repo.ListExpenses(ctx, societyID, cursor, limit)
	if err != nil {
		return nil, err
	}
	if expenses == nil {
		expenses = []model.Expense{}
	}
	return expenses, nil
}

// GetFinancialSummary returns the financial overview for a society.
func (s *PaymentService) GetFinancialSummary(ctx context.Context, societyID uuid.UUID) (*model.FinancialSummary, error) {
	return s.repo.GetFinancialSummary(ctx, societyID)
}
