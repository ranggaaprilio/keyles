-- migrations/000011_extend_user_role_assignments.down.sql
DROP INDEX IF EXISTS idx_ura_user_client_active;

ALTER TABLE user_role_assignments
DROP COLUMN IF EXISTS revoked_by;

ALTER TABLE user_role_assignments
DROP COLUMN IF EXISTS revoked_at;