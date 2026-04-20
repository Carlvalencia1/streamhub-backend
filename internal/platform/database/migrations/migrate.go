package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

var migrations = []struct {
	name string
	sql  string
}{
	{
		name: "001_create_users_table",
		sql: `CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    avatar_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);`,
	},
	{
		name: "002_create_streams_table",
		sql: `CREATE TABLE IF NOT EXISTS streams (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);`,
	},
	{
		name: "003_create_messages_table",
		sql: `CREATE TABLE IF NOT EXISTS messages (
    id VARCHAR(36) PRIMARY KEY,
    stream_id VARCHAR(36) NOT NULL,
    user_id VARCHAR(36) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (stream_id) REFERENCES streams(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);`,
	},
	{
		name: "005_create_device_tokens_table",
		sql: `CREATE TABLE IF NOT EXISTS device_tokens (
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
);`,
	},
	// {
	// 	name: "004_add_streaming_fields",
	// 	sql: `ALTER TABLE streams
	// ADD COLUMN IF NOT EXISTS thumbnail_url VARCHAR(255),
	// ADD COLUMN IF NOT EXISTS category VARCHAR(100),
	// ADD COLUMN IF NOT EXISTS owner_id VARCHAR(36),
	// ADD COLUMN IF NOT EXISTS viewers_count INT DEFAULT 0,
	// ADD COLUMN IF NOT EXISTS is_live BOOLEAN DEFAULT false,
	// ADD COLUMN IF NOT EXISTS started_at TIMESTAMP NULL,
	// ADD COLUMN IF NOT EXISTS stream_key VARCHAR(36) UNIQUE NOT NULL,
	// ADD COLUMN IF NOT EXISTS playback_url TEXT NOT NULL;`,
	// },
	{
		name: "006_add_role_to_users",
		sql:  `ALTER TABLE users ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'viewer';`,
	},
	{
		name: "007_create_followers_table",
		sql: `CREATE TABLE IF NOT EXISTS followers (
    id          VARCHAR(36) PRIMARY KEY,
    follower_id VARCHAR(36) NOT NULL,
    streamer_id VARCHAR(36) NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_follow (follower_id, streamer_id),
    FOREIGN KEY (follower_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (streamer_id) REFERENCES users(id) ON DELETE CASCADE
);`,
	},
	{
		name: "008_add_profile_fields_to_users",
		sql: `ALTER TABLE users
    ADD COLUMN nickname VARCHAR(100) NULL,
    ADD COLUMN bio TEXT NULL,
    ADD COLUMN location VARCHAR(100) NULL;`,
	},
	{
		name: "009_add_banner_url_to_users",
		sql:  `ALTER TABLE users ADD COLUMN banner_url TEXT NULL;`,
	},
	{
		name: "010_create_communities_table",
		sql: `CREATE TABLE IF NOT EXISTS communities (
    id          VARCHAR(36) PRIMARY KEY,
    owner_id    VARCHAR(36) NOT NULL,
    name        VARCHAR(100) NOT NULL,
    description TEXT NULL,
    image_url   TEXT NULL,
    invite_code VARCHAR(20) NOT NULL UNIQUE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);`,
	},
	{
		name: "011_create_community_members_table",
		sql: `CREATE TABLE IF NOT EXISTS community_members (
    id           VARCHAR(36) PRIMARY KEY,
    community_id VARCHAR(36) NOT NULL,
    user_id      VARCHAR(36) NOT NULL,
    role         VARCHAR(20) NOT NULL DEFAULT 'member',
    joined_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_member (community_id, user_id),
    FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`,
	},
	{
		name: "012_create_community_channels_table",
		sql: `CREATE TABLE IF NOT EXISTS community_channels (
    id           VARCHAR(36) PRIMARY KEY,
    community_id VARCHAR(36) NOT NULL,
    name         VARCHAR(100) NOT NULL,
    description  TEXT NULL,
    created_at   TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (community_id) REFERENCES communities(id) ON DELETE CASCADE
);`,
	},
	{
		name: "013_create_channel_messages_table",
		sql: `CREATE TABLE IF NOT EXISTS channel_messages (
    id         VARCHAR(36) PRIMARY KEY,
    channel_id VARCHAR(36) NOT NULL,
    user_id    VARCHAR(36) NOT NULL,
    type       VARCHAR(20) NOT NULL DEFAULT 'text',
    content    TEXT,
    media_url  TEXT,
    poll_id    VARCHAR(36),
    expires_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (channel_id) REFERENCES community_channels(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_channel_created (channel_id, created_at)
);`,
	},
	{
		name: "014_create_polls_table",
		sql: `CREATE TABLE IF NOT EXISTS polls (
    id              VARCHAR(36) PRIMARY KEY,
    question        VARCHAR(500) NOT NULL,
    options         JSON NOT NULL,
    multiple_choice BOOLEAN DEFAULT false,
    expires_at      TIMESTAMP NULL,
    created_at      TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`,
	},
	{
		name: "015_create_poll_votes_table",
		sql: `CREATE TABLE IF NOT EXISTS poll_votes (
    id           VARCHAR(36) PRIMARY KEY,
    poll_id      VARCHAR(36) NOT NULL,
    user_id      VARCHAR(36) NOT NULL,
    option_index INT NOT NULL,
    UNIQUE KEY uq_vote (poll_id, user_id, option_index),
    FOREIGN KEY (poll_id) REFERENCES polls(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);`,
	},
	{
		name: "016_create_channel_settings_table",
		sql: `CREATE TABLE IF NOT EXISTS channel_settings (
    channel_id              VARCHAR(36) PRIMARY KEY,
    disappearing_ttl_seconds INT NOT NULL DEFAULT 0,
    FOREIGN KEY (channel_id) REFERENCES community_channels(id) ON DELETE CASCADE
);`,
	},
	{
		name: "017_create_channel_posts_table",
		sql: `CREATE TABLE IF NOT EXISTS channel_posts (
    id          VARCHAR(36) PRIMARY KEY,
    streamer_id VARCHAR(36) NOT NULL,
    type        VARCHAR(20) NOT NULL DEFAULT 'text',
    content     TEXT,
    media_url   TEXT,
    poll_id     VARCHAR(36),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (streamer_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_streamer_created (streamer_id, created_at DESC)
);`,
	},
}

func Run(db *sql.DB) error {
	log.Println("Running migrations...")

	for _, m := range migrations {
		log.Printf("Executing migration: %s", m.name)

		if _, err := db.Exec(m.sql); err != nil {
			errMsg := err.Error()
			// Skip if column/table/key already exists (idempotent migrations)
			if strings.Contains(errMsg, "Duplicate column name") ||
				strings.Contains(errMsg, "already exists") ||
				strings.Contains(errMsg, "Duplicate key name") ||
				strings.Contains(errMsg, "1060") ||
				strings.Contains(errMsg, "1050") ||
				strings.Contains(errMsg, "1061") {
				log.Printf("Migration %s skipped (already applied)", m.name)
				continue
			}
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}

		log.Printf("Migration %s completed", m.name)
	}

	log.Println("All migrations completed successfully")
	return nil
}
