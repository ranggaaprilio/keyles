ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_login_succeeded';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_login_failed';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_login_throttled';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_consent_approved';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_consent_denied';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_logout';
ALTER TYPE event_type ADD VALUE IF NOT EXISTS 'oauth_invalid_callback';
