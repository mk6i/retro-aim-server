-- Migration: 0018_api_analytics
-- Description: Create tables for Web API usage analytics and tracking
-- Date: 2024-12-28

-- Create API usage logs table for detailed request tracking
CREATE TABLE IF NOT EXISTS api_usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dev_id VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    timestamp INTEGER NOT NULL,
    response_time_ms INTEGER,
    status_code INTEGER,
    ip_address VARCHAR(45),
    user_agent TEXT,
    screen_name VARCHAR(16), -- User making the request (if authenticated)
    error_message TEXT, -- Store error details if request failed
    request_size INTEGER, -- Size of request in bytes
    response_size INTEGER -- Size of response in bytes
);

-- Create indexes for efficient querying
CREATE INDEX idx_usage_dev_id ON api_usage_logs(dev_id);
CREATE INDEX idx_usage_timestamp ON api_usage_logs(timestamp);
CREATE INDEX idx_usage_endpoint ON api_usage_logs(endpoint);
CREATE INDEX idx_usage_status ON api_usage_logs(status_code);
CREATE INDEX idx_usage_screen_name ON api_usage_logs(screen_name);

-- Create aggregated statistics table for performance
CREATE TABLE IF NOT EXISTS api_usage_stats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dev_id VARCHAR(255) NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    period_type VARCHAR(10) NOT NULL, -- 'hour', 'day', 'month'
    period_start INTEGER NOT NULL,
    request_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    total_response_time_ms INTEGER DEFAULT 0,
    avg_response_time_ms INTEGER DEFAULT 0,
    total_request_bytes INTEGER DEFAULT 0,
    total_response_bytes INTEGER DEFAULT 0,
    unique_users INTEGER DEFAULT 0,
    UNIQUE(dev_id, endpoint, period_type, period_start)
);

-- Create indexes for aggregated stats
CREATE INDEX idx_stats_dev_id ON api_usage_stats(dev_id);
CREATE INDEX idx_stats_period ON api_usage_stats(period_type, period_start);
CREATE INDEX idx_stats_endpoint ON api_usage_stats(endpoint);

-- Create table for tracking API key quotas and limits
CREATE TABLE IF NOT EXISTS api_quotas (
    dev_id VARCHAR(255) PRIMARY KEY,
    daily_limit INTEGER DEFAULT 10000,
    monthly_limit INTEGER DEFAULT 300000,
    daily_used INTEGER DEFAULT 0,
    monthly_used INTEGER DEFAULT 0,
    last_reset_daily INTEGER NOT NULL,
    last_reset_monthly INTEGER NOT NULL,
    overage_allowed BOOLEAN DEFAULT FALSE
);
