package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/notice-service/internal/model"
)

// NoticeRepository handles database operations for notices and read receipts.
type NoticeRepository struct {
	pool *pgxpool.Pool
}

// NewNoticeRepository creates a new NoticeRepository.
func NewNoticeRepository(pool *pgxpool.Pool) *NoticeRepository {
	return &NoticeRepository{pool: pool}
}

// Create inserts a new notice.
func (r *NoticeRepository) Create(ctx context.Context, n *model.Notice) (*model.Notice, error) {
	attachJSON, _ := json.Marshal(n.AttachmentURLs)
	if n.Category == "" {
		n.Category = "general"
	}

	query := `INSERT INTO notice (society_id, created_by, title, body, category, is_pinned,
		broadcast_whatsapp, broadcast_telegram, attachment_urls, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err := r.pool.QueryRow(ctx, query,
		n.SocietyID, n.CreatedBy, n.Title, n.Body, n.Category, n.IsPinned,
		n.BroadcastWhatsapp, n.BroadcastTelegram, attachJSON, n.ExpiresAt,
	).Scan(&n.ID, &n.CreatedAt)
	return n, err
}

// FindByID returns a notice with its read count.
func (r *NoticeRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.Notice, error) {
	query := `SELECT n.id, n.society_id, n.created_by, u.name, n.title, n.body, n.category,
		n.is_pinned, n.broadcast_whatsapp, n.broadcast_telegram, n.attachment_urls,
		n.expires_at, n.created_at,
		(SELECT COUNT(*) FROM notice_read_receipt WHERE notice_id = n.id) as read_count,
		(SELECT COUNT(*) FROM user_society_membership WHERE society_id = n.society_id AND is_active = true) as total_members
		FROM notice n
		JOIN app_user u ON u.id = n.created_by
		WHERE n.id = $1`

	var n model.Notice
	var attachRaw json.RawMessage
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&n.ID, &n.SocietyID, &n.CreatedBy, &n.CreatedByName, &n.Title, &n.Body, &n.Category,
		&n.IsPinned, &n.BroadcastWhatsapp, &n.BroadcastTelegram, &attachRaw,
		&n.ExpiresAt, &n.CreatedAt, &n.ReadCount, &n.TotalMembers,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find notice: %w", err)
	}
	json.Unmarshal(attachRaw, &n.AttachmentURLs)
	return &n, nil
}

// List returns notices for a society, pinned first, then by created_at desc.
func (r *NoticeRepository) List(ctx context.Context, societyID uuid.UUID, cursor *uuid.UUID, limit int) ([]model.Notice, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	args := []interface{}{societyID}
	where := "n.society_id = $1"
	argIdx := 2

	if cursor != nil {
		where += fmt.Sprintf(` AND n.created_at < (SELECT created_at FROM notice WHERE id = $%d)`, argIdx)
		args = append(args, *cursor)
		argIdx++
	}

	args = append(args, limit)
	query := fmt.Sprintf(`SELECT n.id, n.society_id, n.created_by, u.name, n.title, n.body, n.category,
		n.is_pinned, n.attachment_urls, n.created_at,
		(SELECT COUNT(*) FROM notice_read_receipt WHERE notice_id = n.id) as read_count
		FROM notice n
		JOIN app_user u ON u.id = n.created_by
		WHERE %s
		ORDER BY n.is_pinned DESC, n.created_at DESC
		LIMIT $%d`, where, argIdx)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notices []model.Notice
	for rows.Next() {
		var n model.Notice
		var attachRaw json.RawMessage
		if err := rows.Scan(
			&n.ID, &n.SocietyID, &n.CreatedBy, &n.CreatedByName, &n.Title, &n.Body, &n.Category,
			&n.IsPinned, &attachRaw, &n.CreatedAt, &n.ReadCount,
		); err != nil {
			return nil, err
		}
		json.Unmarshal(attachRaw, &n.AttachmentURLs)
		notices = append(notices, n)
	}
	return notices, rows.Err()
}

// MarkRead records that a user has read a notice. Idempotent.
func (r *NoticeRepository) MarkRead(ctx context.Context, noticeID, userID uuid.UUID, channel string) error {
	if channel == "" {
		channel = "app"
	}
	_, err := r.pool.Exec(ctx,
		`INSERT INTO notice_read_receipt (notice_id, user_id, channel)
		 VALUES ($1, $2, $3) ON CONFLICT (notice_id, user_id) DO NOTHING`,
		noticeID, userID, channel)
	return err
}

// Delete removes a notice by ID (hard delete since notices have no soft-delete).
func (r *NoticeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM notice WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("notice not found")
	}
	return nil
}
