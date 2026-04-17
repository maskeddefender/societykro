package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/vendor-service/internal/model"
)

// VendorRepository handles all database operations for vendors and domestic help.
type VendorRepository struct {
	pool *pgxpool.Pool
}

// NewVendorRepository creates a new VendorRepository.
func NewVendorRepository(pool *pgxpool.Pool) *VendorRepository {
	return &VendorRepository{pool: pool}
}

// --------------- Vendor CRUD ---------------

// CreateVendor inserts a new vendor and returns it.
func (r *VendorRepository) CreateVendor(ctx context.Context, v *model.Vendor) (*model.Vendor, error) {
	query := `INSERT INTO vendor
		(society_id, name, company_name, phone, whatsapp_phone, category, sub_category, address)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, society_id, name, company_name, phone, whatsapp_phone, category, sub_category,
		address, avg_rating, total_jobs, completed_jobs, response_time_avg_hrs, is_active, created_at, updated_at`

	err := r.pool.QueryRow(ctx, query,
		v.SocietyID, v.Name, v.CompanyName, v.Phone, v.WhatsappPhone,
		v.Category, v.SubCategory, v.Address,
	).Scan(
		&v.ID, &v.SocietyID, &v.Name, &v.CompanyName, &v.Phone, &v.WhatsappPhone,
		&v.Category, &v.SubCategory, &v.Address, &v.AvgRating, &v.TotalJobs,
		&v.CompletedJobs, &v.ResponseTimeAvgHrs, &v.IsActive, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert vendor: %w", err)
	}
	return v, nil
}

// FindVendorByID returns a vendor by its ID.
func (r *VendorRepository) FindVendorByID(ctx context.Context, id uuid.UUID) (*model.Vendor, error) {
	query := `SELECT id, society_id, name, company_name, phone, whatsapp_phone, category, sub_category,
		address, avg_rating, total_jobs, completed_jobs, response_time_avg_hrs, is_active, created_at, updated_at
		FROM vendor WHERE id = $1`

	var v model.Vendor
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.SocietyID, &v.Name, &v.CompanyName, &v.Phone, &v.WhatsappPhone,
		&v.Category, &v.SubCategory, &v.Address, &v.AvgRating, &v.TotalJobs,
		&v.CompletedJobs, &v.ResponseTimeAvgHrs, &v.IsActive, &v.CreatedAt, &v.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find vendor: %w", err)
	}
	return &v, nil
}

// ListVendors returns vendors for a society with optional category filter and cursor pagination.
func (r *VendorRepository) ListVendors(ctx context.Context, f model.VendorListFilter) ([]model.Vendor, error) {
	args := []interface{}{f.SocietyID}
	where := []string{"society_id = $1"}
	argIdx := 2

	if f.Category != nil {
		where = append(where, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *f.Category)
		argIdx++
	}
	if f.SubCategory != nil {
		where = append(where, fmt.Sprintf("sub_category = $%d", argIdx))
		args = append(args, *f.SubCategory)
		argIdx++
	}
	if f.IsActive != nil {
		where = append(where, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *f.IsActive)
		argIdx++
	}
	if f.Cursor != nil {
		where = append(where, fmt.Sprintf("created_at < (SELECT created_at FROM vendor WHERE id = $%d)", argIdx))
		args = append(args, *f.Cursor)
		argIdx++
	}

	if f.Limit <= 0 || f.Limit > 50 {
		f.Limit = 20
	}
	args = append(args, f.Limit)

	query := fmt.Sprintf(`SELECT id, society_id, name, company_name, phone, whatsapp_phone, category, sub_category,
		address, avg_rating, total_jobs, completed_jobs, response_time_avg_hrs, is_active, created_at, updated_at
		FROM vendor
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d`, strings.Join(where, " AND "), argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list vendors: %w", err)
	}
	defer rows.Close()

	var vendors []model.Vendor
	for rows.Next() {
		var v model.Vendor
		if err := rows.Scan(
			&v.ID, &v.SocietyID, &v.Name, &v.CompanyName, &v.Phone, &v.WhatsappPhone,
			&v.Category, &v.SubCategory, &v.Address, &v.AvgRating, &v.TotalJobs,
			&v.CompletedJobs, &v.ResponseTimeAvgHrs, &v.IsActive, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan vendor: %w", err)
		}
		vendors = append(vendors, v)
	}
	return vendors, rows.Err()
}

// UpdateVendor updates the mutable fields of a vendor.
func (r *VendorRepository) UpdateVendor(ctx context.Context, id uuid.UUID, req model.UpdateVendorRequest) error {
	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.CompanyName != nil {
		sets = append(sets, fmt.Sprintf("company_name = $%d", argIdx))
		args = append(args, *req.CompanyName)
		argIdx++
	}
	if req.Phone != nil {
		sets = append(sets, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *req.Phone)
		argIdx++
	}
	if req.WhatsappPhone != nil {
		sets = append(sets, fmt.Sprintf("whatsapp_phone = $%d", argIdx))
		args = append(args, *req.WhatsappPhone)
		argIdx++
	}
	if req.Category != nil {
		sets = append(sets, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, *req.Category)
		argIdx++
	}
	if req.SubCategory != nil {
		sets = append(sets, fmt.Sprintf("sub_category = $%d", argIdx))
		args = append(args, *req.SubCategory)
		argIdx++
	}
	if req.Address != nil {
		sets = append(sets, fmt.Sprintf("address = $%d", argIdx))
		args = append(args, *req.Address)
		argIdx++
	}
	if req.IsActive != nil {
		sets = append(sets, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}

	if len(sets) == 0 {
		return nil
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, id)

	query := fmt.Sprintf(`UPDATE vendor SET %s WHERE id = $%d`, strings.Join(sets, ", "), argIdx)
	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update vendor: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("vendor not found")
	}
	return nil
}

// DeleteVendor performs a soft delete by setting is_active to false.
func (r *VendorRepository) DeleteVendor(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE vendor SET is_active = false, updated_at = NOW() WHERE id = $1`
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete vendor: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("vendor not found")
	}
	return nil
}

// UpdateVendorStats increments total_jobs and recalculates avg_rating.
func (r *VendorRepository) UpdateVendorStats(ctx context.Context, id uuid.UUID, rating float64, completed bool) error {
	var completedInc int
	if completed {
		completedInc = 1
	}
	query := `UPDATE vendor SET
		total_jobs = total_jobs + 1,
		completed_jobs = completed_jobs + $1,
		avg_rating = CASE WHEN total_jobs = 0 THEN $2
			ELSE (avg_rating * total_jobs + $2) / (total_jobs + 1) END,
		updated_at = NOW()
		WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, completedInc, rating, id)
	return err
}

// --------------- Domestic Help ---------------

// CreateDomesticHelp inserts a new domestic help record.
func (r *VendorRepository) CreateDomesticHelp(ctx context.Context, h *model.DomesticHelp) (*model.DomesticHelp, error) {
	if h.EntryMethod == "" {
		h.EntryMethod = "gate"
	}

	query := `INSERT INTO domestic_help (society_id, name, phone, photo_url, role, entry_method)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, society_id, name, phone, photo_url, role, is_verified, avg_rating, entry_method, is_active, created_at`

	err := r.pool.QueryRow(ctx, query,
		h.SocietyID, h.Name, h.Phone, h.PhotoURL, h.Role, h.EntryMethod,
	).Scan(
		&h.ID, &h.SocietyID, &h.Name, &h.Phone, &h.PhotoURL, &h.Role,
		&h.IsVerified, &h.AvgRating, &h.EntryMethod, &h.IsActive, &h.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert domestic help: %w", err)
	}
	return h, nil
}

// FindDomesticHelpByID returns a domestic help record by ID.
func (r *VendorRepository) FindDomesticHelpByID(ctx context.Context, id uuid.UUID) (*model.DomesticHelp, error) {
	query := `SELECT id, society_id, name, phone, photo_url, role, is_verified, avg_rating, entry_method, is_active, created_at
		FROM domestic_help WHERE id = $1`

	var h model.DomesticHelp
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&h.ID, &h.SocietyID, &h.Name, &h.Phone, &h.PhotoURL, &h.Role,
		&h.IsVerified, &h.AvgRating, &h.EntryMethod, &h.IsActive, &h.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find domestic help: %w", err)
	}
	return &h, nil
}

// ListDomesticHelpBySociety returns all active domestic help for a society.
func (r *VendorRepository) ListDomesticHelpBySociety(ctx context.Context, societyID uuid.UUID) ([]model.DomesticHelp, error) {
	query := `SELECT id, society_id, name, phone, photo_url, role, is_verified, avg_rating, entry_method, is_active, created_at
		FROM domestic_help WHERE society_id = $1 AND is_active = true ORDER BY name ASC`

	rows, err := r.pool.Query(ctx, query, societyID)
	if err != nil {
		return nil, fmt.Errorf("list domestic help by society: %w", err)
	}
	defer rows.Close()

	var helpers []model.DomesticHelp
	for rows.Next() {
		var h model.DomesticHelp
		if err := rows.Scan(
			&h.ID, &h.SocietyID, &h.Name, &h.Phone, &h.PhotoURL, &h.Role,
			&h.IsVerified, &h.AvgRating, &h.EntryMethod, &h.IsActive, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan domestic help: %w", err)
		}
		helpers = append(helpers, h)
	}
	return helpers, rows.Err()
}

// ListDomesticHelpByFlat returns domestic help linked to a specific flat.
func (r *VendorRepository) ListDomesticHelpByFlat(ctx context.Context, flatID uuid.UUID) ([]model.DomesticHelp, error) {
	query := `SELECT dh.id, dh.society_id, dh.name, dh.phone, dh.photo_url, dh.role,
		dh.is_verified, dh.avg_rating, dh.entry_method, dh.is_active, dh.created_at
		FROM domestic_help dh
		JOIN domestic_help_flat dhf ON dhf.domestic_help_id = dh.id
		WHERE dhf.flat_id = $1 AND dhf.is_active = true AND dh.is_active = true
		ORDER BY dh.name ASC`

	rows, err := r.pool.Query(ctx, query, flatID)
	if err != nil {
		return nil, fmt.Errorf("list domestic help by flat: %w", err)
	}
	defer rows.Close()

	var helpers []model.DomesticHelp
	for rows.Next() {
		var h model.DomesticHelp
		if err := rows.Scan(
			&h.ID, &h.SocietyID, &h.Name, &h.Phone, &h.PhotoURL, &h.Role,
			&h.IsVerified, &h.AvgRating, &h.EntryMethod, &h.IsActive, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan domestic help: %w", err)
		}
		helpers = append(helpers, h)
	}
	return helpers, rows.Err()
}

// LinkHelpToFlat links a domestic helper to a flat with monthly pay.
func (r *VendorRepository) LinkHelpToFlat(ctx context.Context, link *model.DomesticHelpFlat) (*model.DomesticHelpFlat, error) {
	query := `INSERT INTO domestic_help_flat (domestic_help_id, flat_id, monthly_pay)
		VALUES ($1, $2, $3)
		ON CONFLICT (domestic_help_id, flat_id) WHERE is_active = true DO UPDATE SET monthly_pay = $3
		RETURNING id, domestic_help_id, flat_id, monthly_pay, working_days, is_active`

	err := r.pool.QueryRow(ctx, query,
		link.DomesticHelpID, link.FlatID, link.MonthlyPay,
	).Scan(
		&link.ID, &link.DomesticHelpID, &link.FlatID, &link.MonthlyPay, &link.WorkingDays, &link.IsActive,
	)
	if err != nil {
		return nil, fmt.Errorf("link help to flat: %w", err)
	}
	return link, nil
}

// UnlinkHelpFromFlat deactivates the link between a domestic helper and a flat.
func (r *VendorRepository) UnlinkHelpFromFlat(ctx context.Context, helpID, flatID uuid.UUID) error {
	query := `UPDATE domestic_help_flat SET is_active = false WHERE domestic_help_id = $1 AND flat_id = $2`
	_, err := r.pool.Exec(ctx, query, helpID, flatID)
	return err
}

// --------------- Attendance ---------------

// LogAttendance records an entry for domestic help at a flat.
func (r *VendorRepository) LogAttendance(ctx context.Context, att *model.DomesticHelpAttendance) (*model.DomesticHelpAttendance, error) {
	query := `INSERT INTO domestic_help_attendance (domestic_help_id, society_id, flat_id, entry_at, date)
		VALUES ($1, $2, $3, NOW(), CURRENT_DATE)
		RETURNING id, domestic_help_id, society_id, flat_id, entry_at, exit_at, date`

	err := r.pool.QueryRow(ctx, query,
		att.DomesticHelpID, att.SocietyID, att.FlatID,
	).Scan(
		&att.ID, &att.DomesticHelpID, &att.SocietyID, &att.FlatID, &att.EntryAt, &att.ExitAt, &att.Date,
	)
	if err != nil {
		return nil, fmt.Errorf("log attendance: %w", err)
	}
	return att, nil
}

// LogExit records exit time for an attendance record.
func (r *VendorRepository) LogExit(ctx context.Context, attendanceID uuid.UUID) error {
	query := `UPDATE domestic_help_attendance SET exit_at = NOW() WHERE id = $1 AND exit_at IS NULL`
	tag, err := r.pool.Exec(ctx, query, attendanceID)
	if err != nil {
		return fmt.Errorf("log exit: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("attendance record not found or already exited")
	}
	return nil
}

// GetAttendance returns attendance records for a domestic helper in a given month.
func (r *VendorRepository) GetAttendance(ctx context.Context, helpID uuid.UUID, year int, month time.Month) ([]model.DomesticHelpAttendance, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)

	query := `SELECT id, domestic_help_id, society_id, flat_id, entry_at, exit_at, date
		FROM domestic_help_attendance
		WHERE domestic_help_id = $1 AND date >= $2 AND date < $3
		ORDER BY date ASC, entry_at ASC`

	rows, err := r.pool.Query(ctx, query, helpID, start, end)
	if err != nil {
		return nil, fmt.Errorf("get attendance: %w", err)
	}
	defer rows.Close()

	var records []model.DomesticHelpAttendance
	for rows.Next() {
		var a model.DomesticHelpAttendance
		if err := rows.Scan(
			&a.ID, &a.DomesticHelpID, &a.SocietyID, &a.FlatID, &a.EntryAt, &a.ExitAt, &a.Date,
		); err != nil {
			return nil, fmt.Errorf("scan attendance: %w", err)
		}
		records = append(records, a)
	}
	return records, rows.Err()
}
