-- Rollback Migration: 0019_buddy_feeds
-- Description: Remove buddy feed tables
-- Date: 2024-12-28

DROP TABLE IF EXISTS buddy_feed_subscriptions;
DROP TABLE IF EXISTS buddy_feed_items;
DROP TABLE IF EXISTS buddy_feeds;
