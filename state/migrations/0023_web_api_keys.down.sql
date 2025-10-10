-- Drop indexes
DROP INDEX IF EXISTS idx_web_api_keys_is_active;
DROP INDEX IF EXISTS idx_web_api_keys_dev_key;

-- Drop table
DROP TABLE IF EXISTS web_api_keys;

