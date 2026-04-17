-- name: CreateSociety :one
INSERT INTO society (name, code, address, city, state, pincode, total_flats)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSocietyByID :one
SELECT * FROM society WHERE id = $1 AND is_active = true;

-- name: GetSocietyByCode :one
SELECT * FROM society WHERE code = $1 AND is_active = true;

-- name: UpdateSociety :exec
UPDATE society
SET name = COALESCE(sqlc.narg('name'), name),
    address = COALESCE(sqlc.narg('address'), address),
    maintenance_amount = COALESCE(sqlc.narg('maintenance_amount'), maintenance_amount),
    maintenance_due_day = COALESCE(sqlc.narg('maintenance_due_day'), maintenance_due_day),
    late_fee_percent = COALESCE(sqlc.narg('late_fee_percent'), late_fee_percent),
    default_language = COALESCE(sqlc.narg('default_language'), default_language)
WHERE id = $1;

-- name: ListFlatsBySociety :many
SELECT id, society_id, flat_number, block, floor, flat_type, is_occupied, occupancy
FROM flat
WHERE society_id = $1
ORDER BY block, flat_number;

-- name: CreateFlat :one
INSERT INTO flat (society_id, flat_number, block, floor, flat_type, bedrooms)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: AddMember :one
INSERT INTO user_society_membership (user_id, society_id, flat_id, role, is_primary_member)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMembersBySociety :many
SELECT usm.*, u.name as user_name, u.phone as user_phone, f.flat_number
FROM user_society_membership usm
JOIN app_user u ON u.id = usm.user_id
LEFT JOIN flat f ON f.id = usm.flat_id
WHERE usm.society_id = $1 AND usm.is_active = true
ORDER BY usm.role, u.name;
