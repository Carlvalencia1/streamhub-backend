CREATE TABLE IF NOT EXISTS device_tokens (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    token TEXT NOT NULL,
    platform VARCHAR(50) NOT NULL DEFAULT 'android',
    device_id VARCHAR(255),
    app_version VARCHAR(50),
    is_valid BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE KEY unique_user_token (user_id, token(100)),
    INDEX idx_user_id (user_id),
    INDEX idx_is_valid (is_valid),
    INDEX idx_created_at (created_at)
);
