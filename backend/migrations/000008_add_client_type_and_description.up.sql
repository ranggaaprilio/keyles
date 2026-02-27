-- Enable pg_trgm extension for trigram-based ILIKE search performance
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Add client_type column (confidential or public)
ALTER TABLE clients
    ADD COLUMN client_type VARCHAR(20) NOT NULL DEFAULT 'confidential';

-- Add description column for application metadata
ALTER TABLE clients
    ADD COLUMN description TEXT;

-- Relax client_secret NOT NULL constraint for public clients
ALTER TABLE clients
    ALTER COLUMN client_secret DROP NOT NULL;

-- Add check constraint for valid client types
ALTER TABLE clients
    ADD CONSTRAINT clients_valid_client_type
    CHECK (client_type IN ('confidential', 'public'));

-- Add check constraint: confidential clients MUST have a secret
ALTER TABLE clients
    ADD CONSTRAINT clients_secret_required_for_confidential
    CHECK (client_type = 'public' OR client_secret IS NOT NULL);

-- Add index for client_type queries
CREATE INDEX idx_clients_client_type ON clients(client_type);

-- Add composite index for tenant + active + type queries
CREATE INDEX idx_clients_tenant_active_type ON clients(tenant_id, is_active, client_type);

-- Add index for client_name search (ILIKE performance)
CREATE INDEX idx_clients_client_name_trgm ON clients USING gin (client_name gin_trgm_ops);
