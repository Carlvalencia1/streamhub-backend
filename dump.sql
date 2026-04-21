-- StreamHub Database Schema Dump
-- Generated from source code analysis
-- Database: streamhub

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ---------------------------------------------------------------------------
-- users
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `users` (
  `id`         VARCHAR(36)  NOT NULL,
  `username`   VARCHAR(50)  NOT NULL,
  `email`      VARCHAR(100) NOT NULL,
  `password`   VARCHAR(255) NOT NULL,
  `role`       VARCHAR(20)  NOT NULL DEFAULT 'viewer',
  `nickname`   VARCHAR(100)          DEFAULT NULL,
  `bio`        TEXT                  DEFAULT NULL,
  `location`   VARCHAR(100)          DEFAULT NULL,
  `avatar_url` TEXT                  DEFAULT NULL,
  `banner_url` TEXT                  DEFAULT NULL,
  `created_at` TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_username` (`username`),
  UNIQUE KEY `uq_email`    (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- streams
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `streams` (
  `id`            VARCHAR(36)  NOT NULL,
  `title`         VARCHAR(255) NOT NULL,
  `description`   TEXT                  DEFAULT NULL,
  `thumbnail_url` VARCHAR(255)          DEFAULT NULL,
  `category`      VARCHAR(100)          DEFAULT NULL,
  `owner_id`      VARCHAR(36)  NOT NULL,
  `stream_key`    VARCHAR(36)  NOT NULL,
  `playback_url`  TEXT         NOT NULL,
  `is_live`       BOOLEAN      NOT NULL DEFAULT FALSE,
  `viewers_count` INT          NOT NULL DEFAULT 0,
  `started_at`    DATETIME              DEFAULT NULL,
  `ended_at`      DATETIME              DEFAULT NULL,
  `recording_url` VARCHAR(500)          DEFAULT NULL,
  `created_at`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`    TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_stream_key` (`stream_key`),
  CONSTRAINT `fk_streams_owner` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- followers
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `followers` (
  `follower_id`  VARCHAR(36) NOT NULL,
  `streamer_id`  VARCHAR(36) NOT NULL,
  `created_at`   TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`follower_id`, `streamer_id`),
  CONSTRAINT `fk_followers_follower`  FOREIGN KEY (`follower_id`)  REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_followers_streamer`  FOREIGN KEY (`streamer_id`)  REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- device_tokens
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `device_tokens` (
  `id`          VARCHAR(36)  NOT NULL,
  `user_id`     VARCHAR(36)  NOT NULL,
  `token`       TEXT         NOT NULL,
  `platform`    VARCHAR(50)  NOT NULL DEFAULT 'android',
  `device_id`   VARCHAR(255)          DEFAULT NULL,
  `app_version` VARCHAR(50)           DEFAULT NULL,
  `is_valid`    BOOLEAN               DEFAULT TRUE,
  `last_used_at` TIMESTAMP            DEFAULT NULL,
  `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_user_token` (`user_id`, `token`(100)),
  INDEX `idx_user_id`   (`user_id`),
  INDEX `idx_is_valid`  (`is_valid`),
  INDEX `idx_created_at` (`created_at`),
  CONSTRAINT `fk_device_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- communities
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `communities` (
  `id`          VARCHAR(36)  NOT NULL,
  `owner_id`    VARCHAR(36)  NOT NULL,
  `name`        VARCHAR(255) NOT NULL,
  `description` TEXT                  DEFAULT NULL,
  `image_url`   TEXT                  DEFAULT NULL,
  `invite_code` VARCHAR(36)  NOT NULL,
  `created_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at`  TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_invite_code` (`invite_code`),
  CONSTRAINT `fk_communities_owner` FOREIGN KEY (`owner_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- community_members
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `community_members` (
  `id`           VARCHAR(36) NOT NULL,
  `community_id` VARCHAR(36) NOT NULL,
  `user_id`      VARCHAR(36) NOT NULL,
  `role`         VARCHAR(50) NOT NULL DEFAULT 'member',
  `joined_at`    TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_community_user` (`community_id`, `user_id`),
  CONSTRAINT `fk_cm_community` FOREIGN KEY (`community_id`) REFERENCES `communities` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_cm_user`      FOREIGN KEY (`user_id`)      REFERENCES `users`       (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- community_channels
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `community_channels` (
  `id`           VARCHAR(36)  NOT NULL,
  `community_id` VARCHAR(36)  NOT NULL,
  `name`         VARCHAR(255) NOT NULL,
  `description`  TEXT                  DEFAULT NULL,
  `created_at`   TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_channels_community` FOREIGN KEY (`community_id`) REFERENCES `communities` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- polls
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `polls` (
  `id`              VARCHAR(36)  NOT NULL,
  `question`        TEXT         NOT NULL,
  `options`         JSON         NOT NULL,
  `multiple_choice` BOOLEAN      NOT NULL DEFAULT FALSE,
  `expires_at`      DATETIME              DEFAULT NULL,
  `created_at`      TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- channel_messages
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `channel_messages` (
  `id`         VARCHAR(36)  NOT NULL,
  `channel_id` VARCHAR(36)  NOT NULL,
  `user_id`    VARCHAR(36)  NOT NULL,
  `type`       VARCHAR(50)  NOT NULL DEFAULT 'text',
  `content`    TEXT                  DEFAULT NULL,
  `media_url`  TEXT                  DEFAULT NULL,
  `poll_id`    VARCHAR(36)           DEFAULT NULL,
  `expires_at` DATETIME              DEFAULT NULL,
  `created_at` TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  INDEX `idx_channel_created` (`channel_id`, `created_at`),
  CONSTRAINT `fk_cmsg_channel` FOREIGN KEY (`channel_id`) REFERENCES `community_channels` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_cmsg_user`    FOREIGN KEY (`user_id`)    REFERENCES `users`              (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_cmsg_poll`    FOREIGN KEY (`poll_id`)    REFERENCES `polls`              (`id`) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- message_reactions
-- NOTE: run this CREATE if the table does not yet exist
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `message_reactions` (
  `message_id` VARCHAR(36) NOT NULL,
  `user_id`    VARCHAR(36) NOT NULL,
  `emoji`      VARCHAR(32) NOT NULL,
  `created_at` DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`message_id`, `user_id`),
  CONSTRAINT `fk_mr_message` FOREIGN KEY (`message_id`) REFERENCES `channel_messages` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_mr_user`    FOREIGN KEY (`user_id`)    REFERENCES `users`            (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- poll_votes
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `poll_votes` (
  `id`           VARCHAR(36) NOT NULL,
  `poll_id`      VARCHAR(36) NOT NULL,
  `user_id`      VARCHAR(36) NOT NULL,
  `option_index` INT         NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uq_poll_user` (`poll_id`, `user_id`),
  CONSTRAINT `fk_pv_poll` FOREIGN KEY (`poll_id`) REFERENCES `polls` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_pv_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- channel_settings
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `channel_settings` (
  `channel_id`              VARCHAR(36) NOT NULL,
  `disappearing_ttl_seconds` INT        NOT NULL DEFAULT 0,
  PRIMARY KEY (`channel_id`),
  CONSTRAINT `fk_cs_channel` FOREIGN KEY (`channel_id`) REFERENCES `community_channels` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- ---------------------------------------------------------------------------
-- messages  (legacy table from early migration — may still exist)
-- ---------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS `messages` (
  `id`         VARCHAR(36) NOT NULL,
  `stream_id`  VARCHAR(36) NOT NULL,
  `user_id`    VARCHAR(36) NOT NULL,
  `content`    TEXT        NOT NULL,
  `created_at` TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  CONSTRAINT `fk_msg_stream` FOREIGN KEY (`stream_id`) REFERENCES `streams` (`id`),
  CONSTRAINT `fk_msg_user`   FOREIGN KEY (`user_id`)   REFERENCES `users`   (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET FOREIGN_KEY_CHECKS = 1;
