-- Create refresh_tokens table for OAuth2 refresh flow
CREATE TABLE IF NOT EXISTS refresh_tokens (
    token VARCHAR(255) PRIMARY KEY,
    client_id VARCHAR(255) NOT NULL REFERENCES clients (client_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    scope TEXT NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        expires_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL,
        last_used_at TIMESTAMP
    WITH
        TIME ZONE,
        is_revoked BOOLEAN NOT NULL DEFAULT false,
        revoked_at TIMESTAMP
    WITH
        TIME ZONE,
        revoked_by VARCHAR(255)
);

-- Indexes for performance and cleanup
CREATE INDEX idx_refresh_tokens_client ON refresh_tokens (client_id);

CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);

CREATE INDEX idx_refresh_tokens_tenant ON refresh_tokens (tenant_id);

CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

CREATE INDEX idx_refresh_tokens_is_revoked ON refresh_tokens (is_revoked);

CREATE INDEX idx_refresh_tokens_user_client ON refresh_tokens (user_id, client_id);