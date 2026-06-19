INSERT INTO users (id, email, password_hash, role)
VALUES ('00000000-0000-0000-0000-000000000001', 'admin@eeip.com', 'dummy_hash', 'Super Admin')
ON CONFLICT (email) DO NOTHING;
