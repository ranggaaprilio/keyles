ALTER TABLE clients DROP CONSTRAINT IF EXISTS clients_self_signup_requires_role;
ALTER TABLE clients
    DROP COLUMN IF EXISTS self_signup_default_role,
    DROP COLUMN IF EXISTS allow_self_signup;

-- PostgreSQL enum values cannot be removed safely without rebuilding the type.
-- Keep the OAuth signup audit values when rolling back this migration.
