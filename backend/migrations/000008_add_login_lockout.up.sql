-- Account lockout tracking columns
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;

-- Reset lockout on successful login (application-level), but add index for locked accounts
CREATE INDEX IF NOT EXISTS idx_users_locked ON users (locked_until) WHERE locked_until IS NOT NULL;
