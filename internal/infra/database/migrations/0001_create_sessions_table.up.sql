-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create zpSessions table with camelCase columns
CREATE TABLE IF NOT EXISTS "zpSessions" (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    "deviceJid" VARCHAR(255),
    status VARCHAR(50) DEFAULT 'disconnected',
    "qrCode" TEXT,
    "proxyUrl" VARCHAR(500),
    "webhookUrl" VARCHAR(500),
    "webhookEvents" VARCHAR(500) DEFAULT 'message',
    connected BOOLEAN DEFAULT FALSE,
    "apiKey" VARCHAR(255) NOT NULL UNIQUE,
    "createdAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS "idx_zpSessions_name" ON "zpSessions"(name);
CREATE INDEX IF NOT EXISTS "idx_zpSessions_apiKey" ON "zpSessions"("apiKey");
CREATE INDEX IF NOT EXISTS "idx_zpSessions_status" ON "zpSessions"(status);

-- Create trigger function for updatedAt
CREATE OR REPLACE FUNCTION "update_zpSessions_updatedAt"()
RETURNS TRIGGER AS $$
BEGIN
    NEW."updatedAt" = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER "trigger_zpSessions_updatedAt"
    BEFORE UPDATE ON "zpSessions"
    FOR EACH ROW
    EXECUTE FUNCTION "update_zpSessions_updatedAt"();

-- Comments
COMMENT ON TABLE "zpSessions" IS 'WhatsApp sessions management (camelCase)';
COMMENT ON COLUMN "zpSessions".id IS 'Unique session identifier (UUID)';
COMMENT ON COLUMN "zpSessions".name IS 'Human-readable session name (unique)';
COMMENT ON COLUMN "zpSessions"."deviceJid" IS 'WhatsApp device JID when connected';
COMMENT ON COLUMN "zpSessions".status IS 'Session status: disconnected, connecting, connected';
COMMENT ON COLUMN "zpSessions".connected IS 'Boolean flag for connection state';
COMMENT ON COLUMN "zpSessions"."qrCode" IS 'QR code for WhatsApp pairing';
COMMENT ON COLUMN "zpSessions"."apiKey" IS 'API key for session authentication';
