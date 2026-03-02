-- Add optimistic concurrency control via version column
ALTER TABLE requests ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Index for recovering stuck processing requests
CREATE INDEX IF NOT EXISTS idx_requests_stuck_processing
  ON requests (status, updated_at)
  WHERE status = 'processing';

-- Trigger to auto-increment version on every update
CREATE OR REPLACE FUNCTION increment_request_version()
RETURNS TRIGGER AS $$
BEGIN
  NEW.version = OLD.version + 1;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER request_version_increment
  BEFORE UPDATE ON requests
  FOR EACH ROW
  EXECUTE FUNCTION increment_request_version();
