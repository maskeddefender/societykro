package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/payment-service/internal/model"
)

// PaymentRepository handles all database operations for payments and expenses.
type PaymentRepository struct {
	pool *pgxpool.Pool
}

// NewPaymentRepository creates a new PaymentRepository.
func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

// GenerateBills bulk-inserts payment records for all occupied flats in a society.
// It skips flats that already have a bill for the given month.
func (r *PaymentRepository) GenerateBills(ctx context.Context, societyID uuid.UUID, month time.Time, amount float64, dueDay int) (int, error) {
	billMonth := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	dueDate := time.Date(month.Year(), month.Month(), dueDay, 0, 0, 0, 0, time.UTC)

	query := `
		INSERT INTO payment (society_id, flat_id, invoice_number, bill_month, base_amount, total_due, due_date, status)
		SELECT $1, f.id,
			'INV-' || TO_CHAR($2::date, 'YYYYMM') || '-' || ROW_NUMBER() OVER (ORDER BY f.flat_number),
			$2, $3, $3, $4, 'pending'
		FROM flat f
		WHERE f.society_id = $1
			AND f.is_occupied = true
			AND NOT EXISTS (
				SELECT 1 FROM payment p
				WHERE p.flat_id = f.id AND p.bill_month = $2
			)
		ON CONFLICT DO NOTHING`

	tag, err := r.pool.Exec(ctx, query, societyID, billMonth, amount, dueDate)
	if err != nil {
		return 0, fmt.Errorf("generate bills: %w", err)
	}
	return int(tag.RowsAffected()), nil
}

// FindByID returns a payment with the flat number joined.
func (r *PaymentRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Payment, error) {
	query := `SELECT p.id, p.society_id, p.flat_id, f.flat_number, p.invoice_number,
		p.bill_month, p.base_amount, p.late_fee, p.discount, p.total_due, p.amount_paid,
		p.status, p.payment_method, p.gateway_order_id, p.gateway_payment_id,
		p.paid_at, p.receipt_url, p.due_date, p.reminder_count, p.created_at, p.updated_at
		FROM payment p
		JOIN flat f ON f.id = p.flat_id
		WHERE p.id = $1`

	var pay model.Payment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&pay.ID, &pay.SocietyID, &pay.FlatID, &pay.FlatNumber, &pay.InvoiceNumber,
		&pay.BillMonth, &pay.BaseAmount, &pay.LateFee, &pay.Discount, &pay.TotalDue, &pay.AmountPaid,
		&pay.Status, &pay.PaymentMethod, &pay.GatewayOrderID, &pay.GatewayPaymentID,
		&pay.PaidAt, &pay.ReceiptURL, &pay.DueDate, &pay.ReminderCount, &pay.CreatedAt, &pay.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find payment: %w", err)
	}
	return &pay, nil
}

// ListByFlat returns payment history for a specific flat with cursor pagination.
func (r *PaymentRepository) ListByFlat(ctx context.Context, flatID uuid.UUID, cursor *uuid.UUID, limit int) ([]model.Payment, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	args := []interface{}{flatID}
	where := []string{"p.flat_id = $1"}
	argIdx := 2

	if cursor != nil {
		where = append(where, fmt.Sprintf("p.created_at < (SELECT created_at FROM payment WHERE id = $%d)", argIdx))
		args = append(args, *cursor)
		argIdx++
	}

	args = append(args, limit)

	query := fmt.Sprintf(`SELECT p.id, p.society_id, p.flat_id, f.flat_number, p.invoice_number,
		p.bill_month, p.base_amount, p.late_fee, p.discount, p.total_due, p.amount_paid,
		p.status, p.payment_method, p.paid_at, p.due_date, p.reminder_count,
		p.created_at, p.updated_at
		FROM payment p
		JOIN flat f ON f.id = p.flat_id
		WHERE %s
		ORDER BY p.bill_month DESC
		LIMIT $%d`, strings.Join(where, " AND "), argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list payments by flat: %w", err)
	}
	defer rows.Close()

	return scanPaymentList(rows)
}

// ListBySociety returns all payments for a society with optional filters and cursor pagination.
func (r *PaymentRepository) ListBySociety(ctx context.Context, f model.PaymentListFilter) ([]model.Payment, error) {
	args := []interface{}{f.SocietyID}
	where := []string{"p.society_id = $1"}
	argIdx := 2

	if f.FlatID != nil {
		where = append(where, fmt.Sprintf("p.flat_id = $%d", argIdx))
		args = append(args, *f.FlatID)
		argIdx++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("p.status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}
	if f.Cursor != nil {
		where = append(where, fmt.Sprintf("p.created_at < (SELECT created_at FROM payment WHERE id = $%d)", argIdx))
		args = append(args, *f.Cursor)
		argIdx++
	}

	if f.Limit <= 0 || f.Limit > 50 {
		f.Limit = 20
	}
	args = append(args, f.Limit)

	query := fmt.Sprintf(`SELECT p.id, p.society_id, p.flat_id, f.flat_number, p.invoice_number,
		p.bill_month, p.base_amount, p.late_fee, p.discount, p.total_due, p.amount_paid,
		p.status, p.payment_method, p.paid_at, p.due_date, p.reminder_count,
		p.created_at, p.updated_at
		FROM payment p
		JOIN flat f ON f.id = p.flat_id
		WHERE %s
		ORDER BY p.bill_month DESC
		LIMIT $%d`, strings.Join(where, " AND "), argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list payments by society: %w", err)
	}
	defer rows.Close()

	return scanPaymentList(rows)
}

// UpdatePayment marks a payment as paid with gateway references.
func (r *PaymentRepository) UpdatePayment(ctx context.Context, id uuid.UUID, status, method string, gatewayOrderID, gatewayPaymentID *string, paidAt time.Time) error {
	query := `UPDATE payment
		SET status = $1, payment_method = $2, gateway_order_id = $3, gateway_payment_id = $4,
			amount_paid = total_due, paid_at = $5, updated_at = NOW()
		WHERE id = $6`

	tag, err := r.pool.Exec(ctx, query, status, method, gatewayOrderID, gatewayPaymentID, paidAt, id)
	if err != nil {
		return fmt.Errorf("update payment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("payment not found")
	}
	return nil
}

// RecordCash records a cash, cheque, or bank transfer payment by an admin.
func (r *PaymentRepository) RecordCash(ctx context.Context, id uuid.UUID, method, notes string, recordedBy uuid.UUID) error {
	query := `UPDATE payment
		SET status = 'paid', payment_method = $1, amount_paid = total_due,
			paid_at = NOW(), updated_at = NOW()
		WHERE id = $2 AND status IN ('pending', 'overdue')`

	tag, err := r.pool.Exec(ctx, query, method, id)
	if err != nil {
		return fmt.Errorf("record cash payment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("payment not found or already paid")
	}
	return nil
}

// GetPendingDues returns the total unpaid amount for a flat.
func (r *PaymentRepository) GetPendingDues(ctx context.Context, flatID uuid.UUID) (float64, error) {
	var total *float64
	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_due - amount_paid), 0)
		 FROM payment WHERE flat_id = $1 AND status IN ('pending', 'overdue')`,
		flatID).Scan(&total)
	if err != nil || total == nil {
		return 0, err
	}
	return *total, nil
}

// Defaulter represents a flat with overdue payments.
type Defaulter struct {
	FlatID       uuid.UUID `json:"flat_id"`
	FlatNumber   string    `json:"flat_number"`
	OwnerName    string    `json:"owner_name"`
	TotalOverdue float64   `json:"total_overdue"`
	MonthsOverdue int      `json:"months_overdue"`
}

// GetDefaulters returns a list of flats with overdue payments for a society.
func (r *PaymentRepository) GetDefaulters(ctx context.Context, societyID uuid.UUID) ([]Defaulter, error) {
	query := `SELECT p.flat_id, f.flat_number, COALESCE(u.name, ''),
		SUM(p.total_due - p.amount_paid), COUNT(*)
		FROM payment p
		JOIN flat f ON f.id = p.flat_id
		LEFT JOIN app_user u ON u.id = f.owner_id
		WHERE p.society_id = $1 AND p.status = 'overdue'
		GROUP BY p.flat_id, f.flat_number, u.name
		ORDER BY SUM(p.total_due - p.amount_paid) DESC`

	rows, err := r.pool.Query(ctx, query, societyID)
	if err != nil {
		return nil, fmt.Errorf("get defaulters: %w", err)
	}
	defer rows.Close()

	var defaulters []Defaulter
	for rows.Next() {
		var d Defaulter
		if err := rows.Scan(&d.FlatID, &d.FlatNumber, &d.OwnerName, &d.TotalOverdue, &d.MonthsOverdue); err != nil {
			return nil, fmt.Errorf("scan defaulter: %w", err)
		}
		defaulters = append(defaulters, d)
	}
	return defaulters, rows.Err()
}

// CreateExpense inserts a new expense record.
func (r *PaymentRepository) CreateExpense(ctx context.Context, e *model.Expense) (*model.Expense, error) {
	query := `INSERT INTO expense (society_id, category, description, amount, expense_date, receipt_url, vendor_id, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		e.SocietyID, e.Category, e.Description, e.Amount, e.ExpenseDate,
		e.ReceiptURL, e.VendorID, e.CreatedBy,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create expense: %w", err)
	}
	return e, nil
}

// ListExpenses returns expenses for a society with cursor pagination.
func (r *PaymentRepository) ListExpenses(ctx context.Context, societyID uuid.UUID, cursor *uuid.UUID, limit int) ([]model.Expense, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	args := []interface{}{societyID}
	where := []string{"e.society_id = $1"}
	argIdx := 2

	if cursor != nil {
		where = append(where, fmt.Sprintf("e.created_at < (SELECT created_at FROM expense WHERE id = $%d)", argIdx))
		args = append(args, *cursor)
		argIdx++
	}

	args = append(args, limit)

	query := fmt.Sprintf(`SELECT e.id, e.society_id, e.category, e.description, e.amount,
		e.expense_date, e.receipt_url, e.vendor_id, e.created_by, e.created_at, e.updated_at
		FROM expense e
		WHERE %s
		ORDER BY e.expense_date DESC
		LIMIT $%d`, strings.Join(where, " AND "), argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	var expenses []model.Expense
	for rows.Next() {
		var e model.Expense
		if err := rows.Scan(
			&e.ID, &e.SocietyID, &e.Category, &e.Description, &e.Amount,
			&e.ExpenseDate, &e.ReceiptURL, &e.VendorID, &e.CreatedBy, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan expense: %w", err)
		}
		expenses = append(expenses, e)
	}
	return expenses, rows.Err()
}

// GetFinancialSummary returns totals for collected, pending, and expenses for a society.
func (r *PaymentRepository) GetFinancialSummary(ctx context.Context, societyID uuid.UUID) (*model.FinancialSummary, error) {
	summary := &model.FinancialSummary{}

	err := r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount_paid), 0), COALESCE(SUM(total_due - amount_paid), 0)
		 FROM payment WHERE society_id = $1`,
		societyID).Scan(&summary.TotalCollected, &summary.TotalPending)
	if err != nil {
		return nil, fmt.Errorf("financial summary payments: %w", err)
	}

	err = r.pool.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM expense WHERE society_id = $1`,
		societyID).Scan(&summary.TotalExpenses)
	if err != nil {
		return nil, fmt.Errorf("financial summary expenses: %w", err)
	}

	summary.NetBalance = summary.TotalCollected - summary.TotalExpenses
	return summary, nil
}

// scanPaymentList scans rows into a slice of Payment (shared by list queries).
func scanPaymentList(rows pgx.Rows) ([]model.Payment, error) {
	var payments []model.Payment
	for rows.Next() {
		var pay model.Payment
		if err := rows.Scan(
			&pay.ID, &pay.SocietyID, &pay.FlatID, &pay.FlatNumber, &pay.InvoiceNumber,
			&pay.BillMonth, &pay.BaseAmount, &pay.LateFee, &pay.Discount, &pay.TotalDue, &pay.AmountPaid,
			&pay.Status, &pay.PaymentMethod, &pay.PaidAt, &pay.DueDate, &pay.ReminderCount,
			&pay.CreatedAt, &pay.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan payment: %w", err)
		}
		payments = append(payments, pay)
	}
	return payments, rows.Err()
}
