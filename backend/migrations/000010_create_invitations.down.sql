-- migrations/000010_create_invitations.down.sql
DROP TRIGGER IF EXISTS update_invitations_updated_at ON invitations;
DROP TABLE IF EXISTS invitations CASCADE;