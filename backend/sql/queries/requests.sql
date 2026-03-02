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

-- name: UpdateRequestRawText :exec
UPDATE requests SET raw_text = $2, updated_at = now() WHERE id = $1;

-- name: GetRequestByDealer :one
SELECT * FROM requests WHERE id = $1 AND dealer_id = $2;

-- name: UpdateRequestNotesByDealer :exec
UPDATE requests SET notes = $2, updated_at = now() WHERE id = $1 AND dealer_id = $3;

-- name: ClaimPendingRequests :many
UPDATE requests SET status = 'processing', updated_at = now()
WHERE id IN (
  SELECT id FROM requests
  WHERE status = 'pending'
  ORDER BY created_at ASC
  LIMIT $1
  FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: RetryFailedRequests :many
UPDATE requests SET status = 'pending', retry_count = retry_count + 1, updated_at = now()
WHERE id IN (
  SELECT id FROM requests
  WHERE status = 'failed' AND retry_count < $1
  ORDER BY updated_at ASC
  LIMIT $2
  FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: SetRequestFailed :exec
UPDATE requests SET status = 'failed', last_error = $2, retry_count = retry_count, updated_at = now()
WHERE id = $1;

-- name: RecoverStuckRequests :one
WITH recovered AS (
  UPDATE requests SET status = 'pending', last_error = 'recovered from stuck processing state', updated_at = now()
  WHERE status = 'processing' AND updated_at < now() - $1::interval
  LIMIT $2
  RETURNING id
)
SELECT count(*) FROM recovered;

-- Optimistic concurrency: version-checked status update
-- name: UpdateRequestStatusVersioned :one
UPDATE requests
SET status = $2, updated_at = now()
WHERE id = $1 AND version = $3
RETURNING *;

-- Optimistic concurrency: version-checked structured items update
-- name: UpdateRequestStructuredItemsVersioned :one
UPDATE requests
SET structured_items = $2, ai_confidence = $3, status = 'parsed', updated_at = now()
WHERE id = $1 AND version = $4
RETURNING *;
