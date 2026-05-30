-- migrations/000009_extend_users.down.sql
DROP INDEX IF EXISTS idx_users_email_trgm;
DROP INDEX IF NOT EXISTS idx_users_display_name_trgm;
DROP INDEX IF NOT EXISTS idx_users_tenant_email_lower;
DROP INDEX IF NOT EXISTS idx_users_tenant_status;
DROP INDEX IF NOT EXISTS idx_users_status;

ALTER TABLE users DROP CONSTRAINT IF EXISTS users_valid_status;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS status;
ALTER TABLE users DROP COLUMN IF EXISTS display_name;