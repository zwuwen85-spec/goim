-- ============================================
-- GoIM Chat Database Schema
-- QQ-like chat software database tables
-- ============================================

-- 1. Users Table
CREATE TABLE IF NOT EXISTS `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'User ID (corresponds to goim mid)',
  `username` VARCHAR(50) NOT NULL COMMENT 'Username (unique)',
  `password_hash` VARCHAR(255) NOT NULL COMMENT 'Password hash (bcrypt)',
  `nickname` VARCHAR(100) NOT NULL COMMENT 'Display nickname',
  `avatar_url` VARCHAR(500) DEFAULT NULL COMMENT 'Avatar image URL',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'Status: 1=online, 2=offline, 3=busy',
  `signature` VARCHAR(200) DEFAULT NULL COMMENT 'Personal signature/bio',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Account creation time',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  KEY `idx_nickname` (`nickname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User accounts';

-- 2. User Tokens Table (for goim authentication)
CREATE TABLE IF NOT EXISTS `user_tokens` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
  `token` VARCHAR(128) NOT NULL COMMENT 'Access token (JWT)',
  `device_id` VARCHAR(100) DEFAULT NULL COMMENT 'Device identifier',
  `platform` VARCHAR(20) NOT NULL COMMENT 'Platform: web/android/ios',
  `expires_at` TIMESTAMP NOT NULL COMMENT 'Token expiration time',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Token creation time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_token` (`token`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_expires_at` (`expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User authentication tokens';

-- 3. Friendships Table
CREATE TABLE IF NOT EXISTS `friendships` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
  `friend_id` BIGINT UNSIGNED NOT NULL COMMENT 'Friend ID',
  `remark` VARCHAR(100) DEFAULT NULL COMMENT 'Friend remark/alias',
  `group_name` VARCHAR(50) DEFAULT '默认分组' COMMENT 'Friend group name',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'Status: 1=normal, 2=deleted, 3=blocked',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Friend relationship created',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_friend` (`user_id`, `friend_id`),
  KEY `idx_friend_id` (`friend_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Friend relationships';

-- 4. Friend Requests Table
CREATE TABLE IF NOT EXISTS `friend_requests` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `from_user_id` BIGINT UNSIGNED NOT NULL COMMENT 'Applicant user ID',
  `to_user_id` BIGINT UNSIGNED NOT NULL COMMENT 'Target user ID',
  `message` VARCHAR(200) DEFAULT NULL COMMENT 'Application message',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'Status: 1=pending, 2=accepted, 3=rejected',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Request created',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update time',
  PRIMARY KEY (`id`),
  KEY `idx_from_user` (`from_user_id`),
  KEY `idx_to_user` (`to_user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Friend requests';

-- 5. Groups Table
CREATE TABLE IF NOT EXISTS `groups` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'Group ID',
  `group_no` VARCHAR(20) NOT NULL COMMENT 'Group number (unique)',
  `name` VARCHAR(100) NOT NULL COMMENT 'Group name',
  `avatar_url` VARCHAR(500) DEFAULT NULL COMMENT 'Group avatar URL',
  `owner_id` BIGINT UNSIGNED NOT NULL COMMENT 'Group owner ID',
  `max_members` INT NOT NULL DEFAULT 500 COMMENT 'Maximum members',
  `join_type` TINYINT NOT NULL DEFAULT 1 COMMENT 'Join type: 1=open, 2=need approval, 3=closed',
  `mute_all` TINYINT NOT NULL DEFAULT 0 COMMENT 'Mute all members: 0=no, 1=yes',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Group created',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_no` (`group_no`),
  KEY `idx_name` (`name`),
  KEY `idx_owner_id` (`owner_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Chat groups';

-- 6. Group Members Table
CREATE TABLE IF NOT EXISTS `group_members` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` BIGINT UNSIGNED NOT NULL COMMENT 'Group ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
  `role` TINYINT NOT NULL DEFAULT 1 COMMENT 'Role: 1=member, 2=admin, 3=owner',
  `nickname` VARCHAR(100) DEFAULT NULL COMMENT 'Nickname in group',
  `mute_until` TIMESTAMP NULL DEFAULT NULL COMMENT 'Mute expiration time',
  `joined_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Join time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_user` (`group_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Group members';

-- 7. Group Join Requests Table
CREATE TABLE IF NOT EXISTS `group_join_requests` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `group_id` BIGINT UNSIGNED NOT NULL COMMENT 'Group ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'Applicant user ID',
  `message` VARCHAR(200) DEFAULT NULL COMMENT 'Application message',
  `status` TINYINT NOT NULL DEFAULT 1 COMMENT 'Status: 1=pending, 2=accepted, 3=rejected',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Request created',
  PRIMARY KEY (`id`),
  KEY `idx_group_id` (`group_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Group join requests';

-- 8. Messages Table (single chat + group chat + AI chat)
CREATE TABLE IF NOT EXISTS `messages` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `msg_id` VARCHAR(40) NOT NULL COMMENT 'Unique message ID (UUID)',
  `from_user_id` BIGINT UNSIGNED NOT NULL COMMENT 'Sender user ID',
  `conversation_id` BIGINT UNSIGNED NOT NULL COMMENT 'Conversation ID (single: user_pair_id, group: group_id, AI: bot_id)',
  `conversation_type` TINYINT NOT NULL COMMENT 'Conversation type: 1=single, 2=group, 3=AI',
  `msg_type` TINYINT NOT NULL COMMENT 'Message type: 1=text, 2=image, 3=voice, 4=video, 5=file, 6=system',
  `content` TEXT NOT NULL COMMENT 'Message content (JSON format)',
  `seq` BIGINT NOT NULL COMMENT 'Message sequence number (for ordering and sync)',
  `is_deleted` TINYINT NOT NULL DEFAULT 0 COMMENT 'Is deleted: 0=no, 1=yes',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Message sent time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_msg_id` (`msg_id`),
  KEY `idx_conversation` (`conversation_id`, `conversation_type`, `created_at`),
  KEY `idx_seq` (`conversation_id`, `conversation_type`, `seq`),
  KEY `idx_from_user` (`from_user_id`, `created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Chat messages';

-- 9. Message Read Status Table
CREATE TABLE IF NOT EXISTS `message_read_status` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `msg_id` BIGINT UNSIGNED NOT NULL COMMENT 'Message ID (internal auto-increment)',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'User ID who read the message',
  `read_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Read time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_msg_user` (`msg_id`, `user_id`),
  KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Message read status';

-- 10. Conversations Table (recent chat list)
CREATE TABLE IF NOT EXISTS `conversations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
  `target_id` BIGINT UNSIGNED NOT NULL COMMENT 'Target ID (single: other_user_id, group: group_id, AI: bot_id)',
  `conversation_type` TINYINT NOT NULL COMMENT 'Conversation type: 1=single, 2=group, 3=AI',
  `unread_count` INT NOT NULL DEFAULT 0 COMMENT 'Unread message count',
  `last_msg_id` BIGINT UNSIGNED DEFAULT NULL COMMENT 'Last message ID',
  `last_msg_content` TEXT DEFAULT NULL COMMENT 'Last message content preview',
  `last_msg_time` TIMESTAMP NULL DEFAULT NULL COMMENT 'Last message time',
  `is_pinned` TINYINT NOT NULL DEFAULT 0 COMMENT 'Is pinned: 0=no, 1=yes',
  `is_muted` TINYINT NOT NULL DEFAULT 0 COMMENT 'Is muted: 0=no, 1=yes',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_target` (`user_id`, `target_id`, `conversation_type`),
  KEY `idx_user_unread` (`user_id`, `unread_count`),
  KEY `idx_updated` (`user_id`, `updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='User conversation list';

-- 11. AI Bots Table
CREATE TABLE IF NOT EXISTS `ai_bots` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `bot_id` BIGINT UNSIGNED NOT NULL COMMENT 'Bot user ID (corresponds to users.id)',
  `name` VARCHAR(100) NOT NULL COMMENT 'Bot display name',
  `avatar_url` VARCHAR(500) DEFAULT NULL COMMENT 'Bot avatar URL',
  `personality` TEXT NOT NULL COMMENT 'Personality config (JSON)',
  `prompt_template` TEXT DEFAULT NULL COMMENT 'System prompt template',
  `model_name` VARCHAR(50) NOT NULL DEFAULT 'gpt-3.5-turbo' COMMENT 'AI model name',
  `max_tokens` INT NOT NULL DEFAULT 500 COMMENT 'Max response tokens',
  `temperature` DECIMAL(3,2) NOT NULL DEFAULT 0.70 COMMENT 'Temperature parameter (0.0-1.0)',
  `is_active` TINYINT NOT NULL DEFAULT 1 COMMENT 'Is active: 0=disabled, 1=enabled',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Bot created',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last update time',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_bot_id` (`bot_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI chatbot configurations';

-- 12. AI Conversations Table (context history)
CREATE TABLE IF NOT EXISTS `ai_conversations` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `bot_id` BIGINT UNSIGNED NOT NULL COMMENT 'Bot ID',
  `user_id` BIGINT UNSIGNED NOT NULL COMMENT 'User ID',
  `messages` JSON NOT NULL COMMENT 'Conversation history messages',
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'First interaction',
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Last interaction',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_bot_user` (`bot_id`, `user_id`),
  KEY `idx_updated` (`updated_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='AI chat history context';
