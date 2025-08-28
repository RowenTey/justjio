-- CREATE EXTENSION IF NOT EXISTS pg_cron;

CREATE MATERIALIZED VIEW IF NOT EXISTS user_non_friends AS
    SELECT u.id AS user_id, v.id AS non_friend_id FROM users u
    JOIN users v ON u.id != v.id
    LEFT JOIN user_friends uf ON uf.user_id = u.id AND uf.friend_id = v.id
    WHERE uf.friend_id IS NULL;

-- required for CONCURRENTLY refresh
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_non_friends 
    ON user_non_friends(user_id, non_friend_id);
    
-- -- background cron task to sync
-- SELECT cron.schedule('refresh_user_non_friends', '*/5 * * * *', $$
--     REFRESH MATERIALIZED VIEW CONCURRENTLY user_non_friends;
-- $$);