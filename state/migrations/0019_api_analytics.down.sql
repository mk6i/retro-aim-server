-- Rollback Migration: 0018_api_analytics
-- Description: Remove Web API usage analytics tables
-- Date: 2024-12-28

DROP TABLE IF EXISTS api_quotas;
DROP TABLE IF EXISTS api_usage_stats;
DROP TABLE IF EXISTS api_usage_logs;
