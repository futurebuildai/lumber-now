-- name: GetUser :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE dealer_id = $1 AND email = $2;

-- name: ListUsersByDealer :many
SELECT * FROM users WHERE dealer_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: ListUsersByRole :many
SELECT * FROM users WHERE dealer_id = $1 AND role = $2 ORDER BY full_name ASC;

-- name: ListContractorsByRep :many
SELECT * FROM users WHERE assigned_rep_id = $1 ORDER BY full_name ASC;

-- name: CreateUser :one
INSERT INTO users (dealer_id, email, password_hash, full_name, phone, role)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
SET full_name = $2, phone = $3, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: AssignContractorToRep :exec
UPDATE users SET assigned_rep_id = $2, updated_at = now() WHERE id = $1;

-- name: SetUserActive :exec
UPDATE users SET active = $2, updated_at = now() WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: IncrementLoginFailures :one
UPDATE users
SET failed_login_attempts = failed_login_attempts + 1,
    locked_until = CASE
      WHEN failed_login_attempts + 1 >= 5 THEN now() + interval '15 minutes'
      ELSE locked_until
    END,
    updated_at = now()
WHERE dealer_id = $1 AND email = $2
RETURNING failed_login_attempts, locked_until;

-- name: ResetLoginFailures :exec
UPDATE users
SET failed_login_attempts = 0, locked_until = NULL, updated_at = now()
WHERE dealer_id = $1 AND email = $2;

-- name: GetUserLockoutStatus :one
SELECT failed_login_attempts, locked_until FROM users
WHERE dealer_id = $1 AND email = $2;
