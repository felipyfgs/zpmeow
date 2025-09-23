-- Create zpWebhooks table with camelCase columns
CREATE TABLE IF NOT EXISTS "zpWebhooks" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"(id) ON DELETE CASCADE,
    url VARCHAR(500) NOT NULL,
    events TEXT[] DEFAULT '{}',
    "isActive" BOOLEAN DEFAULT TRUE,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS "idx_zpWebhooks_sessionId" ON "zpWebhooks"("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zpWebhooks_isActive" ON "zpWebhooks"("isActive");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zpWebhooks_sessionId_unique" ON "zpWebhooks"("sessionId");

-- Remove webhook fields from zpSessions table if they exist (for compatibility)
DO $$
BEGIN
    -- Check if webhookUrl column exists and remove it
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'zpSessions'
        AND column_name = 'webhookUrl'
        AND table_schema = current_schema()
    ) THEN
        -- Migrate existing webhook data to the new zpWebhooks table first
        INSERT INTO "zpWebhooks" ("sessionId", url, events, "isActive")
        SELECT
            id,
            "webhookUrl",
            CASE
                WHEN "webhookEvents" IS NOT NULL AND "webhookEvents" != ''
                THEN string_to_array("webhookEvents", ',')
                ELSE ARRAY['message']::TEXT[]
            END,
            CASE
                WHEN "webhookUrl" IS NOT NULL AND "webhookUrl" != ''
                THEN TRUE
                ELSE FALSE
            END
        FROM "zpSessions"
        WHERE "webhookUrl" IS NOT NULL AND "webhookUrl" != ''
        ON CONFLICT ("sessionId") DO NOTHING;

        -- Remove the webhook columns from zpSessions
        ALTER TABLE "zpSessions" DROP COLUMN IF EXISTS "webhookUrl";
        ALTER TABLE "zpSessions" DROP COLUMN IF EXISTS "webhookEvents";

        RAISE NOTICE 'Migrated existing webhook data from zpSessions to zpWebhooks table';
    ELSE
        RAISE NOTICE 'No webhook columns found in zpSessions - clean installation';
    END IF;
END $$;

-- Create trigger function for updatedAt
CREATE OR REPLACE FUNCTION "update_zpWebhooks_updatedAt"()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER "trigger_zpWebhooks_updatedAt"
    BEFORE UPDATE ON "zpWebhooks"
    FOR EACH ROW
    EXECUTE FUNCTION "update_zpWebhooks_updatedAt"();

-- Comments
COMMENT ON TABLE "zpWebhooks" IS 'Webhook configurations for sessions (camelCase)';
COMMENT ON COLUMN "zpWebhooks".id IS 'Unique webhook configuration identifier (UUID)';
COMMENT ON COLUMN "zpWebhooks"."sessionId" IS 'Reference to the WhatsApp session (unique per session)';
COMMENT ON COLUMN "zpWebhooks".url IS 'Webhook URL endpoint';
COMMENT ON COLUMN "zpWebhooks".events IS 'Array of subscribed event types';
COMMENT ON COLUMN "zpWebhooks"."isActive" IS 'Whether webhook is active and should receive events';
COMMENT ON COLUMN "zpWebhooks"."createdAt" IS 'Timestamp when webhook was created';
COMMENT ON COLUMN "zpWebhooks"."updatedAt" IS 'Timestamp when webhook was last updated';

-- Update zpSessions table comment to reflect webhook separation
COMMENT ON TABLE "zpSessions" IS 'WhatsApp sessions management (camelCase) - webhook config moved to zpWebhooks table';
