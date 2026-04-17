package repository

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/societykro/auth-service/internal/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func HashPhone(phone string) string {
	h := sha256.Sum256([]byte(phone))
	return fmt.Sprintf("%x", h)
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*model.User, error) {
	query := `SELECT id, phone, phone_hash, name, email, avatar_url, preferred_language,
		is_senior_citizen, whatsapp_opted_in, telegram_chat_id, fcm_token,
		last_active_at, is_active, created_at, updated_at
		FROM app_user WHERE phone = $1 AND is_active = true`

	var u model.User
	err := r.pool.QueryRow(ctx, query, phone).Scan(
		&u.ID, &u.Phone, &u.PhoneHash, &u.Name, &u.Email, &u.AvatarURL,
		&u.PreferredLanguage, &u.IsSeniorCitizen, &u.WhatsappOptedIn,
		&u.TelegramChatID, &u.FCMToken, &u.LastActiveAt, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	query := `SELECT id, phone, phone_hash, name, email, avatar_url, preferred_language,
		is_senior_citizen, whatsapp_opted_in, telegram_chat_id, fcm_token,
		last_active_at, is_active, created_at, updated_at
		FROM app_user WHERE id = $1 AND is_active = true`

	var u model.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Phone, &u.PhoneHash, &u.Name, &u.Email, &u.AvatarURL,
		&u.PreferredLanguage, &u.IsSeniorCitizen, &u.WhatsappOptedIn,
		&u.TelegramChatID, &u.FCMToken, &u.LastActiveAt, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, phone, name, language string) (*model.User, error) {
	query := `INSERT INTO app_user (phone, phone_hash, name, preferred_language)
		VALUES ($1, $2, $3, $4)
		RETURNING id, phone, phone_hash, name, email, avatar_url, preferred_language,
		is_senior_citizen, whatsapp_opted_in, telegram_chat_id, fcm_token,
		last_active_at, is_active, created_at, updated_at`

	var u model.User
	err := r.pool.QueryRow(ctx, query, phone, HashPhone(phone), name, language).Scan(
		&u.ID, &u.Phone, &u.PhoneHash, &u.Name, &u.Email, &u.AvatarURL,
		&u.PreferredLanguage, &u.IsSeniorCitizen, &u.WhatsappOptedIn,
		&u.TelegramChatID, &u.FCMToken, &u.LastActiveAt, &u.IsActive,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) UpdateFCMToken(ctx context.Context, userID uuid.UUID, token string) error {
	_, err := r.pool.Exec(ctx, `UPDATE app_user SET fcm_token = $1 WHERE id = $2`, token, userID)
	return err
}

func (r *UserRepository) UpdateLastActive(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE app_user SET last_active_at = NOW() WHERE id = $1`, userID)
	return err
}

func (r *UserRepository) GetMemberships(ctx context.Context, userID uuid.UUID) ([]model.UserSocietyMembership, error) {
	query := `SELECT id, user_id, society_id, flat_id, role, is_primary_member, joined_at, is_active
		FROM user_society_membership WHERE user_id = $1 AND is_active = true`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberships []model.UserSocietyMembership
	for rows.Next() {
		var m model.UserSocietyMembership
		if err := rows.Scan(&m.ID, &m.UserID, &m.SocietyID, &m.FlatID, &m.Role, &m.IsPrimaryMember, &m.JoinedAt, &m.IsActive); err != nil {
			return nil, err
		}
		memberships = append(memberships, m)
	}
	return memberships, nil
}
