-- =============================================================================
-- Chatwoot Integration Table - Rollback Migration
-- =============================================================================
-- This migration removes the chatwoot table and all related objects.

-- Drop triggers first
DROP TRIGGER IF EXISTS trigger_chatwoot_updated_at ON chatwoot;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_chatwoot_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_chatwoot_session_id;
DROP INDEX IF EXISTS idx_chatwoot_enabled;
DROP INDEX IF EXISTS idx_chatwoot_account;
DROP INDEX IF EXISTS idx_chatwoot_sync_status;
DROP INDEX IF EXISTS idx_chatwoot_inbox_id;

-- Drop the table (this will also drop all constraints)
DROP TABLE IF EXISTS chatwoot;
