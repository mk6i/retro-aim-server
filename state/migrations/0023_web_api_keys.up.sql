-- Create table for Web API authentication keys
CREATE TABLE IF NOT EXISTS web_api_keys
(
    dev_id          VARCHAR(255) PRIMARY KEY,
    dev_key         VARCHAR(255) UNIQUE NOT NULL,
    app_name        VARCHAR(255) NOT NULL,
    created_at      INTEGER NOT NULL,
    last_used       INTEGER,
    is_active       BOOLEAN DEFAULT 1,
    rate_limit      INTEGER DEFAULT 60,
    allowed_origins TEXT, -- JSON array of allowed CORS origins
    capabilities    TEXT  -- JSON array of enabled features/endpoints
);

-- Create indexes for efficient lookups
CREATE INDEX idx_web_api_keys_dev_key ON web_api_keys(dev_key);
CREATE INDEX idx_web_api_keys_is_active ON web_api_keys(is_active);

