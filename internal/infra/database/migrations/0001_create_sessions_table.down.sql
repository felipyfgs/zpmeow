-- Drop trigger
DROP TRIGGER IF EXISTS trigger_sessions_updated_at ON sessions;

-- Drop function
DROP FUNCTION IF EXISTS update_sessions_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_sessions_name;
DROP INDEX IF EXISTS idx_sessions_apikey;
DROP INDEX IF EXISTS idx_sessions_status;

-- Drop table
DROP TABLE IF EXISTS sessions;
