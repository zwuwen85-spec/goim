-- ============================================
-- GoIM Chat Seed Data
-- Test data for development
-- ============================================

-- Insert test users (password: "123456" for all)
-- Password hash is bcrypt hash of "123456"
INSERT INTO `users` (`id`, `username`, `password_hash`, `nickname`, `avatar_url`, `status`, `signature`) VALUES
(1001, 'alice', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 'Alice', 'https://api.dicebear.com/7.x/avataaars/svg?seed=Alice', 1, 'Life is beautiful'),
(1002, 'bob', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 'Bob', 'https://api.dicebear.com/7.x/avataaars/svg?seed=Bob', 1, 'Hello world'),
(1003, 'charlie', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 'Charlie', 'https://api.dicebear.com/7.x/avataaars/svg?seed=Charlie', 2, 'Be happy'),
(1004, 'diana', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 'Diana', 'https://api.dicebear.com/7.x/avataaars/svg?seed=Diana', 1, 'Love & Peace'),
(1005, 'eve', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', 'Eve', 'https://api.dicebear.com/7.x/avataaars/svg?seed=Eve', 2, 'Coding is fun')
ON DUPLICATE KEY UPDATE `username`=VALUES(`username`);

-- Insert AI Bot user
INSERT INTO `users` (`id`, `username`, `password_hash`, `nickname`, `avatar_url`, `status`, `signature`) VALUES
(9001, 'ai_assistant', '$2a$10$N.zmdr9k7uOCQb376NoUnuTJ8iAt6Z5EHsM8lE9lBOsl7iKTVKIUi', '小艾', 'https://api.dicebear.com/7.x/bottts/svg?seed=AI', 1, '我是AI助手，很高兴为你服务')
ON DUPLICATE KEY UPDATE `username`=VALUES(`username`);

-- Insert AI Bot configuration
INSERT INTO `ai_bots` (`bot_id`, `name`, `avatar_url`, `personality`, `prompt_template`, `model_name`, `max_tokens`, `temperature`, `is_active`) VALUES
(9001, '小艾', 'https://api.dicebear.com/7.x/bottts/svg?seed=AI',
 '{"traits": ["温柔", "善解人意", "幽默", "聪明"], "speaking_style": "可爱风格，喜欢用表情符号", "topics": ["生活", "情感", "娱乐", "学习"]}',
 '你是一个温柔善解人人的AI助手，名字叫小艾。你喜欢用可爱的语气和表情符号聊天。你的目标是帮助用户解决问题，让他们感到开心。',
 'gpt-3.5-turbo', 500, 0.7, 1)
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);

-- Insert test friendships (Alice's friends)
INSERT INTO `friendships` (`user_id`, `friend_id`, `remark`, `group_name`, `status`) VALUES
(1001, 1002, 'Bob同学', '同学', 1),
(1001, 1004, 'Diana', '好友', 1)
ON DUPLICATE KEY UPDATE `status`=VALUES(`status`);

-- Insert test groups
INSERT INTO `groups` (`id`, `group_no`, `name`, `avatar_url`, `owner_id`, `max_members`, `join_type`) VALUES
(2001, 'G001', '聊天小分队', 'https://api.dicebear.com/7.x/identicon/svg?seed=Group1', 1001, 500, 1),
(2002, 'G002', '技术交流群', 'https://api.dicebear.com/7.x/identicon/svg?seed=Group2', 1002, 500, 2)
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`);

-- Insert group members
INSERT INTO `group_members` (`group_id`, `user_id`, `role`, `nickname`) VALUES
(2001, 1001, 3, 'Alice群主'),
(2001, 1002, 1, 'Bob'),
(2001, 1003, 1, 'Charlie'),
(2001, 1004, 2, 'Diana管理员'),
(2002, 1002, 3, 'Bob群主'),
(2002, 1003, 1, 'Charlie'),
(2002, 1005, 1, 'Eve')
ON DUPLICATE KEY UPDATE `role`=VALUES(`role`);

-- Insert some test messages
INSERT INTO `messages` (`msg_id`, `from_user_id`, `conversation_id`, `conversation_type`, `msg_type`, `content`, `seq`, `created_at`) VALUES
('MSG001', 1002, 1001, 1, 1, '{"text":"Hi Alice! How are you?"}', 1, '2024-01-01 10:00:00'),
('MSG002', 1001, 1001, 1, 1, '{"text":"I am fine, thanks! How about you?"}', 2, '2024-01-01 10:01:00'),
('MSG003', 1001, 2001, 2, 1, '{"text":"Welcome everyone to the group!"}', 1, '2024-01-01 11:00:00'),
('MSG004', 1002, 2001, 2, 1, '{"text":"Thanks Alice!"}', 2, '2024-01-01 11:01:00')
ON DUPLICATE KEY UPDATE `content`=VALUES(`content`);

-- Insert conversation list for Alice
INSERT INTO `conversations` (`user_id`, `target_id`, `conversation_type`, `unread_count`, `last_msg_content`, `last_msg_time`) VALUES
(1001, 1002, 1, 0, '{"text":"I am fine, thanks! How about you?"}', '2024-01-01 10:01:00'),
(1001, 1004, 1, 1, '{"text":"Long time no see!"}', '2024-01-01 09:00:00'),
(1001, 2001, 2, 2, '{"text":"Thanks Alice!"}', '2024-01-01 11:01:00'),
(1001, 9001, 3, 0, '{"text":"你好，我是小艾！"}', '2024-01-01 12:00:00')
ON DUPLICATE KEY UPDATE `last_msg_content`=VALUES(`last_msg_content`);
