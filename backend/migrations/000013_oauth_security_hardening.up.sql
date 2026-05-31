ALTER TABLE refresh_tokens
ADD COLUMN IF NOT EXISTS id BIGSERIAL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_id
ON refresh_tokens (id);

ALTER TABLE refresh_tokens
ADD COLUMN IF NOT EXISTS revoked_reason VARCHAR(255),
ADD COLUMN IF NOT EXISTS family_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS parent_token_hash VARCHAR(255),
ADD COLUMN IF NOT EXISTS replaced_by_token_hash VARCHAR(255);

UPDATE refresh_tokens
SET family_id = token
WHERE family_id IS NULL;

ALTER TABLE refresh_tokens
ALTER COLUMN family_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family_id
ON refresh_tokens (family_id);
