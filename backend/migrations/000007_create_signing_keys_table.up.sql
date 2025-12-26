-- Create signing_keys table for JWT signing
CREATE TABLE IF NOT EXISTS signing_keys (
    kid VARCHAR(255) PRIMARY KEY,
    algorithm VARCHAR(50) NOT NULL DEFAULT 'RS256',
    private_key TEXT NOT NULL,
    public_key TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        expires_at TIMESTAMP
    WITH
        TIME ZONE,
        CONSTRAINT signing_keys_algorithm_check CHECK (algorithm IN ('RS256', 'RS384', 'RS512'))
);

-- Indexes for key lookups and rotation
CREATE INDEX idx_signing_keys_is_active ON signing_keys (is_active);

CREATE INDEX idx_signing_keys_expires_at ON signing_keys (expires_at);