-- Rollback initial schema

DROP INDEX IF EXISTS idx_transcriptions_status;
DROP INDEX IF EXISTS idx_transcriptions_user_id;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_sessions_token;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS transcriptions;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
