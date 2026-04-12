-- Migration 004: Add streaming fields to streams table
ALTER TABLE streams 
ADD COLUMN stream_key VARCHAR(36) UNIQUE NOT NULL AFTER owner_id,
ADD COLUMN playback_url TEXT NOT NULL AFTER stream_key;

-- Ensure other required columns exist
ALTER TABLE streams 
MODIFY COLUMN owner_id VARCHAR(36) NOT NULL,
MODIFY COLUMN thumbnail_url VARCHAR(255),
MODIFY COLUMN viewers_count INT DEFAULT 0,
MODIFY COLUMN is_live BOOLEAN DEFAULT false;
