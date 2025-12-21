-- Drop users table
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TABLE IF EXISTS users CASCADE;
DROP TYPE IF EXISTS user_role;
