-- Rollback: restore webhook fields to zpSessions and drop zpWebhooks table
DO $$
BEGIN
    -- Add webhook columns back to zpSessions if they don't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'zpSessions'
        AND column_name = 'webhookUrl'
        AND table_schema = current_schema()
    ) THEN
        ALTER TABLE "zpSessions" ADD COLUMN "webhookUrl" VARCHAR(500);
        ALTER TABLE "zpSessions" ADD COLUMN "webhookEvents" VARCHAR(500) DEFAULT 'message';

        -- Migrate data back from zpWebhooks to zpSessions if zpWebhooks exists
        IF EXISTS (
            SELECT 1 FROM information_schema.tables
            WHERE table_name = 'zpWebhooks'
            AND table_schema = current_schema()
        ) THEN
            UPDATE "zpSessions"
            SET "webhookUrl" = w.url,
                "webhookEvents" = array_to_string(w.events, ',')
            FROM "zpWebhooks" w
            WHERE "zpSessions".id = w."sessionId";
        END IF;

        RAISE NOTICE 'Restored webhook columns to zpSessions table';
    END IF;
END $$;

-- Drop trigger
DROP TRIGGER IF EXISTS "trigger_zpWebhooks_updatedAt" ON "zpWebhooks";

-- Drop function
DROP FUNCTION IF EXISTS "update_zpWebhooks_updatedAt"();

-- Drop indexes
DROP INDEX IF EXISTS "idx_zpWebhooks_sessionId";
DROP INDEX IF EXISTS "idx_zpWebhooks_isActive";
DROP INDEX IF EXISTS "idx_zpWebhooks_sessionId_unique";

-- Drop table
DROP TABLE IF EXISTS "zpWebhooks";
