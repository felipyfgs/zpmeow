-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create sessions table
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    device_jid VARCHAR(255),
    status VARCHAR(50) DEFAULT 'disconnected',
    qr_code TEXT,
    proxy_url VARCHAR(500),
    webhook_url VARCHAR(500),
    webhook_events VARCHAR(500) DEFAULT 'message',
    connected BOOLEAN DEFAULT FALSE,
    apikey VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_sessions_name ON sessions(name);
CREATE INDEX IF NOT EXISTS idx_sessions_apikey ON sessions(apikey);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_sessions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER trigger_sessions_updated_at
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_sessions_updated_at();

-- Comments
COMMENT ON TABLE sessions IS 'WhatsApp sessions management';
COMMENT ON COLUMN sessions.id IS 'Unique session identifier (UUID)';
COMMENT ON COLUMN sessions.name IS 'Human-readable session name (unique)';
COMMENT ON COLUMN sessions.device_jid IS 'WhatsApp device JID when connected';
COMMENT ON COLUMN sessions.status IS 'Session status: disconnected, connecting, connected';
COMMENT ON COLUMN sessions.connected IS 'Boolean flag for connection state';
COMMENT ON COLUMN sessions.qr_code IS 'QR code for WhatsApp pairing';
COMMENT ON COLUMN sessions.apikey IS 'API key for session authentication';
