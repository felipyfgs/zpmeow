-- =============================================================================
-- Sessions Table - Rollback Migration
-- =============================================================================
-- This migration completely removes the sessions table and all related objects

-- Drop indexes first (if they exist)
DROP INDEX IF EXISTS idx_sessions_apikey;

-- Drop constraints (if they exist)
ALTER TABLE IF EXISTS sessions DROP CONSTRAINT IF EXISTS unique_session_name;

-- Drop the table completely
DROP TABLE IF EXISTS sessions;

-- Note: This will permanently delete all session data!
