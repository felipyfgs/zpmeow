-- Create chatwoot table
CREATE TABLE IF NOT EXISTS chatwoot (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    enabled BOOLEAN DEFAULT FALSE,
    account_id VARCHAR(50),
    token VARCHAR(255),
    url VARCHAR(500),
    name_inbox VARCHAR(255),
    sign_msg BOOLEAN DEFAULT FALSE,
    sign_delimiter VARCHAR(50) DEFAULT '\n\n',
    number VARCHAR(20) NOT NULL,
    reopen_conversation BOOLEAN DEFAULT FALSE,
    conversation_pending BOOLEAN DEFAULT FALSE,
    merge_brazil_contacts BOOLEAN DEFAULT TRUE,
    import_contacts BOOLEAN DEFAULT FALSE,
    import_messages BOOLEAN DEFAULT FALSE,
    days_limit_import_messages INTEGER DEFAULT 0,
    auto_create BOOLEAN DEFAULT TRUE,
    organization VARCHAR(255) DEFAULT 'zpmeow',
    logo VARCHAR(500) DEFAULT '',
    ignore_jids JSONB DEFAULT '[]'::jsonb,
    inbox_id INTEGER,
    inbox_name VARCHAR(255),
    last_sync TIMESTAMP WITH TIME ZONE,
    sync_status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    messages_count INTEGER DEFAULT 0,
    contacts_count INTEGER DEFAULT 0,
    conversations_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_chatwoot_session_id ON chatwoot(session_id);
CREATE INDEX IF NOT EXISTS idx_chatwoot_enabled ON chatwoot(enabled);
CREATE INDEX IF NOT EXISTS idx_chatwoot_sync_status ON chatwoot(sync_status);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_chatwoot_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER trigger_chatwoot_updated_at
    BEFORE UPDATE ON chatwoot
    FOR EACH ROW
    EXECUTE FUNCTION update_chatwoot_updated_at();

-- Comments
COMMENT ON TABLE chatwoot IS 'Chatwoot integration configurations';
COMMENT ON COLUMN chatwoot.id IS 'Unique configuration identifier (UUID)';
COMMENT ON COLUMN chatwoot.session_id IS 'Reference to the WhatsApp session';
COMMENT ON COLUMN chatwoot.enabled IS 'Whether Chatwoot integration is enabled';
COMMENT ON COLUMN chatwoot.account_id IS 'Chatwoot account ID';
COMMENT ON COLUMN chatwoot.token IS 'Chatwoot API access token';
COMMENT ON COLUMN chatwoot.url IS 'Chatwoot instance URL';
