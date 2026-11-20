ALTER TABLE users ADD COLUMN display_name VARCHAR(100) NOT NULL DEFAULT '';
UPDATE users SET display_name = login;
