ALTER TABLE users 
ADD COLUMN IF NOT EXISTS role VARCHAR(50) NOT NULL DEFAULT 'USER',
ADD COLUMN IF NOT EXISTS password VARCHAR(255);

-- Insert the default administrator
-- Password is 'admin123!' (bcrypt hash)
INSERT INTO users (id, email, name, provider, provider_id, role, password, created_at, updated_at)
VALUES (
    gen_random_uuid(), 
    'admin@titiknol.com', 
    'System Administrator', 
    'LOCAL', 
    'local-admin', 
    'ADMIN', 
    '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 
    NOW(), 
    NOW()
) 
ON CONFLICT (email) DO NOTHING;
