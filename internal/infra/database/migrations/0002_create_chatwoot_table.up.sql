-- =============================================================================
-- Chatwoot Integration Table - Migration
-- =============================================================================
-- This migration creates the chatwoot table to store Chatwoot integration
-- configurations for each WhatsApp session.

CREATE TABLE IF NOT EXISTS chatwoot (
    -- Primary identifier - automatically generated UUID
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Foreign key to sessions table
    session_id TEXT NOT NULL,

    -- Chatwoot configuration
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    account_id TEXT,
    token TEXT,
    url TEXT,
    name_inbox TEXT,

    -- Message settings
    sign_msg BOOLEAN DEFAULT FALSE,
    sign_delimiter TEXT DEFAULT '',
    number TEXT DEFAULT '',

    -- Conversation settings
    reopen_conversation BOOLEAN DEFAULT TRUE,
    conversation_pending BOOLEAN DEFAULT FALSE,
    merge_brazil_contacts BOOLEAN DEFAULT TRUE,

    -- Import settings
    import_contacts BOOLEAN DEFAULT FALSE,
    import_messages BOOLEAN DEFAULT FALSE,
    days_limit_import_messages INTEGER DEFAULT 30,

    -- Auto creation settings
    auto_create BOOLEAN DEFAULT FALSE,
    organization TEXT DEFAULT '',
    logo TEXT DEFAULT '',

    -- Ignore settings (JSON array of JIDs to ignore)
    ignore_jids JSONB DEFAULT '[]'::jsonb,

    -- Chatwoot inbox information (populated after creation)
    inbox_id INTEGER,
    inbox_name TEXT,

    -- Integration status
    last_sync TIMESTAMPTZ,
    sync_status TEXT DEFAULT 'pending', -- pending, syncing, completed, error
    error_message TEXT,

    -- Metrics
    messages_count INTEGER DEFAULT 0,
    contacts_count INTEGER DEFAULT 0,
    conversations_count INTEGER DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================================
-- Constraints and Indexes
-- =============================================================================

-- Foreign key constraint to sessions table
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_chatwoot_session_id') THEN
        ALTER TABLE chatwoot
        ADD CONSTRAINT fk_chatwoot_session_id
        FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Unique constraint - one chatwoot config per session
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'unique_chatwoot_session') THEN
        ALTER TABLE chatwoot
        ADD CONSTRAINT unique_chatwoot_session
        UNIQUE (session_id);
    END IF;
END $$;

-- Index on session_id for faster lookups
CREATE INDEX idx_chatwoot_session_id ON chatwoot (session_id);

-- Index on enabled status for filtering
CREATE INDEX idx_chatwoot_enabled ON chatwoot (enabled);

-- Index on account_id and url for external lookups
CREATE INDEX idx_chatwoot_account ON chatwoot (account_id, url) WHERE enabled = TRUE;

-- Index on sync_status for monitoring
CREATE INDEX idx_chatwoot_sync_status ON chatwoot (sync_status);

-- Partial index on inbox_id for active integrations
CREATE INDEX idx_chatwoot_inbox_id ON chatwoot (inbox_id) WHERE enabled = TRUE AND inbox_id IS NOT NULL;

-- =============================================================================
-- Check Constraints
-- =============================================================================

-- Ensure required fields are present when enabled
ALTER TABLE chatwoot 
ADD CONSTRAINT check_chatwoot_enabled_fields 
CHECK (
    NOT enabled OR (
        enabled AND 
        account_id IS NOT NULL AND 
        account_id != '' AND
        token IS NOT NULL AND 
        token != '' AND
        url IS NOT NULL AND 
        url != ''
    )
);

-- Ensure days_limit_import_messages is within reasonable range
ALTER TABLE chatwoot 
ADD CONSTRAINT check_days_limit_range 
CHECK (days_limit_import_messages >= 1 AND days_limit_import_messages <= 365);

-- Ensure messages_count is non-negative
ALTER TABLE chatwoot 
ADD CONSTRAINT check_messages_count_positive 
CHECK (messages_count >= 0);

-- Ensure contacts_count is non-negative
ALTER TABLE chatwoot 
ADD CONSTRAINT check_contacts_count_positive 
CHECK (contacts_count >= 0);

-- Ensure conversations_count is non-negative
ALTER TABLE chatwoot 
ADD CONSTRAINT check_conversations_count_positive 
CHECK (conversations_count >= 0);

-- Ensure sync_status has valid values
ALTER TABLE chatwoot 
ADD CONSTRAINT check_sync_status_valid 
CHECK (sync_status IN ('pending', 'syncing', 'completed', 'error', 'disabled'));

-- =============================================================================
-- Triggers for automatic updated_at
-- =============================================================================

-- Function to update the updated_at timestamp
CREATE OR REPLACE FUNCTION update_chatwoot_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update updated_at on row changes
CREATE TRIGGER trigger_chatwoot_updated_at
    BEFORE UPDATE ON chatwoot
    FOR EACH ROW
    EXECUTE FUNCTION update_chatwoot_updated_at();

-- =============================================================================
-- Comments for Documentation
-- =============================================================================

COMMENT ON TABLE chatwoot IS 'Chatwoot integration configurations for WhatsApp sessions';
COMMENT ON COLUMN chatwoot.id IS 'Unique configuration identifier (UUID)';
COMMENT ON COLUMN chatwoot.session_id IS 'Reference to the WhatsApp session';
COMMENT ON COLUMN chatwoot.enabled IS 'Whether Chatwoot integration is enabled';
COMMENT ON COLUMN chatwoot.account_id IS 'Chatwoot account ID';
COMMENT ON COLUMN chatwoot.token IS 'Chatwoot API access token';
COMMENT ON COLUMN chatwoot.url IS 'Chatwoot instance URL';
COMMENT ON COLUMN chatwoot.name_inbox IS 'Name of the Chatwoot inbox';
COMMENT ON COLUMN chatwoot.sign_msg IS 'Whether to add signature to messages';
COMMENT ON COLUMN chatwoot.sign_delimiter IS 'Signature delimiter text';
COMMENT ON COLUMN chatwoot.number IS 'WhatsApp number for this integration';
COMMENT ON COLUMN chatwoot.reopen_conversation IS 'Whether to reopen resolved conversations';
COMMENT ON COLUMN chatwoot.conversation_pending IS 'Whether to create conversations as pending';
COMMENT ON COLUMN chatwoot.merge_brazil_contacts IS 'Whether to merge Brazilian contacts (9th digit)';
COMMENT ON COLUMN chatwoot.import_contacts IS 'Whether to import existing contacts';
COMMENT ON COLUMN chatwoot.import_messages IS 'Whether to import message history';
COMMENT ON COLUMN chatwoot.days_limit_import_messages IS 'Days limit for message import';
COMMENT ON COLUMN chatwoot.auto_create IS 'Whether to auto-create inbox if not exists';
COMMENT ON COLUMN chatwoot.organization IS 'Organization name for auto-created contacts';
COMMENT ON COLUMN chatwoot.logo IS 'Logo URL for auto-created contacts';
COMMENT ON COLUMN chatwoot.ignore_jids IS 'JSON array of JIDs to ignore';
COMMENT ON COLUMN chatwoot.inbox_id IS 'Chatwoot inbox ID (populated after creation)';
COMMENT ON COLUMN chatwoot.inbox_name IS 'Actual inbox name in Chatwoot';
COMMENT ON COLUMN chatwoot.last_sync IS 'Timestamp of last synchronization';
COMMENT ON COLUMN chatwoot.sync_status IS 'Current synchronization status';
COMMENT ON COLUMN chatwoot.error_message IS 'Last error message if any';
COMMENT ON COLUMN chatwoot.messages_count IS 'Total messages processed';
COMMENT ON COLUMN chatwoot.contacts_count IS 'Total contacts synchronized';
COMMENT ON COLUMN chatwoot.conversations_count IS 'Total conversations created';

-- =============================================================================
-- Sample Data (Optional - for development)
-- =============================================================================

-- Uncomment the following lines to insert sample data for development
-- INSERT INTO chatwoot (session_id, enabled, account_id, token, url, name_inbox, auto_create)
-- SELECT 
--     s.id,
--     FALSE,
--     '1',
--     'sample-token',
--     'https://app.chatwoot.com',
--     s.name || '-chatwoot',
--     TRUE
-- FROM sessions s
-- WHERE NOT EXISTS (SELECT 1 FROM chatwoot c WHERE c.session_id = s.id)
-- LIMIT 1;
