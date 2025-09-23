-- Drop trigger
DROP TRIGGER IF EXISTS trigger_chats_updated_at ON chats;

-- Drop function
DROP FUNCTION IF EXISTS update_chats_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_chats_session_id;
DROP INDEX IF EXISTS idx_chats_chat_jid;
DROP INDEX IF EXISTS idx_chats_phone_number;
DROP INDEX IF EXISTS idx_chats_chatwoot_conversation_id;
DROP INDEX IF EXISTS idx_chats_last_message_at;
DROP INDEX IF EXISTS idx_chats_session_chat_unique;

-- Drop table
DROP TABLE IF EXISTS chats;
