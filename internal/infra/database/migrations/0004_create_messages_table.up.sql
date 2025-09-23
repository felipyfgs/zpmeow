-- Create zpMessages table with camelCase columns
CREATE TABLE IF NOT EXISTS "zpMessages" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "chatId" UUID NOT NULL REFERENCES "zpChats"(id) ON DELETE CASCADE,
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"(id) ON DELETE CASCADE,
    "msgId" VARCHAR(255) NOT NULL, -- WhatsApp message ID (nome mais curto)
    "msgType" VARCHAR(50) NOT NULL, -- 'text', 'image', 'video', 'audio', 'document', 'sticker', 'location', 'contact', 'system'
    content TEXT,
    "senderJid" VARCHAR(255) NOT NULL,
    "senderName" VARCHAR(255),
    "isFromMe" BOOLEAN NOT NULL DEFAULT FALSE,
    "isForwarded" BOOLEAN DEFAULT FALSE,
    "isBroadcast" BOOLEAN DEFAULT FALSE,
    "quotedMsgId" UUID REFERENCES "zpMessages"(id),
    "quotedContent" TEXT,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'sent', 'delivered', 'read', 'failed'
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    "editTimestamp" TIMESTAMP WITH TIME ZONE,
    "isDeleted" BOOLEAN DEFAULT FALSE,
    "deletedAt" TIMESTAMP WITH TIME ZONE,
    reaction VARCHAR(10), -- emoji reaction
    "mediaInfo" JSONB DEFAULT '{}'::jsonb, -- stores media URL, mimeType, filename, size, thumbnail
    metadata JSONB DEFAULT '{}'::jsonb,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS "idx_zpMessages_chatId" ON "zpMessages"("chatId");
CREATE INDEX IF NOT EXISTS "idx_zpMessages_sessionId" ON "zpMessages"("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zpMessages_msgId" ON "zpMessages"("msgId"); -- WhatsApp message ID
CREATE INDEX IF NOT EXISTS "idx_zpMessages_senderJid" ON "zpMessages"("senderJid");
CREATE INDEX IF NOT EXISTS "idx_zpMessages_timestamp" ON "zpMessages"(timestamp);
CREATE INDEX IF NOT EXISTS "idx_zpMessages_status" ON "zpMessages"(status);
CREATE INDEX IF NOT EXISTS "idx_zpMessages_msgType" ON "zpMessages"("msgType");
CREATE INDEX IF NOT EXISTS "idx_zpMessages_isFromMe" ON "zpMessages"("isFromMe");
CREATE INDEX IF NOT EXISTS "idx_zpMessages_quotedMsgId" ON "zpMessages"("quotedMsgId");

-- Create unique constraint for sessionId + msgId (camelCase com aspas duplas)
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zpMessages_session_msgId_unique" ON "zpMessages"("sessionId", "msgId");

-- Create trigger function for updatedAt
CREATE OR REPLACE FUNCTION "update_zpMessages_updatedAt"()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER "trigger_zpMessages_updatedAt"
    BEFORE UPDATE ON "zpMessages"
    FOR EACH ROW
    EXECUTE FUNCTION "update_zpMessages_updatedAt"();

-- Comments
COMMENT ON TABLE "zpMessages" IS 'WhatsApp messages (camelCase)';
COMMENT ON COLUMN "zpMessages".id IS 'Unique message identifier (UUID)';
COMMENT ON COLUMN "zpMessages"."chatId" IS 'Reference to the chat (UUID)';
COMMENT ON COLUMN "zpMessages"."msgId" IS 'WhatsApp message ID';
COMMENT ON COLUMN "zpMessages"."msgType" IS 'Type of message content';
COMMENT ON COLUMN "zpMessages"."senderJid" IS 'WhatsApp JID of the sender';
COMMENT ON COLUMN "zpMessages"."isFromMe" IS 'Whether message was sent by us';
COMMENT ON COLUMN "zpMessages"."quotedMsgId" IS 'Reference to quoted message (UUID)';
COMMENT ON COLUMN "zpMessages".status IS 'Message delivery status';
COMMENT ON COLUMN "zpMessages"."mediaInfo" IS 'JSONB with media information (URL, mimeType, filename, size, thumbnail)';
