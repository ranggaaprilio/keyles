-- migrations/000009_extend_users.up.sql
-- Extends the existing users table with display name, account status, and last-login tracking.
-- Prerequisite: CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add display_name for UI presentation
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS display_name VARCHAR(255);

-- Add status with pending/active/disabled lifecycle
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'active';

ALTER TABLE users
    ADD CONSTRAINT users_valid_status
    CHECK (status IN ('pending', 'active', 'disabled'));

-- Track last successful authentication
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP WITH TIME ZONE;

-- Index on status for admin filter queries
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Composite index for tenant + status admin queries
CREATE INDEX IF NOT EXISTS idx_users_tenant_status ON users(tenant_id, status);

-- Composite index for tenant + email (case-insensitive search)
CREATE INDEX IF NOT EXISTS idx_users_tenant_email_lower ON users(tenant_id, LOWER(email));

-- Trigram index for display_name and email search (requires pg_trgm)
CREATE INDEX IF NOT EXISTS idx_users_display_name_trgm ON users USING gin (display_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_email_trgm ON users USING gin (email gin_trgm_ops);
