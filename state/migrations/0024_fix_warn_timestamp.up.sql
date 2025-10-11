-- Change lastWarnUpdate from DATETIME to INTEGER (Unix timestamp)
-- The old data is corrupted (Go string representation), so we'll drop and recreate

ALTER TABLE users DROP COLUMN lastWarnUpdate;

ALTER TABLE users ADD COLUMN lastWarnUpdate INTEGER NOT NULL DEFAULT 0;

