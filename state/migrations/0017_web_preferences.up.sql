-- Create table for Web API user preferences
CREATE TABLE IF NOT EXISTS web_preferences
(
    screen_name         VARCHAR(16) PRIMARY KEY,
    preferences         TEXT,        -- JSON object of preference key-value pairs
    created_at          INTEGER NOT NULL,
    updated_at          INTEGER NOT NULL
);

-- Create index for efficient lookups
CREATE INDEX idx_web_preferences_screen_name ON web_preferences(screen_name);

-- Ensure buddyListMode table exists (for PD mode storage)
-- This should already exist from migration 0010, but we'll add IF NOT EXISTS for safety
CREATE TABLE IF NOT EXISTS buddyListMode
(
    screenName       VARCHAR(16),
    clientSidePDMode INTEGER DEFAULT 0,
    useFeedbag       BOOLEAN DEFAULT false,
    PRIMARY KEY (screenName)
);

-- Ensure clientSideBuddyList table exists (for permit/deny lists)
-- This should already exist from migration 0010, but we'll add IF NOT EXISTS for safety
CREATE TABLE IF NOT EXISTS clientSideBuddyList
(
    me       VARCHAR(16),
    them     VARCHAR(16),
    isBuddy  BOOLEAN DEFAULT false,
    isPermit BOOLEAN DEFAULT false,
    isDeny   BOOLEAN DEFAULT false,
    PRIMARY KEY (me, them)
);
