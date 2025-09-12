-- Migration: 0019_buddy_feeds
-- Description: Create tables for buddy feed functionality
-- Date: 2024-12-28

-- Create buddy feeds table for storing user feed configurations
CREATE TABLE IF NOT EXISTS buddy_feeds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    screen_name VARCHAR(16) NOT NULL,
    feed_type VARCHAR(50) NOT NULL, -- 'rss', 'atom', 'status', 'blog', 'social'
    title TEXT,
    description TEXT,
    link TEXT,
    published_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE
);

-- Create indexes for efficient querying
CREATE INDEX idx_buddy_feeds_screen_name ON buddy_feeds(screen_name);
CREATE INDEX idx_buddy_feeds_published ON buddy_feeds(published_at);
CREATE INDEX idx_buddy_feeds_type ON buddy_feeds(feed_type);
CREATE INDEX idx_buddy_feeds_active ON buddy_feeds(is_active);

-- Create buddy feed items table for individual feed entries
CREATE TABLE IF NOT EXISTS buddy_feed_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    feed_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    link TEXT,
    guid TEXT,
    author VARCHAR(16), -- Screen name of the author
    categories TEXT, -- JSON array of categories
    published_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (feed_id) REFERENCES buddy_feeds(id) ON DELETE CASCADE
);

-- Create indexes for feed items
CREATE INDEX idx_feed_items_feed_id ON buddy_feed_items(feed_id);
CREATE INDEX idx_feed_items_published ON buddy_feed_items(published_at);
CREATE INDEX idx_feed_items_guid ON buddy_feed_items(guid);

-- Create buddy feed subscriptions table for tracking who follows which feeds
CREATE TABLE IF NOT EXISTS buddy_feed_subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    subscriber_screen_name VARCHAR(16) NOT NULL,
    feed_id INTEGER NOT NULL,
    subscribed_at INTEGER NOT NULL,
    last_checked_at INTEGER,
    FOREIGN KEY (feed_id) REFERENCES buddy_feeds(id) ON DELETE CASCADE,
    UNIQUE(subscriber_screen_name, feed_id)
);

-- Create indexes for subscriptions
CREATE INDEX idx_feed_subs_subscriber ON buddy_feed_subscriptions(subscriber_screen_name);
CREATE INDEX idx_feed_subs_feed_id ON buddy_feed_subscriptions(feed_id);
