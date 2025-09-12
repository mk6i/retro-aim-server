-- Migration: 0020_vanity_urls
-- Description: Create tables for vanity URL management
-- Date: 2024-12-28

-- Create vanity URLs table for custom user URLs
CREATE TABLE IF NOT EXISTS vanity_urls (
    screen_name VARCHAR(16) PRIMARY KEY,
    vanity_url VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(100),
    bio TEXT,
    location VARCHAR(100),
    website VARCHAR(255),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    click_count INTEGER DEFAULT 0,
    last_accessed INTEGER
);

-- Create indexes for efficient lookups
CREATE INDEX idx_vanity_urls_url ON vanity_urls(vanity_url);
CREATE INDEX idx_vanity_urls_active ON vanity_urls(is_active);
CREATE INDEX idx_vanity_urls_created ON vanity_urls(created_at);

-- Create vanity URL redirects table for tracking and analytics
CREATE TABLE IF NOT EXISTS vanity_url_redirects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    vanity_url VARCHAR(255) NOT NULL,
    accessed_at INTEGER NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referer TEXT,
    FOREIGN KEY (vanity_url) REFERENCES vanity_urls(vanity_url) ON DELETE CASCADE
);

-- Create index for redirect analytics
CREATE INDEX idx_vanity_redirects_url ON vanity_url_redirects(vanity_url);
CREATE INDEX idx_vanity_redirects_time ON vanity_url_redirects(accessed_at);
