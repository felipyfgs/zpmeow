-- Create zpCwMessages table with camelCase columns
CREATE TABLE IF NOT EXISTS "zpCwMessages" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"(id) ON DELETE CASCADE,
    "msgId" UUID NOT NULL REFERENCES "zpMessages"(id) ON DELETE CASCADE,
    "chatwootMsgId" BIGINT NOT NULL,
    "chatwootConvId" BIGINT NOT NULL,
    direction VARCHAR(20) NOT NULL, -- 'incoming', 'outgoing'
    "syncStatus" VARCHAR(50) DEFAULT 'synced', -- 'pending', 'synced', 'failed', 'partial'
    "sourceId" VARCHAR(255), -- WAID:message_id for WhatsApp messages
    metadata JSONB DEFAULT '{}'::jsonb,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS "idx_zpCwMessages_sessionId" ON "zpCwMessages"("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zpCwMessages_msgId" ON "zpCwMessages"("msgId");
CREATE INDEX IF NOT EXISTS "idx_zpCwMessages_chatwootMsgId" ON "zpCwMessages"("chatwootMsgId");
CREATE INDEX IF NOT EXISTS "idx_zpCwMessages_chatwootConvId" ON "zpCwMessages"("chatwootConvId");
CREATE INDEX IF NOT EXISTS "idx_zpCwMessages_direction" ON "zpCwMessages"(direction);
CREATE INDEX IF NOT EXISTS "idx_zpCwMessages_syncStatus" ON "zpCwMessages"("syncStatus");
CREATE INDEX IF NOT EXISTS "idx_zpCwMessages_sourceId" ON "zpCwMessages"("sourceId");

-- Create unique constraints
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zpCwMessages_msgId_unique" ON "zpCwMessages"("msgId");
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zpCwMessages_chatwoot_unique" ON "zpCwMessages"("sessionId", "chatwootMsgId");

-- Create trigger function for updatedAt
CREATE OR REPLACE FUNCTION "update_zpCwMessages_updatedAt"()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER "trigger_zpCwMessages_updatedAt"
    BEFORE UPDATE ON "zpCwMessages"
    FOR EACH ROW
    EXECUTE FUNCTION "update_zpCwMessages_updatedAt"();

-- Comments
COMMENT ON TABLE "zpCwMessages" IS 'Relations between zpmeow and chatwoot messages (camelCase)';
COMMENT ON COLUMN "zpCwMessages".id IS 'Unique relation identifier (UUID)';
COMMENT ON COLUMN "zpCwMessages"."msgId" IS 'Reference to zpmeow message (UUID)';
COMMENT ON COLUMN "zpCwMessages"."chatwootMsgId" IS 'Chatwoot message ID (integer)';
COMMENT ON COLUMN "zpCwMessages"."sourceId" IS 'Chatwoot source ID (WAID:message_id)';
COMMENT ON COLUMN "zpCwMessages".direction IS 'Message direction: incoming or outgoing';
COMMENT ON COLUMN "zpCwMessages"."syncStatus" IS 'Synchronization status';
COMMENT ON COLUMN "zpCwMessages".metadata IS 'Additional metadata in JSONB format';
