-- Create zpChatwoot table with camelCase columns
CREATE TABLE IF NOT EXISTS "zpChatwoot" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"(id) ON DELETE CASCADE,
    "isActive" BOOLEAN DEFAULT FALSE,
    "accountId" VARCHAR(50),
    token VARCHAR(255),
    url VARCHAR(500),
    "nameInbox" VARCHAR(255),
    "inboxId" INTEGER,
    "lastSync" TIMESTAMP WITH TIME ZONE,
    "syncStatus" VARCHAR(50) DEFAULT 'pending',
    config JSONB DEFAULT '{}'::jsonb,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS "idx_zpChatwoot_sessionId" ON "zpChatwoot"("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zpChatwoot_isActive" ON "zpChatwoot"("isActive");
CREATE INDEX IF NOT EXISTS "idx_zpChatwoot_syncStatus" ON "zpChatwoot"("syncStatus");

-- Create trigger function for updatedAt
CREATE OR REPLACE FUNCTION "update_zpChatwoot_updatedAt"()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER "trigger_zpChatwoot_updatedAt"
    BEFORE UPDATE ON "zpChatwoot"
    FOR EACH ROW
    EXECUTE FUNCTION "update_zpChatwoot_updatedAt"();

-- Comments
COMMENT ON TABLE "zpChatwoot" IS 'Chatwoot integration configurations (camelCase)';
COMMENT ON COLUMN "zpChatwoot".id IS 'Unique configuration identifier (UUID)';
COMMENT ON COLUMN "zpChatwoot"."sessionId" IS 'Reference to the WhatsApp session';
COMMENT ON COLUMN "zpChatwoot"."isActive" IS 'Whether Chatwoot integration is active';
COMMENT ON COLUMN "zpChatwoot"."accountId" IS 'Chatwoot account ID';
COMMENT ON COLUMN "zpChatwoot".token IS 'Chatwoot API access token';
COMMENT ON COLUMN "zpChatwoot".url IS 'Chatwoot instance URL';
COMMENT ON COLUMN "zpChatwoot".config IS 'JSONB configuration for flexible settings';
