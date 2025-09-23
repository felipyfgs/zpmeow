-- Drop trigger
DROP TRIGGER IF EXISTS "trigger_zpChatwoot_updatedAt" ON "zpChatwoot";

-- Drop function
DROP FUNCTION IF EXISTS "update_zpChatwoot_updatedAt"();

-- Drop indexes
DROP INDEX IF EXISTS "idx_zpChatwoot_sessionId";
DROP INDEX IF EXISTS "idx_zpChatwoot_enabled";
DROP INDEX IF EXISTS "idx_zpChatwoot_syncStatus";

-- Drop table
DROP TABLE IF EXISTS "zpChatwoot";
