ALTER TABLE clients
    ADD COLUMN IF NOT EXISTS allow_self_signup BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS self_signup_default_role VARCHAR(100);

ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_self_signup_requires_role;
ALTER TABLE clients
    ADD CONSTRAINT clients_self_signup_requires_role CHECK (
        NOT allow_self_signup
        OR (self_signup_default_role IS NOT NULL AND btrim(self_signup_default_role) <> '')
    );

ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_signup_requested';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_signup_completed';
