-- name: GetRequest :one
SELECT * FROM requests WHERE id = $1;

-- name: ListRequestsByDealer :many
SELECT * FROM requests WHERE dealer_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: ListRequestsByContractor :many
SELECT * FROM requests WHERE contractor_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: ListRequestsByRep :many
SELECT * FROM requests WHERE assigned_rep_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: ListRequestsByStatus :many
SELECT * FROM requests WHERE dealer_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4;

-- name: CreateRequest :one
INSERT INTO requests (dealer_id, contractor_id, assigned_rep_id, input_type, raw_text, media_url)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateRequestStatus :one
UPDATE requests
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateRequestStructuredItems :one
UPDATE requests
SET structured_items = $2, ai_confidence = $3, status = 'parsed', updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateRequestNotes :exec
UPDATE requests SET notes = $2, updated_at = now() WHERE id = $1;

-- name: AssignRequestToRep :exec
UPDATE requests SET assigned_rep_id = $2, updated_at = now() WHERE id = $1;

-- name: DeleteRequest :exec
DELETE FROM requests WHERE id = $1;

-- name: CountRequestsByDealer :one
SELECT count(*) FROM requests WHERE dealer_id = $1;

-- name: CountRequestsByStatus :one
SELECT count(*) FROM requests WHERE dealer_id = $1 AND status = $2;
