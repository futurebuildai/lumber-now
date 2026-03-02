DROP TRIGGER IF EXISTS enforce_request_status_transition ON requests;
DROP FUNCTION IF EXISTS check_request_status_transition();
DROP INDEX IF EXISTS idx_requests_failed_retry;
ALTER TABLE requests DROP COLUMN IF EXISTS retry_count;
ALTER TABLE requests DROP COLUMN IF EXISTS last_error;
