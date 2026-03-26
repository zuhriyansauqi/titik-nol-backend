-- 000005_drop_provider_id_unique.up.sql
-- Drop the unique constraint on provider_id to allow local auth users
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_provider_id_key;
