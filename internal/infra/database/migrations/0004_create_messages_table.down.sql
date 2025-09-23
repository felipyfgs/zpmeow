-- Drop trigger
DROP TRIGGER IF EXISTS "trigger_zpMessages_updatedAt" ON "zpMessages";

-- Drop function
DROP FUNCTION IF EXISTS "update_zpMessages_updatedAt"();

-- Drop indexes
DROP INDEX IF EXISTS "idx_zpMessages_chatId";
DROP INDEX IF EXISTS "idx_zpMessages_sessionId";
DROP INDEX IF EXISTS "idx_zpMessages_msgId";
DROP INDEX IF EXISTS "idx_zpMessages_senderJid";
DROP INDEX IF EXISTS "idx_zpMessages_timestamp";
DROP INDEX IF EXISTS "idx_zpMessages_status";
DROP INDEX IF EXISTS "idx_zpMessages_msgType";
DROP INDEX IF EXISTS "idx_zpMessages_isFromMe";
DROP INDEX IF EXISTS "idx_zpMessages_quotedMsgId";
DROP INDEX IF EXISTS "idx_zpMessages_session_msgId_unique";

-- Drop table
DROP TABLE IF EXISTS "zpMessages";
