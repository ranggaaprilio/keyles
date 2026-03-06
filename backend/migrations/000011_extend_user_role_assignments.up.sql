-- migrations/000011_extend_user_role_assignments.up.sql

-- Track who revoked an assignment and when
ALTER TABLE user_role_assignments
    ADD COLUMN IF NOT EXISTS revoked_at TIMESTAMP WITH TIME ZONE;

ALTER TABLE user_role_assignments
    ADD COLUMN IF NOT EXISTS revoked_by VARCHAR(255) REFERENCES users(id);

-- Index for active-only role lookups (most common query path)
CREATE INDEX IF NOT EXISTS idx_ura_user_client_active
    ON user_role_assignments(user_id, client_id)
    WHERE is_active = true;
