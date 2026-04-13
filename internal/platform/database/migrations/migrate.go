package migrations

import (
	"database/sql"
	"fmt"
	"log"
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
}

func Run(db *sql.DB) error {
	log.Println("Running migrations...")

	for _, m := range migrations {
		log.Printf("Executing migration: %s", m.name)

		if _, err := db.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %s failed: %w", m.name, err)
		}

		log.Printf("Migration %s completed", m.name)
	}

	log.Println("All migrations completed successfully")
	return nil
}
