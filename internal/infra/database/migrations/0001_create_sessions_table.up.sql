-- =============================================================================
-- Sessions Table - Consolidated Migration
-- =============================================================================
-- This migration creates the complete sessions table with all fields and constraints
-- including automatic UUID generation and proper indexing.

CREATE TABLE IF NOT EXISTS sessions (
    -- Primary identifier - automatically generated UUID
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Session identification
    name TEXT NOT NULL,
    device_jid TEXT,

    -- Session state
    status TEXT NOT NULL DEFAULT 'disconnected',
    connected BOOLEAN DEFAULT FALSE,

    -- WhatsApp data
    qr_code TEXT,

    -- Configuration
    proxy_url TEXT,
    webhook_url TEXT DEFAULT '',
    webhook_events TEXT DEFAULT '',

    -- Authentication - API key for session access
    apikey TEXT NOT NULL DEFAULT '',

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- Constraints and Indexes
-- =============================================================================

-- Unique constraint on session names
ALTER TABLE sessions ADD CONSTRAINT unique_session_name UNIQUE (name);

-- Unique index on API keys (excluding empty values)
CREATE UNIQUE INDEX idx_sessions_apikey ON sessions (apikey) WHERE apikey != '';

-- =============================================================================
-- Comments for Documentation
-- =============================================================================

COMMENT ON TABLE sessions IS 'WhatsApp sessions management table';
COMMENT ON COLUMN sessions.id IS 'Unique session identifier (UUID)';
COMMENT ON COLUMN sessions.name IS 'Human-readable session name (unique)';
COMMENT ON COLUMN sessions.device_jid IS 'WhatsApp device JID when connected';
COMMENT ON COLUMN sessions.status IS 'Session status: disconnected, connecting, connected';
COMMENT ON COLUMN sessions.connected IS 'Boolean flag for connection state';
COMMENT ON COLUMN sessions.qr_code IS 'QR code for WhatsApp pairing';
COMMENT ON COLUMN sessions.proxy_url IS 'Optional proxy URL for connection';
COMMENT ON COLUMN sessions.webhook_url IS 'Webhook URL for events';
COMMENT ON COLUMN sessions.webhook_events IS 'JSON array of subscribed events';
COMMENT ON COLUMN sessions.apikey IS 'API key for session authentication';
