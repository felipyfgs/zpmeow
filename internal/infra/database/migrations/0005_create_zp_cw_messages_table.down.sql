-- Drop trigger
DROP TRIGGER IF EXISTS trigger_zp_cw_messages_updated_at ON zp_cw_messages;

-- Drop function
DROP FUNCTION IF EXISTS update_zp_cw_messages_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_zp_cw_messages_session_id;
DROP INDEX IF EXISTS idx_zp_cw_messages_zpmeow_message_id;
DROP INDEX IF EXISTS idx_zp_cw_messages_chatwoot_message_id;
DROP INDEX IF EXISTS idx_zp_cw_messages_chatwoot_conversation_id;
DROP INDEX IF EXISTS idx_zp_cw_messages_direction;
DROP INDEX IF EXISTS idx_zp_cw_messages_sync_status;
DROP INDEX IF EXISTS idx_zp_cw_messages_chatwoot_source_id;
DROP INDEX IF EXISTS idx_zp_cw_messages_chatwoot_echo_id;
DROP INDEX IF EXISTS idx_zp_cw_messages_zpmeow_unique;
DROP INDEX IF EXISTS idx_zp_cw_messages_chatwoot_unique;

-- Drop table
DROP TABLE IF EXISTS zp_cw_messages;
