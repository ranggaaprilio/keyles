-- Create audit_logs table
CREATE TYPE event_type AS ENUM (
    'registration_attempt',
    'registration_success',
    'registration_failure',
    'otp_generated',
    'otp_sent',
    'otp_verified',
    'otp_expired',
    'otp_failed',
    'login_attempt',
    'login_success',
    'login_failure',
    'logout',
    'tenant_activated',
    'tenant_suspended'
);

CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NULL REFERENCES tenants(id) ON DELETE SET NULL,
    user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
    event_type event_type NOT NULL,
    event_data JSONB NULL,
    ip_address VARCHAR(45) NULL,
    user_agent TEXT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_event_data ON audit_logs USING gin(event_data);

-- Comments for documentation
COMMENT ON TABLE audit_logs IS 'Security audit trail for all tenant lifecycle events';
COMMENT ON COLUMN audit_logs.event_data IS 'Additional event context stored as JSON';
COMMENT ON COLUMN audit_logs.ip_address IS 'IPv4 or IPv6 address of requester';
