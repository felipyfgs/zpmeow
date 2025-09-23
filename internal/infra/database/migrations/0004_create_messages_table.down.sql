-- Drop trigger
DROP TRIGGER IF EXISTS trigger_messages_updated_at ON messages;

-- Drop function
DROP FUNCTION IF EXISTS update_messages_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_messages_chat_id;
DROP INDEX IF EXISTS idx_messages_session_id;
DROP INDEX IF EXISTS idx_messages_whatsapp_message_id;
DROP INDEX IF EXISTS idx_messages_sender_jid;
DROP INDEX IF EXISTS idx_messages_timestamp;
DROP INDEX IF EXISTS idx_messages_status;
DROP INDEX IF EXISTS idx_messages_message_type;
DROP INDEX IF EXISTS idx_messages_is_from_me;
DROP INDEX IF EXISTS idx_messages_quoted_message_id;
DROP INDEX IF EXISTS idx_messages_session_whatsapp_unique;

-- Drop table
DROP TABLE IF EXISTS messages;
