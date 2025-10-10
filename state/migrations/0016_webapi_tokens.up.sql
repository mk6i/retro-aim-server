-- Create table for storing Web API authentication tokens
CREATE TABLE IF NOT EXISTS webapi_tokens (
    token TEXT PRIMARY KEY,
    screen_name TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Index for cleaning up expired tokens
CREATE INDEX IF NOT EXISTS idx_webapi_tokens_expires_at ON webapi_tokens(expires_at);

-- Index for looking up tokens by screen name
CREATE INDEX IF NOT EXISTS idx_webapi_tokens_screen_name ON webapi_tokens(screen_name);
