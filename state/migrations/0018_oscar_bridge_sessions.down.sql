-- Drop the OSCAR bridge sessions table and its indexes

DROP INDEX IF EXISTS idx_oscar_bridge_screen_name;
DROP INDEX IF EXISTS idx_oscar_bridge_last_accessed;
DROP TABLE IF EXISTS oscar_bridge_sessions;
