package repository

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/auth-service/internal/model"
)

// SocietyRepository handles all database operations for societies, flats, and memberships.
type SocietyRepository struct {
	pool *pgxpool.Pool
}

// NewSocietyRepository creates a SocietyRepository backed by the given connection pool.
func NewSocietyRepository(pool *pgxpool.Pool) *SocietyRepository {
	return &SocietyRepository{pool: pool}
}

// Create inserts a new society with an auto-generated unique join code.
func (r *SocietyRepository) Create(ctx context.Context, name, address, city, state, pincode string, totalFlats int) (*model.Society, error) {
	code := generateCode()
	query := `INSERT INTO society (name, code, address, city, state, pincode, total_flats)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, name, code, address, city, state, pincode, total_flats, subscription,
		default_language, maintenance_amount, maintenance_due_day, is_active, created_at`

	var s model.Society
	err := r.pool.QueryRow(ctx, query, name, code, address, city, state, pincode, totalFlats).Scan(
		&s.ID, &s.Name, &s.Code, &s.Address, &s.City, &s.State, &s.Pincode,
		&s.TotalFlats, &s.Subscription, &s.DefaultLanguage,
		&s.MaintenanceAmount, &s.MaintenanceDueDay, &s.IsActive, &s.CreatedAt,
	)
	return &s, err
}

// FindByID returns a society by its UUID.
func (r *SocietyRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Society, error) {
	query := `SELECT id, name, code, address, city, state, pincode, total_flats, subscription,
		default_language, maintenance_amount, maintenance_due_day, is_active, created_at
		FROM society WHERE id = $1 AND is_active = true`

	var s model.Society
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.Name, &s.Code, &s.Address, &s.City, &s.State, &s.Pincode,
		&s.TotalFlats, &s.Subscription, &s.DefaultLanguage,
		&s.MaintenanceAmount, &s.MaintenanceDueDay, &s.IsActive, &s.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// FindByCode returns a society by its unique join code.
func (r *SocietyRepository) FindByCode(ctx context.Context, code string) (*model.Society, error) {
	query := `SELECT id, name, code, address, city, state, pincode, total_flats, subscription,
		default_language, maintenance_amount, maintenance_due_day, is_active, created_at
		FROM society WHERE code = $1 AND is_active = true`

	var s model.Society
	err := r.pool.QueryRow(ctx, query, code).Scan(
		&s.ID, &s.Name, &s.Code, &s.Address, &s.City, &s.State, &s.Pincode,
		&s.TotalFlats, &s.Subscription, &s.DefaultLanguage,
		&s.MaintenanceAmount, &s.MaintenanceDueDay, &s.IsActive, &s.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// AddMember inserts a user-society membership. Returns error if already exists.
func (r *SocietyRepository) AddMember(ctx context.Context, userID, societyID uuid.UUID, flatID *uuid.UUID, role string) (*model.UserSocietyMembership, error) {
	isPrimary := role == "admin" || role == "secretary" || role == "president"
	query := `INSERT INTO user_society_membership (user_id, society_id, flat_id, role, is_primary_member)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, society_id) DO UPDATE SET is_active = true, role = $4, flat_id = COALESCE($3, user_society_membership.flat_id)
		RETURNING id, user_id, society_id, flat_id, role, is_primary_member, joined_at, is_active`

	var m model.UserSocietyMembership
	err := r.pool.QueryRow(ctx, query, userID, societyID, flatID, role, isPrimary).Scan(
		&m.ID, &m.UserID, &m.SocietyID, &m.FlatID, &m.Role, &m.IsPrimaryMember, &m.JoinedAt, &m.IsActive,
	)
	return &m, err
}

// ListFlats returns all flats for a society, ordered by block and number.
func (r *SocietyRepository) ListFlats(ctx context.Context, societyID uuid.UUID) ([]model.Flat, error) {
	query := `SELECT id, society_id, flat_number, block, floor, flat_type, is_occupied, occupancy
		FROM flat WHERE society_id = $1 ORDER BY block, flat_number`

	rows, err := r.pool.Query(ctx, query, societyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flats []model.Flat
	for rows.Next() {
		var f model.Flat
		if err := rows.Scan(&f.ID, &f.SocietyID, &f.FlatNumber, &f.Block, &f.Floor, &f.FlatType, &f.IsOccupied, &f.Occupancy); err != nil {
			return nil, err
		}
		flats = append(flats, f)
	}
	return flats, rows.Err()
}

// FindFlatByNumber returns a single flat by society ID and flat number.
func (r *SocietyRepository) FindFlatByNumber(ctx context.Context, societyID uuid.UUID, flatNumber string) (*model.Flat, error) {
	query := `SELECT id, society_id, flat_number, block, floor, flat_type, is_occupied, occupancy
		FROM flat WHERE society_id = $1 AND flat_number = $2`

	var f model.Flat
	err := r.pool.QueryRow(ctx, query, societyID, flatNumber).Scan(
		&f.ID, &f.SocietyID, &f.FlatNumber, &f.Block, &f.Floor, &f.FlatType, &f.IsOccupied, &f.Occupancy,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &f, err
}

// GenerateFlats auto-creates flats for a society based on block/floor/unit layout.
// Example: 2 blocks, 4 floors, 3 flats/floor = A-101, A-102, A-103, A-201, ..., B-403
func (r *SocietyRepository) GenerateFlats(ctx context.Context, societyID uuid.UUID, blocks, floors, flatsPerFloor int) error {
	batch := &pgx.Batch{}
	blockNames := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for b := 0; b < blocks && b < len(blockNames); b++ {
		block := string(blockNames[b])
		for f := 1; f <= floors; f++ {
			for u := 1; u <= flatsPerFloor; u++ {
				flatNum := fmt.Sprintf("%s-%d%02d", block, f, u)
				batch.Queue(
					`INSERT INTO flat (society_id, flat_number, block, floor, flat_type)
					 VALUES ($1, $2, $3, $4, 'apartment') ON CONFLICT DO NOTHING`,
					societyID, flatNum, "Block "+block, f,
				)
			}
		}
	}

	if batch.Len() == 0 {
		return nil
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()
	for i := 0; i < batch.Len(); i++ {
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("batch insert flat %d: %w", i, err)
		}
	}
	return nil
}

// GetMembersBySociety returns all active members of a society with user details.
func (r *SocietyRepository) GetMembersBySociety(ctx context.Context, societyID uuid.UUID) ([]map[string]interface{}, error) {
	query := `SELECT u.id, u.name, u.phone, usm.role, usm.is_primary_member, f.flat_number
		FROM user_society_membership usm
		JOIN app_user u ON u.id = usm.user_id
		LEFT JOIN flat f ON f.id = usm.flat_id
		WHERE usm.society_id = $1 AND usm.is_active = true
		ORDER BY usm.role, u.name`

	rows, err := r.pool.Query(ctx, query, societyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, phone, role string
		var isPrimary bool
		var flatNumber *string
		if err := rows.Scan(&id, &name, &phone, &role, &isPrimary, &flatNumber); err != nil {
			return nil, err
		}
		members = append(members, map[string]interface{}{
			"id":               id,
			"name":             name,
			"phone":            phone,
			"role":             role,
			"is_primary_member": isPrimary,
			"flat_number":      flatNumber,
		})
	}
	return members, rows.Err()
}

func generateCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		code[i] = chars[rand.Intn(len(chars))]
	}
	return string(code)
}
