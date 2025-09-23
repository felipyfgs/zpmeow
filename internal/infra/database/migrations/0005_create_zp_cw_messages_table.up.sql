-- Create zp_cw_messages table
CREATE TABLE IF NOT EXISTS zp_cw_messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    zpmeow_message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    chatwoot_message_id BIGINT NOT NULL,
    chatwoot_conversation_id BIGINT NOT NULL,
    chatwoot_account_id BIGINT NOT NULL,
    direction VARCHAR(20) NOT NULL, -- 'incoming', 'outgoing'
    sync_status VARCHAR(50) DEFAULT 'synced', -- 'pending', 'synced', 'failed', 'partial'
    sync_error TEXT,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    chatwoot_source_id VARCHAR(255), -- WAID:message_id for WhatsApp messages
    chatwoot_echo_id VARCHAR(255), -- Echo ID for outgoing messages
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_session_id ON zp_cw_messages(session_id);
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_zpmeow_message_id ON zp_cw_messages(zpmeow_message_id);
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_chatwoot_message_id ON zp_cw_messages(chatwoot_message_id);
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_chatwoot_conversation_id ON zp_cw_messages(chatwoot_conversation_id);
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_direction ON zp_cw_messages(direction);
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_sync_status ON zp_cw_messages(sync_status);
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_chatwoot_source_id ON zp_cw_messages(chatwoot_source_id);
CREATE INDEX IF NOT EXISTS idx_zp_cw_messages_chatwoot_echo_id ON zp_cw_messages(chatwoot_echo_id);

-- Create unique constraints
CREATE UNIQUE INDEX IF NOT EXISTS idx_zp_cw_messages_zpmeow_unique ON zp_cw_messages(zpmeow_message_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_zp_cw_messages_chatwoot_unique ON zp_cw_messages(session_id, chatwoot_message_id);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_zp_cw_messages_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER trigger_zp_cw_messages_updated_at
    BEFORE UPDATE ON zp_cw_messages
    FOR EACH ROW
    EXECUTE FUNCTION update_zp_cw_messages_updated_at();

-- Comments
COMMENT ON TABLE zp_cw_messages IS 'Relations between zpmeow and chatwoot messages';
COMMENT ON COLUMN zp_cw_messages.id IS 'Unique relation identifier (UUID)';
COMMENT ON COLUMN zp_cw_messages.zpmeow_message_id IS 'Reference to zpmeow message (UUID)';
COMMENT ON COLUMN zp_cw_messages.chatwoot_message_id IS 'Chatwoot message ID (integer)';
COMMENT ON COLUMN zp_cw_messages.chatwoot_source_id IS 'Chatwoot source ID (WAID:message_id)';
COMMENT ON COLUMN zp_cw_messages.direction IS 'Message direction: incoming or outgoing';
COMMENT ON COLUMN zp_cw_messages.sync_status IS 'Synchronization status';
COMMENT ON COLUMN zp_cw_messages.chatwoot_echo_id IS 'Echo ID for outgoing messages';
