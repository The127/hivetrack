-- Replace full unique constraint on email with a partial one
-- that only applies to non-empty emails, allowing multiple users
-- with empty email (upsert is on sub, not email)
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_nonempty ON users (email) WHERE email <> '';
