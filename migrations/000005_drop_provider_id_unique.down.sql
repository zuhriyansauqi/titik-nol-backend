-- 000005_drop_provider_id_unique.down.sql
-- Restore the unique constraint on provider_id
ALTER TABLE users ADD CONSTRAINT users_provider_id_key UNIQUE (provider_id);
