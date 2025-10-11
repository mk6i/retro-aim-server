-- Revert lastWarnUpdate back to DATETIME type

ALTER TABLE users DROP COLUMN lastWarnUpdate;

ALTER TABLE users ADD COLUMN lastWarnUpdate DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00';

