-- migrations/000012_create_user_events.up.sql
CREATE TABLE IF NOT EXISTS user_events (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     VARCHAR(255) NOT NULL,
    user_id       VARCHAR(255) NOT NULL,
    client_id     VARCHAR(255),
    event_type    VARCHAR(50)  NOT NULL,
    ip_address    INET,
    country_code  CHAR(2),
    details       JSONB        NOT NULL DEFAULT '{}',
    occurred_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT user_events_valid_type CHECK (event_type IN (
        'login_success',
        'login_failure',
        'token_refresh',
        'logout',
        'session_terminated',
        'account_disabled',
        'account_enabled',
        'role_assigned',
        'role_revoked',
        'user_invited',
        'invitation_accepted',
        'invitation_expired',
        'invitation_resent',
        'user_deleted'
    ))
);

-- Primary access pattern: list events for a user, most recent first
CREATE INDEX idx_user_events_user_recent
    ON user_events(user_id, occurred_at DESC);

-- Admin filtering by tenant + event type
CREATE INDEX idx_user_events_tenant_type
    ON user_events(tenant_id, event_type, occurred_at DESC);

-- Cleanup: find events older than retention window
CREATE INDEX idx_user_events_occurred_at
    ON user_events(occurred_at);
