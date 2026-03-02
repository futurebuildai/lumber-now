-- Add retry tracking for dead-letter queue pattern
ALTER TABLE requests
  ADD COLUMN retry_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN last_error TEXT NOT NULL DEFAULT '';

-- DB-level state machine: only allow valid status transitions
-- pending → processing, failed
-- processing → parsed, failed
-- parsed → confirmed
-- confirmed → sent
-- failed → pending (retry)
CREATE OR REPLACE FUNCTION check_request_status_transition()
RETURNS TRIGGER AS $$
BEGIN
  -- Allow the same status (idempotent updates)
  IF OLD.status = NEW.status THEN
    RETURN NEW;
  END IF;

  -- Valid transitions
  IF (OLD.status = 'pending' AND NEW.status IN ('processing', 'failed')) OR
     (OLD.status = 'processing' AND NEW.status IN ('parsed', 'failed')) OR
     (OLD.status = 'parsed' AND NEW.status = 'confirmed') OR
     (OLD.status = 'confirmed' AND NEW.status = 'sent') OR
     (OLD.status = 'failed' AND NEW.status = 'pending') THEN
    RETURN NEW;
  END IF;

  RAISE EXCEPTION 'invalid status transition from % to %', OLD.status, NEW.status;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_request_status_transition
  BEFORE UPDATE OF status ON requests
  FOR EACH ROW
  EXECUTE FUNCTION check_request_status_transition();

-- Index for worker dead-letter query
CREATE INDEX idx_requests_failed_retry ON requests (status, retry_count) WHERE status = 'failed';
