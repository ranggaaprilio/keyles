-- migrations/000010_create_invitations.up.sql
CREATE TABLE IF NOT EXISTS invitations (
    id               VARCHAR(255) PRIMARY KEY,
    tenant_id        VARCHAR(255) NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email            VARCHAR(255) NOT NULL,
    display_name     VARCHAR(255),
    token_hash       VARCHAR(255) NOT NULL UNIQUE,  -- bcrypt hash of the invitation token
    status           VARCHAR(20)  NOT NULL DEFAULT 'pending',
    invited_by       VARCHAR(255) NOT NULL REFERENCES users(id),
    expires_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at      TIMESTAMP WITH TIME ZONE,
    created_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT invitations_valid_status CHECK (status IN ('pending', 'accepted', 'expired')),
    -- At most one pending invitation per email per tenant
    CONSTRAINT invitations_unique_pending_email UNIQUE (tenant_id, email, status)
        DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX idx_invitations_tenant ON invitations(tenant_id);
CREATE INDEX idx_invitations_email ON invitations(tenant_id, email);
CREATE INDEX idx_invitations_status ON invitations(status);
CREATE INDEX idx_invitations_expires_at ON invitations(expires_at);

CREATE TRIGGER update_invitations_updated_at BEFORE UPDATE ON invitations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();