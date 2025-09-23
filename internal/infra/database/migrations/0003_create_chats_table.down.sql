-- Drop trigger
DROP TRIGGER IF EXISTS "trigger_zpChats_updatedAt" ON "zpChats";

-- Drop function
DROP FUNCTION IF EXISTS "update_zpChats_updatedAt"();

-- Drop indexes
DROP INDEX IF EXISTS "idx_zpChats_sessionId";
DROP INDEX IF EXISTS "idx_zpChats_chatJid";
DROP INDEX IF EXISTS "idx_zpChats_phoneNumber";
DROP INDEX IF EXISTS "idx_zpChats_chatwootConversationId";
DROP INDEX IF EXISTS "idx_zpChats_lastMsgAt";
DROP INDEX IF EXISTS "idx_zpChats_session_chat_unique";

-- Drop table
DROP TABLE IF EXISTS "zpChats";
