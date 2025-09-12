-- Rollback Migration: 0020_vanity_urls
-- Description: Remove vanity URL tables
-- Date: 2024-12-28

DROP TABLE IF EXISTS vanity_url_redirects;
DROP TABLE IF EXISTS vanity_urls;
