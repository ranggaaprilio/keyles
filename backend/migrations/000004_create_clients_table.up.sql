-- Create clients table for OAuth2/OIDC client applications
CREATE TABLE IF NOT EXISTS clients (
    client_id VARCHAR(255) PRIMARY KEY,
    tenant_id VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    client_secret VARCHAR(255) NOT NULL,
    client_name VARCHAR(255) NOT NULL,
    redirect_uris TEXT[] NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT clients_redirect_uris_not_empty CHECK (array_length(redirect_uris, 1) > 0)
);

-- Indexes for performance
CREATE INDEX idx_clients_tenant ON clients(tenant_id);
CREATE INDEX idx_clients_is_active ON clients(is_active);

-- Trigger to auto-update updated_at timestamp
CREATE TRIGGER update_clients_updated_at 
    BEFORE UPDATE ON clients
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
