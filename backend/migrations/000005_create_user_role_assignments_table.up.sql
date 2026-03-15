-- Create user_role_assignments table for RBAC
CREATE TABLE IF NOT EXISTS user_role_assignments (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    client_id VARCHAR(255) NOT NULL REFERENCES clients (client_id) ON DELETE CASCADE,
    tenant_id UUID NOT NULL REFERENCES tenants (id) ON DELETE CASCADE,
    role VARCHAR(100) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    granted_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        granted_by VARCHAR(255),
        CONSTRAINT user_role_assignments_unique UNIQUE (user_id, client_id, role)
);

-- Indexes for performance
CREATE INDEX idx_user_role_assignments_user ON user_role_assignments (user_id);

CREATE INDEX idx_user_role_assignments_client ON user_role_assignments (client_id);

CREATE INDEX idx_user_role_assignments_tenant ON user_role_assignments (tenant_id);

CREATE INDEX idx_user_role_assignments_user_client ON user_role_assignments (user_id, client_id);

CREATE INDEX idx_user_role_assignments_is_active ON user_role_assignments (is_active);