-- Create zpChats table with camelCase columns
CREATE TABLE IF NOT EXISTS "zpChats" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    "sessionId" UUID NOT NULL REFERENCES "zpSessions"(id) ON DELETE CASCADE,
    "chatJid" VARCHAR(255) NOT NULL,
    "chatName" VARCHAR(255),
    "phoneNumber" VARCHAR(50),
    "isGroup" BOOLEAN NOT NULL DEFAULT FALSE,
    "groupSubject" VARCHAR(255),
    "groupDescription" TEXT,
    "chatwootConversationId" BIGINT,
    "chatwootContactId" BIGINT,
    "lastMsgAt" TIMESTAMP WITH TIME ZONE,
    "unreadCount" INTEGER DEFAULT 0,
    "isArchived" BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}'::jsonb,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS "idx_zpChats_sessionId" ON "zpChats"("sessionId");
CREATE INDEX IF NOT EXISTS "idx_zpChats_chatJid" ON "zpChats"("chatJid");
CREATE INDEX IF NOT EXISTS "idx_zpChats_phoneNumber" ON "zpChats"("phoneNumber");
CREATE INDEX IF NOT EXISTS "idx_zpChats_chatwootConversationId" ON "zpChats"("chatwootConversationId");
CREATE INDEX IF NOT EXISTS "idx_zpChats_lastMsgAt" ON "zpChats"("lastMsgAt");

-- Create unique constraint for sessionId + chatJid
CREATE UNIQUE INDEX IF NOT EXISTS "idx_zpChats_session_chat_unique" ON "zpChats"("sessionId", "chatJid");

-- Create trigger function for updatedAt
CREATE OR REPLACE FUNCTION "update_zpChats_updatedAt"()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER "trigger_zpChats_updatedAt"
    BEFORE UPDATE ON "zpChats"
    FOR EACH ROW
    EXECUTE FUNCTION "update_zpChats_updatedAt"();

-- Comments
COMMENT ON TABLE "zpChats" IS 'WhatsApp chats/conversations (camelCase)';
COMMENT ON COLUMN "zpChats".id IS 'Unique chat identifier (UUID)';
COMMENT ON COLUMN "zpChats"."sessionId" IS 'Reference to the WhatsApp session';
COMMENT ON COLUMN "zpChats"."chatJid" IS 'WhatsApp chat JID';
COMMENT ON COLUMN "zpChats"."isGroup" IS 'Whether this is a group chat';
COMMENT ON COLUMN "zpChats"."chatwootConversationId" IS 'Linked Chatwoot conversation ID';
COMMENT ON COLUMN "zpChats"."chatwootContactId" IS 'Linked Chatwoot contact ID';
