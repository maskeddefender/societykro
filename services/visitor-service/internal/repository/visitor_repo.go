package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/visitor-service/internal/model"
)

// VisitorRepository handles all database operations for visitors and passes.
type VisitorRepository struct {
	pool *pgxpool.Pool
}

// NewVisitorRepository creates a new VisitorRepository.
func NewVisitorRepository(pool *pgxpool.Pool) *VisitorRepository {
	return &VisitorRepository{pool: pool}
}

// Pool returns the underlying connection pool for direct queries.
func (r *VisitorRepository) Pool() *pgxpool.Pool {
	return r.pool
}

// Create inserts a new visitor entry and returns it with a generated ID.
func (r *VisitorRepository) Create(ctx context.Context, v *model.Visitor) (*model.Visitor, error) {
	query := `INSERT INTO visitor
		(society_id, flat_id, visitor_name, visitor_phone, purpose, vehicle_number, visitor_photo_url, status, logged_by_guard, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'pending', $8, 'app')
		RETURNING id, society_id, flat_id, visitor_name, visitor_phone, purpose, vehicle_number, visitor_photo_url,
		status, logged_by_guard, created_at`

	err := r.pool.QueryRow(ctx, query,
		v.SocietyID, v.FlatID, v.Name, v.Phone, v.Purpose, v.VehicleNumber, v.PhotoURL, v.LoggedBy,
	).Scan(
		&v.ID, &v.SocietyID, &v.FlatID, &v.Name, &v.Phone, &v.Purpose, &v.VehicleNumber, &v.PhotoURL,
		&v.Status, &v.LoggedBy, &v.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert visitor: %w", err)
	}
	return v, nil
}

// FindByID returns a visitor with the flat number joined.
func (r *VisitorRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Visitor, error) {
	query := `SELECT v.id, v.society_id, v.flat_id, COALESCE(f.flat_number, ''), v.visitor_name, v.visitor_phone,
		v.purpose, v.vehicle_number, v.visitor_photo_url, v.status,
		v.otp_code, v.otp_expires_at, v.approved_by, v.approved_via, v.denial_reason,
		v.logged_by_guard, v.checked_in_at, v.checked_out_at, v.created_at
		FROM visitor v
		LEFT JOIN flat f ON f.id = v.flat_id
		WHERE v.id = $1`

	var vis model.Visitor
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&vis.ID, &vis.SocietyID, &vis.FlatID, &vis.FlatNumber, &vis.Name, &vis.Phone,
		&vis.Purpose, &vis.VehicleNumber, &vis.PhotoURL, &vis.Status,
		&vis.OTPCode, &vis.OTPExpiresAt, &vis.ApprovedBy, &vis.ApprovedVia, &vis.DenyReason,
		&vis.LoggedBy, &vis.CheckedInAt, &vis.CheckedOutAt, &vis.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find visitor: %w", err)
	}
	return &vis, nil
}

// List returns visitors for a society with optional filters and cursor pagination.
func (r *VisitorRepository) List(ctx context.Context, f model.VisitorListFilter) ([]model.Visitor, error) {
	args := []interface{}{f.SocietyID}
	where := []string{"v.society_id = $1"}
	argIdx := 2

	if f.Status != nil {
		where = append(where, fmt.Sprintf("v.status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}
	if f.FlatID != nil {
		where = append(where, fmt.Sprintf("v.flat_id = $%d", argIdx))
		args = append(args, *f.FlatID)
		argIdx++
	}
	if f.Cursor != nil {
		where = append(where, fmt.Sprintf("v.created_at < (SELECT created_at FROM visitor WHERE id = $%d)", argIdx))
		args = append(args, *f.Cursor)
		argIdx++
	}

	if f.Limit <= 0 || f.Limit > 50 {
		f.Limit = 20
	}
	args = append(args, f.Limit)

	query := fmt.Sprintf(`SELECT v.id, v.society_id, v.flat_id, COALESCE(f.flat_number, ''),
		v.visitor_name, v.visitor_phone, v.purpose, v.vehicle_number, v.status,
		v.approved_via, v.checked_in_at, v.checked_out_at, v.created_at
		FROM visitor v
		LEFT JOIN flat f ON f.id = v.flat_id
		WHERE %s
		ORDER BY v.created_at DESC
		LIMIT $%d`, strings.Join(where, " AND "), argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list visitors: %w", err)
	}
	defer rows.Close()

	var visitors []model.Visitor
	for rows.Next() {
		var vis model.Visitor
		if err := rows.Scan(
			&vis.ID, &vis.SocietyID, &vis.FlatID, &vis.FlatNumber,
			&vis.Name, &vis.Phone, &vis.Purpose, &vis.VehicleNumber, &vis.Status,
			&vis.ApprovedVia, &vis.CheckedInAt, &vis.CheckedOutAt, &vis.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan visitor: %w", err)
		}
		visitors = append(visitors, vis)
	}
	return visitors, rows.Err()
}

// UpdateStatus changes the visitor status with optional approval metadata.
func (r *VisitorRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, approvedBy *uuid.UUID, approvedVia *string) error {
	var extras []string
	args := []interface{}{status}
	argIdx := 2

	switch status {
	case "approved", "checked_in":
		extras = append(extras, "checked_in_at = NOW()")
		if approvedBy != nil {
			extras = append(extras, fmt.Sprintf("approved_by = $%d", argIdx))
			args = append(args, *approvedBy)
			argIdx++
		}
		if approvedVia != nil {
			extras = append(extras, fmt.Sprintf("approved_via = $%d", argIdx))
			args = append(args, *approvedVia)
			argIdx++
		}
	case "denied":
		// deny_reason is set separately via DenyVisitor
	case "checked_out":
		extras = append(extras, "checked_out_at = NOW()")
	}

	setClause := "status = $1, updated_at = NOW()"
	if len(extras) > 0 {
		setClause += ", " + strings.Join(extras, ", ")
	}

	args = append(args, id)
	query := fmt.Sprintf(`UPDATE visitor SET %s WHERE id = $%d`, setClause, argIdx)

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update visitor status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("visitor not found")
	}
	return nil
}

// SetDenyReason updates the deny reason on a visitor record.
func (r *VisitorRepository) SetDenyReason(ctx context.Context, id uuid.UUID, reason string) error {
	query := `UPDATE visitor SET deny_reason = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, reason, id)
	return err
}

// SetOTP stores a pre-approval OTP on a visitor record.
func (r *VisitorRepository) SetOTP(ctx context.Context, id uuid.UUID, otp string, expiresAt interface{}) error {
	query := `UPDATE visitor SET otp_code = $1, otp_expires_at = $2, is_pre_approved = true WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, otp, expiresAt, id)
	return err
}

// VerifyOTP finds a visitor by society and OTP code.
func (r *VisitorRepository) VerifyOTP(ctx context.Context, societyID uuid.UUID, otpCode string) (*model.Visitor, error) {
	query := `SELECT v.id, v.society_id, v.flat_id, COALESCE(f.flat_number, ''), v.visitor_name, v.visitor_phone,
		v.purpose, v.vehicle_number, v.status, v.otp_code, v.otp_expires_at,
		v.created_at
		FROM visitor v
		LEFT JOIN flat f ON f.id = v.flat_id
		WHERE v.society_id = $1 AND v.otp_code = $2 AND v.is_pre_approved = true AND v.status = 'pending'
		LIMIT 1`

	var vis model.Visitor
	err := r.pool.QueryRow(ctx, query, societyID, otpCode).Scan(
		&vis.ID, &vis.SocietyID, &vis.FlatID, &vis.FlatNumber, &vis.Name, &vis.Phone,
		&vis.Purpose, &vis.VehicleNumber, &vis.Status, &vis.OTPCode, &vis.OTPExpiresAt,
		&vis.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("verify otp: %w", err)
	}
	return &vis, nil
}

// CreatePass inserts a new visitor pass.
func (r *VisitorRepository) CreatePass(ctx context.Context, p *model.VisitorPass) (*model.VisitorPass, error) {
	query := `INSERT INTO visitor_pass
		(flat_id, society_id, visitor_name, phone, purpose, vehicle_number, valid_from, valid_until, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		p.FlatID, p.SocietyID, p.VisitorName, p.Phone, p.Purpose, p.VehicleNumber,
		p.ValidFrom, p.ValidUntil, p.CreatedBy,
	).Scan(&p.ID, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create pass: %w", err)
	}
	return p, nil
}

// ListPasses returns active visitor passes for a flat.
func (r *VisitorRepository) ListPasses(ctx context.Context, flatID uuid.UUID) ([]model.VisitorPass, error) {
	query := `SELECT id, flat_id, society_id, visitor_name, phone, purpose, vehicle_number,
		valid_from, valid_until, created_by, created_at
		FROM visitor_pass
		WHERE flat_id = $1 AND valid_until > NOW()
		ORDER BY valid_until ASC`

	rows, err := r.pool.Query(ctx, query, flatID)
	if err != nil {
		return nil, fmt.Errorf("list passes: %w", err)
	}
	defer rows.Close()

	var passes []model.VisitorPass
	for rows.Next() {
		var p model.VisitorPass
		if err := rows.Scan(
			&p.ID, &p.FlatID, &p.SocietyID, &p.VisitorName, &p.Phone, &p.Purpose, &p.VehicleNumber,
			&p.ValidFrom, &p.ValidUntil, &p.CreatedBy, &p.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan pass: %w", err)
		}
		passes = append(passes, p)
	}
	return passes, rows.Err()
}

// DeletePass removes a visitor pass by ID.
func (r *VisitorRepository) DeletePass(ctx context.Context, passID uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM visitor_pass WHERE id = $1`, passID)
	if err != nil {
		return fmt.Errorf("delete pass: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("pass not found")
	}
	return nil
}

// ListActive returns currently checked-in visitors for a society.
func (r *VisitorRepository) ListActive(ctx context.Context, societyID uuid.UUID) ([]model.Visitor, error) {
	query := `SELECT v.id, v.society_id, v.flat_id, COALESCE(f.flat_number, ''),
		v.visitor_name, v.visitor_phone, v.purpose, v.vehicle_number, v.status,
		v.approved_via, v.checked_in_at, v.created_at
		FROM visitor v
		LEFT JOIN flat f ON f.id = v.flat_id
		WHERE v.society_id = $1 AND v.status = 'checked_in'
		ORDER BY v.checked_in_at DESC`

	rows, err := r.pool.Query(ctx, query, societyID)
	if err != nil {
		return nil, fmt.Errorf("list active visitors: %w", err)
	}
	defer rows.Close()

	var visitors []model.Visitor
	for rows.Next() {
		var vis model.Visitor
		if err := rows.Scan(
			&vis.ID, &vis.SocietyID, &vis.FlatID, &vis.FlatNumber,
			&vis.Name, &vis.Phone, &vis.Purpose, &vis.VehicleNumber, &vis.Status,
			&vis.ApprovedVia, &vis.CheckedInAt, &vis.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan active visitor: %w", err)
		}
		visitors = append(visitors, vis)
	}
	return visitors, rows.Err()
}
