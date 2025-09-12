-- Migration 0021: Web API Chat Rooms Support
-- This migration adds tables for Web API chat room functionality

-- Chat rooms table
CREATE TABLE IF NOT EXISTS web_chat_rooms (
    room_id VARCHAR(255) PRIMARY KEY,
    room_name VARCHAR(255) NOT NULL,
    description TEXT,
    room_type VARCHAR(50) DEFAULT 'userCreated',
    category_id VARCHAR(50),
    creator_screen_name VARCHAR(16) NOT NULL,
    created_at INTEGER NOT NULL,
    closed_at INTEGER,
    max_participants INTEGER DEFAULT 100,
    INDEX idx_web_chat_rooms_name (room_name),
    INDEX idx_web_chat_rooms_creator (creator_screen_name),
    INDEX idx_web_chat_rooms_created (created_at),
    INDEX idx_web_chat_rooms_closed (closed_at)
);

-- Chat sessions table (maps users to chat rooms)
CREATE TABLE IF NOT EXISTS web_chat_sessions (
    chat_sid VARCHAR(255) PRIMARY KEY,
    aimsid VARCHAR(255) NOT NULL,
    room_id VARCHAR(255) NOT NULL,
    screen_name VARCHAR(16) NOT NULL,
    instance_id INTEGER NOT NULL,
    joined_at INTEGER NOT NULL,
    left_at INTEGER,
    FOREIGN KEY (room_id) REFERENCES web_chat_rooms(room_id) ON DELETE CASCADE,
    INDEX idx_web_chat_sessions_aimsid (aimsid),
    INDEX idx_web_chat_sessions_room (room_id),
    INDEX idx_web_chat_sessions_user (screen_name),
    INDEX idx_web_chat_sessions_joined (joined_at)
);

-- Chat messages table
CREATE TABLE IF NOT EXISTS web_chat_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id VARCHAR(255) NOT NULL,
    screen_name VARCHAR(16) NOT NULL,
    message TEXT NOT NULL,
    whisper_target VARCHAR(16),
    timestamp INTEGER NOT NULL,
    FOREIGN KEY (room_id) REFERENCES web_chat_rooms(room_id) ON DELETE CASCADE,
    INDEX idx_web_chat_messages_room (room_id),
    INDEX idx_web_chat_messages_timestamp (timestamp),
    INDEX idx_web_chat_messages_user (screen_name)
);

-- Chat participants table (current participants in each room)
CREATE TABLE IF NOT EXISTS web_chat_participants (
    room_id VARCHAR(255) NOT NULL,
    screen_name VARCHAR(16) NOT NULL,
    chat_sid VARCHAR(255) NOT NULL,
    joined_at INTEGER NOT NULL,
    typing_status VARCHAR(20) DEFAULT 'none',
    typing_updated_at INTEGER,
    PRIMARY KEY (room_id, screen_name),
    FOREIGN KEY (room_id) REFERENCES web_chat_rooms(room_id) ON DELETE CASCADE,
    INDEX idx_web_chat_participants_room (room_id),
    INDEX idx_web_chat_participants_user (screen_name)
);






