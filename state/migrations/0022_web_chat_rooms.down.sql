-- Rollback migration 0021: Web API Chat Rooms Support
DROP TABLE IF EXISTS web_chat_messages;
DROP TABLE IF EXISTS web_chat_participants;
DROP TABLE IF EXISTS web_chat_sessions;
DROP TABLE IF EXISTS web_chat_rooms;











