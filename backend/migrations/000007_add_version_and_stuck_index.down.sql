DROP TRIGGER IF EXISTS request_version_increment ON requests;
DROP FUNCTION IF EXISTS increment_request_version();
DROP INDEX IF EXISTS idx_requests_stuck_processing;
ALTER TABLE requests DROP COLUMN IF EXISTS version;
