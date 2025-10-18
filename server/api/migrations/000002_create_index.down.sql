-- Drop all indexes created in 000002_create_index.up.sql
DROP INDEX IF EXISTS idx_subscriptions_user;

DROP INDEX IF EXISTS idx_notifications_user;

DROP INDEX IF EXISTS idx_messages_room;

DROP INDEX IF EXISTS idx_room_invites_room_status;
DROP INDEX IF EXISTS idx_room_invites_user_status;

DROP INDEX IF EXISTS idx_room_users_user_room;
DROP INDEX IF EXISTS idx_room_users_room_user;

DROP INDEX IF EXISTS idx_user_friends_friend;
DROP INDEX IF EXISTS idx_user_friends_user;

DROP INDEX IF EXISTS idx_friend_requests_receiver_status;

DROP INDEX IF EXISTS idx_users_search_vector;