-- Remove the seeded admin
DELETE FROM users WHERE email = 'admin@titiknol.com';

-- Remove the columns
ALTER TABLE users 
DROP COLUMN IF EXISTS role,
DROP COLUMN IF EXISTS password;
