CREATE TABLE IF NOT EXISTS followers (
    id          VARCHAR(36) PRIMARY KEY,
    follower_id VARCHAR(36) NOT NULL,
    streamer_id VARCHAR(36) NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_follow (follower_id, streamer_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (streamer_id) REFERENCES users(id) ON DELETE CASCADE
);
