-- Drop trigger
DROP TRIGGER IF EXISTS "trigger_zpSessions_updatedAt" ON "zpSessions";

-- Drop function
DROP FUNCTION IF EXISTS "update_zpSessions_updatedAt"();

-- Drop indexes
DROP INDEX IF EXISTS "idx_zpSessions_name";
DROP INDEX IF EXISTS "idx_zpSessions_apiKey";
DROP INDEX IF EXISTS "idx_zpSessions_status";

-- Drop table
DROP TABLE IF EXISTS "zpSessions";
