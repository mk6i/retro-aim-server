-- Create table for storing WebAPI to OSCAR bridge sessions
-- This table maps WebAPI sessions to OSCAR authentication cookies
-- and tracks connection details for bridged sessions

CREATE TABLE IF NOT EXISTS oscar_bridge_sessions (
    -- WebAPI session identifier (aimsid)
    web_session_id VARCHAR(64) PRIMARY KEY,
    
    -- OSCAR authentication cookie (hex encoded)
    oscar_cookie BLOB NOT NULL,
    
    -- BOS server connection details
    bos_host VARCHAR(255) NOT NULL,
    bos_port INTEGER NOT NULL,
    use_ssl BOOLEAN DEFAULT FALSE,
    
    -- Screen name associated with the session
    screen_name VARCHAR(97) NOT NULL,
    
    -- Session metadata
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Optional: Track client info
    client_name VARCHAR(255),
    client_version VARCHAR(50)
);

-- Create index for quick lookups by screen name
CREATE INDEX idx_oscar_bridge_screen_name ON oscar_bridge_sessions(screen_name);

-- Create index for cleanup of old sessions
CREATE INDEX idx_oscar_bridge_last_accessed ON oscar_bridge_sessions(last_accessed);

-- Optional: Create a trigger to auto-update last_accessed on SELECT/UPDATE
-- (SQLite doesn't support this directly, would need application logic)
