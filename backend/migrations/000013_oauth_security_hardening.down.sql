DROP INDEX IF EXISTS idx_refresh_tokens_family_id;
DROP INDEX IF EXISTS idx_refresh_tokens_id;

ALTER TABLE refresh_tokens
DROP COLUMN IF EXISTS replaced_by_token_hash,
DROP COLUMN IF EXISTS parent_token_hash,
DROP COLUMN IF EXISTS family_id,
DROP COLUMN IF EXISTS revoked_reason,
DROP COLUMN IF EXISTS id;
