-- name: GetUserByPhone :one
SELECT id, phone, phone_hash, name, email, avatar_url, preferred_language,
       is_senior_citizen, whatsapp_opted_in, telegram_chat_id, fcm_token,
       last_active_at, is_active, created_at, updated_at
FROM app_user
WHERE phone = $1 AND is_active = true;

-- name: GetUserByID :one
SELECT id, phone, phone_hash, name, email, avatar_url, preferred_language,
       is_senior_citizen, whatsapp_opted_in, telegram_chat_id, fcm_token,
       last_active_at, is_active, created_at, updated_at
FROM app_user
WHERE id = $1 AND is_active = true;

-- name: CreateUser :one
INSERT INTO app_user (phone, phone_hash, name, preferred_language)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUserProfile :exec
UPDATE app_user
SET name = COALESCE(sqlc.narg('name'), name),
    email = COALESCE(sqlc.narg('email'), email),
    preferred_language = COALESCE(sqlc.narg('preferred_language'), preferred_language),
    is_senior_citizen = COALESCE(sqlc.narg('is_senior_citizen'), is_senior_citizen)
WHERE id = $1;

-- name: UpdateFCMToken :exec
UPDATE app_user SET fcm_token = $2 WHERE id = $1;

-- name: UpdateLastActive :exec
UPDATE app_user SET last_active_at = NOW() WHERE id = $1;

-- name: GetUserMemberships :many
SELECT usm.id, usm.user_id, usm.society_id, usm.flat_id, usm.role,
       usm.is_primary_member, usm.joined_at, usm.is_active,
       s.name as society_name, s.code as society_code
FROM user_society_membership usm
JOIN society s ON s.id = usm.society_id
WHERE usm.user_id = $1 AND usm.is_active = true;
