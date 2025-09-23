-- Drop trigger
DROP TRIGGER IF EXISTS "trigger_zpCwMessages_updatedAt" ON "zpCwMessages";

-- Drop function
DROP FUNCTION IF EXISTS "update_zpCwMessages_updatedAt"();

-- Drop indexes
DROP INDEX IF EXISTS "idx_zpCwMessages_sessionId";
DROP INDEX IF EXISTS "idx_zpCwMessages_msgId";
DROP INDEX IF EXISTS "idx_zpCwMessages_chatwootMsgId";
DROP INDEX IF EXISTS "idx_zpCwMessages_chatwootConvId";
DROP INDEX IF EXISTS "idx_zpCwMessages_direction";
DROP INDEX IF EXISTS "idx_zpCwMessages_syncStatus";
DROP INDEX IF EXISTS "idx_zpCwMessages_sourceId";
DROP INDEX IF EXISTS "idx_zpCwMessages_msgId_unique";
DROP INDEX IF EXISTS "idx_zpCwMessages_chatwoot_unique";

-- Drop table
DROP TABLE IF EXISTS "zpCwMessages";
