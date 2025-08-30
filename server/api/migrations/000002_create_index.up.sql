CREATE INDEX IF NOT EXISTS idx_users_search_vector ON users USING GIN(search_vector);

CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver_status
    ON friend_requests(receiver_id, status);
    
CREATE INDEX IF NOT EXISTS idx_user_friends_user ON user_friends(user_id);
CREATE INDEX IF NOT EXISTS idx_user_friends_friend ON user_friends(friend_id);

CREATE INDEX IF NOT EXISTS idx_room_users_room_user ON room_users(room_id, user_id);
CREATE INDEX IF NOT EXISTS idx_room_users_user_room ON room_users(user_id, room_id);

CREATE INDEX IF NOT EXISTS idx_room_invites_user_status ON room_invites(user_id, status);
CREATE INDEX IF NOT EXISTS idx_room_invites_room_status ON room_invites(room_id, status);

CREATE INDEX IF NOT EXISTS idx_messages_room ON messages(room_id);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);

CREATE INDEX IF NOT EXISTS idx_subscriptions_user ON subscriptions(user_id);