-- name: CheckUsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = $1);

-- name: CheckEmailExistsForRole :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 AND role = $2);

-- name: CreateUser :one
INSERT INTO users (id, username, password_hash, email, role, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING *;

-- name: GetUserByUsernameAndRole :one
SELECT id, username, password_hash, email, role, created_at
FROM users 
WHERE username = $1 AND role = $2;

-- name: GetUserByID :one
SELECT id, username, email, role, created_at
FROM users 
WHERE id = $1;

-- name: VerifyAdminByID :one
SELECT id, username, role
FROM users 
WHERE id = $1 AND role = 'admin';

-- name: VerifyUserByID :one
SELECT id, username, role
FROM users 
WHERE id = $1 AND role = 'user';

-- name: GetUsersByRole :many
SELECT id, username, email, role, created_at
FROM users 
WHERE role = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;