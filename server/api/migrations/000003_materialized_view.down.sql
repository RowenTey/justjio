-- Drop the unique index first
DROP INDEX IF EXISTS idx_user_non_friends;

DROP MATERIALIZED VIEW IF EXISTS user_non_friends;
