-- Drop trigger
DROP TRIGGER IF EXISTS trigger_chatwoot_updated_at ON chatwoot;

-- Drop function
DROP FUNCTION IF EXISTS update_chatwoot_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_chatwoot_session_id;
DROP INDEX IF EXISTS idx_chatwoot_enabled;
DROP INDEX IF EXISTS idx_chatwoot_sync_status;

-- Drop table
DROP TABLE IF EXISTS chatwoot;
