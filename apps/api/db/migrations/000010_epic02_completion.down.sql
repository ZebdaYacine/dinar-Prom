DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS notification_outbox;
DROP TABLE IF EXISTS auth_policies;
DROP INDEX IF EXISTS users_google_subject_idx;
ALTER TABLE users DROP COLUMN IF EXISTS google_subject;
