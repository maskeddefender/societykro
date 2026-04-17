package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/complaint-service/internal/model"
)

// ComplaintRepository handles all database operations for complaints and comments.
type ComplaintRepository struct {
	pool *pgxpool.Pool
}

// NewComplaintRepository creates a new ComplaintRepository.
func NewComplaintRepository(pool *pgxpool.Pool) *ComplaintRepository {
	return &ComplaintRepository{pool: pool}
}

// Create inserts a new complaint and returns it with a generated ticket number.
func (r *ComplaintRepository) Create(ctx context.Context, c *model.Complaint) (*model.Complaint, error) {
	// Generate ticket number: {SOCIETY_CODE}-C-{NNNN}
	var count int
	r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM complaint WHERE society_id = $1`, c.SocietyID).Scan(&count)
	c.TicketNumber = fmt.Sprintf("C-%04d", count+1)

	imageJSON, _ := json.Marshal(c.ImageURLs)
	if c.Priority == "" {
		c.Priority = "normal"
	}
	if c.Source == "" {
		c.Source = "app"
	}

	query := `INSERT INTO complaint
		(society_id, flat_id, raised_by, ticket_number, category, title, description,
		 image_urls, status, priority, is_emergency, is_common_area, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'open', $9, $10, $11, $12)
		RETURNING id, society_id, flat_id, raised_by, ticket_number, category, title, description,
		image_urls, status, priority, is_emergency, is_common_area, source,
		escalation_count, created_at, updated_at`

	var imgRaw json.RawMessage
	err := r.pool.QueryRow(ctx, query,
		c.SocietyID, c.FlatID, c.RaisedBy, c.TicketNumber, c.Category, c.Title, c.Description,
		imageJSON, c.Priority, c.IsEmergency, c.IsCommonArea, c.Source,
	).Scan(
		&c.ID, &c.SocietyID, &c.FlatID, &c.RaisedBy, &c.TicketNumber, &c.Category, &c.Title, &c.Description,
		&imgRaw, &c.Status, &c.Priority, &c.IsEmergency, &c.IsCommonArea, &c.Source,
		&c.EscalationCount, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert complaint: %w", err)
	}
	json.Unmarshal(imgRaw, &c.ImageURLs)
	return c, nil
}

// FindByID returns a complaint with the raiser's name and assigned vendor name.
func (r *ComplaintRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Complaint, error) {
	query := `SELECT c.id, c.society_id, c.flat_id, c.raised_by, u.name, c.ticket_number,
		c.category, c.title, c.description, c.description_original, c.description_english,
		c.original_language, c.voice_url, c.image_urls, c.status, c.priority,
		c.is_emergency, c.is_common_area, c.assigned_vendor_id, COALESCE(v.name, ''),
		c.assigned_at, c.resolved_at, c.closed_at, c.resolution_rating, c.resolution_feedback,
		c.source, c.escalation_count, c.created_at, c.updated_at
		FROM complaint c
		JOIN app_user u ON u.id = c.raised_by
		LEFT JOIN vendor v ON v.id = c.assigned_vendor_id
		WHERE c.id = $1`

	var comp model.Complaint
	var imgRaw json.RawMessage
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&comp.ID, &comp.SocietyID, &comp.FlatID, &comp.RaisedBy, &comp.RaisedByName,
		&comp.TicketNumber, &comp.Category, &comp.Title, &comp.Description,
		&comp.DescriptionOriginal, &comp.DescriptionEnglish, &comp.OriginalLanguage,
		&comp.VoiceURL, &imgRaw, &comp.Status, &comp.Priority,
		&comp.IsEmergency, &comp.IsCommonArea, &comp.AssignedVendorID, &comp.AssignedVendorName,
		&comp.AssignedAt, &comp.ResolvedAt, &comp.ClosedAt,
		&comp.ResolutionRating, &comp.ResolutionFeedback,
		&comp.Source, &comp.EscalationCount, &comp.CreatedAt, &comp.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find complaint: %w", err)
	}
	json.Unmarshal(imgRaw, &comp.ImageURLs)
	return &comp, nil
}

// List returns complaints for a society with optional filters and cursor pagination.
func (r *ComplaintRepository) List(ctx context.Context, f model.ComplaintListFilter) ([]model.Complaint, error) {
	args := []interface{}{f.SocietyID}
	where := []string{"c.society_id = $1"}
	argIdx := 2

	if f.Status != nil {
		where = append(where, fmt.Sprintf("c.status = $%d", argIdx))
		args = append(args, *f.Status)
		argIdx++
	}
	if f.Category != nil {
		where = append(where, fmt.Sprintf("c.category = $%d", argIdx))
		args = append(args, *f.Category)
		argIdx++
	}
	if f.FlatID != nil {
		where = append(where, fmt.Sprintf("c.flat_id = $%d", argIdx))
		args = append(args, *f.FlatID)
		argIdx++
	}
	if f.Cursor != nil {
		where = append(where, fmt.Sprintf("c.created_at < (SELECT created_at FROM complaint WHERE id = $%d)", argIdx))
		args = append(args, *f.Cursor)
		argIdx++
	}

	if f.Limit <= 0 || f.Limit > 50 {
		f.Limit = 20
	}
	args = append(args, f.Limit)

	query := fmt.Sprintf(`SELECT c.id, c.society_id, c.flat_id, c.raised_by, u.name,
		c.ticket_number, c.category, c.title, c.description, c.image_urls,
		c.status, c.priority, c.is_emergency, c.source, c.escalation_count,
		c.created_at, c.updated_at
		FROM complaint c
		JOIN app_user u ON u.id = c.raised_by
		WHERE %s
		ORDER BY c.created_at DESC
		LIMIT $%d`, strings.Join(where, " AND "), argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list complaints: %w", err)
	}
	defer rows.Close()

	var complaints []model.Complaint
	for rows.Next() {
		var comp model.Complaint
		var imgRaw json.RawMessage
		if err := rows.Scan(
			&comp.ID, &comp.SocietyID, &comp.FlatID, &comp.RaisedBy, &comp.RaisedByName,
			&comp.TicketNumber, &comp.Category, &comp.Title, &comp.Description, &imgRaw,
			&comp.Status, &comp.Priority, &comp.IsEmergency, &comp.Source, &comp.EscalationCount,
			&comp.CreatedAt, &comp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan complaint: %w", err)
		}
		json.Unmarshal(imgRaw, &comp.ImageURLs)
		complaints = append(complaints, comp)
	}
	return complaints, rows.Err()
}

// UpdateStatus changes the complaint status and records timestamps.
func (r *ComplaintRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, userID uuid.UUID) error {
	var extra string
	switch status {
	case "resolved":
		extra = ", resolved_at = NOW()"
	case "closed":
		extra = ", closed_at = NOW()"
	case "reopened":
		extra = ", reopened_at = NOW(), escalation_count = escalation_count + 1"
	}

	query := fmt.Sprintf(`UPDATE complaint SET status = $1 %s WHERE id = $2`, extra)
	tag, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("complaint not found")
	}
	return nil
}

// AssignVendor assigns a vendor to a complaint.
func (r *ComplaintRepository) AssignVendor(ctx context.Context, complaintID, vendorID, assignedBy uuid.UUID) error {
	query := `UPDATE complaint SET assigned_vendor_id = $1, assigned_by = $2, assigned_at = NOW(),
		status = CASE WHEN status = 'open' THEN 'in_progress' ELSE status END
		WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, vendorID, assignedBy, complaintID)
	return err
}

// Rate records a resolution rating for a complaint.
func (r *ComplaintRepository) Rate(ctx context.Context, id uuid.UUID, rating int, feedback *string) error {
	query := `UPDATE complaint SET resolution_rating = $1, resolution_feedback = $2 WHERE id = $3`
	_, err := r.pool.Exec(ctx, query, rating, feedback, id)
	return err
}

// AddComment inserts a comment into the complaint thread.
func (r *ComplaintRepository) AddComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	query := `INSERT INTO complaint_comment (complaint_id, user_id, comment, image_url, is_internal, is_status_change, old_status, new_status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		comment.ComplaintID, comment.UserID, comment.Comment, comment.ImageURL,
		comment.IsInternal, comment.IsStatusChange, comment.OldStatus, comment.NewStatus,
	).Scan(&comment.ID, &comment.CreatedAt)
	return comment, err
}

// ListComments returns all comments for a complaint.
func (r *ComplaintRepository) ListComments(ctx context.Context, complaintID uuid.UUID, includeInternal bool) ([]model.Comment, error) {
	where := "cc.complaint_id = $1"
	if !includeInternal {
		where += " AND cc.is_internal = false"
	}

	query := fmt.Sprintf(`SELECT cc.id, cc.complaint_id, cc.user_id, u.name, cc.comment, cc.image_url,
		cc.is_internal, cc.is_status_change, cc.old_status, cc.new_status, cc.created_at
		FROM complaint_comment cc
		JOIN app_user u ON u.id = cc.user_id
		WHERE %s ORDER BY cc.created_at ASC`, where)

	rows, err := r.pool.Query(ctx, query, complaintID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.ComplaintID, &c.UserID, &c.UserName, &c.Comment, &c.ImageURL,
			&c.IsInternal, &c.IsStatusChange, &c.OldStatus, &c.NewStatus, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}

// GetStats returns complaint counts by status for a society.
func (r *ComplaintRepository) GetStats(ctx context.Context, societyID uuid.UUID) (map[string]int, error) {
	query := `SELECT status, COUNT(*) FROM complaint WHERE society_id = $1 GROUP BY status`
	rows, err := r.pool.Query(ctx, query, societyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[string]int{"open": 0, "in_progress": 0, "resolved": 0, "closed": 0}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[status] = count
	}
	return stats, rows.Err()
}

// GetAvgResolutionTime returns the average resolution time in hours.
func (r *ComplaintRepository) GetAvgResolutionTime(ctx context.Context, societyID uuid.UUID) (float64, error) {
	var avg *float64
	err := r.pool.QueryRow(ctx,
		`SELECT EXTRACT(EPOCH FROM AVG(resolved_at - created_at))/3600
		 FROM complaint WHERE society_id = $1 AND resolved_at IS NOT NULL`,
		societyID).Scan(&avg)
	if err != nil || avg == nil {
		return 0, err
	}
	return *avg, nil
}
