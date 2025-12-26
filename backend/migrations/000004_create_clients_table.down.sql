-- Drop clients table and associated trigger
DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;

DROP TABLE IF EXISTS clients CASCADE;