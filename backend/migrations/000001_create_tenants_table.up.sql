-- Create tenants table
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TYPE tenant_status AS ENUM ('pending_verification', 'active', 'suspended', 'deleted');

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_name VARCHAR(100) NOT NULL,
    status tenant_status NOT NULL DEFAULT 'pending_verification',
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    verified_at TIMESTAMP NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE UNIQUE INDEX idx_tenants_org_name ON tenants(LOWER(organization_name));
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_created_at ON tenants(created_at DESC);

-- Updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE tenants IS 'Multi-tenant organizations using the SSO platform';
COMMENT ON COLUMN tenants.organization_name IS 'Unique organization name (case-insensitive)';
COMMENT ON COLUMN tenants.status IS 'Tenant lifecycle status';
COMMENT ON COLUMN tenants.verified_at IS 'Timestamp when email verification completed';
