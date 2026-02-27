-- name: GetDealer :one
SELECT * FROM dealers WHERE id = $1;

-- name: GetDealerBySlug :one
SELECT * FROM dealers WHERE slug = $1;

-- name: GetDealerBySubdomain :one
SELECT * FROM dealers WHERE subdomain = $1;

-- name: ListDealers :many
SELECT * FROM dealers ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListActiveDealers :many
SELECT * FROM dealers WHERE active = true ORDER BY name ASC;

-- name: CreateDealer :one
INSERT INTO dealers (name, slug, subdomain, logo_url, primary_color, secondary_color, contact_email, contact_phone, address)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateDealer :one
UPDATE dealers
SET name = $2, logo_url = $3, primary_color = $4, secondary_color = $5,
    contact_email = $6, contact_phone = $7, address = $8, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: SetDealerActive :exec
UPDATE dealers SET active = $2, updated_at = now() WHERE id = $1;

-- name: DeleteDealer :exec
DELETE FROM dealers WHERE id = $1;
