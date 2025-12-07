-- Drop tenants table
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS tenants CASCADE;
DROP TYPE IF EXISTS tenant_status;
DROP EXTENSION IF EXISTS "uuid-ossp";
