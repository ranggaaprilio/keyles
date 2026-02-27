-- Remove indexes
DROP INDEX IF EXISTS idx_clients_client_name_trgm;
DROP INDEX IF EXISTS idx_clients_tenant_active_type;
DROP INDEX IF EXISTS idx_clients_client_type;

-- Remove constraints
ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_secret_required_for_confidential;
ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_valid_client_type;

-- Restore NOT NULL on client_secret (set NULL values to empty string first)
UPDATE clients SET client_secret = '' WHERE client_secret IS NULL;
ALTER TABLE clients ALTER COLUMN client_secret SET NOT NULL;

-- Remove columns
ALTER TABLE clients DROP COLUMN IF EXISTS description;
ALTER TABLE clients DROP COLUMN IF EXISTS client_type;
