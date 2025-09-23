-- Create chats table
CREATE TABLE IF NOT EXISTS chats (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    chat_jid VARCHAR(255) NOT NULL,
    chat_name VARCHAR(255),
    chat_type VARCHAR(50) NOT NULL DEFAULT 'individual', -- 'individual', 'group'
    phone_number VARCHAR(50),
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    group_subject VARCHAR(255),
    group_description TEXT,
    chatwoot_conversation_id BIGINT,
    chatwoot_contact_id BIGINT,
    last_message_at TIMESTAMP WITH TIME ZONE,
    unread_count INTEGER DEFAULT 0,
    is_archived BOOLEAN DEFAULT FALSE,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_chats_session_id ON chats(session_id);
CREATE INDEX IF NOT EXISTS idx_chats_chat_jid ON chats(chat_jid);
CREATE INDEX IF NOT EXISTS idx_chats_phone_number ON chats(phone_number);
CREATE INDEX IF NOT EXISTS idx_chats_chatwoot_conversation_id ON chats(chatwoot_conversation_id);
CREATE INDEX IF NOT EXISTS idx_chats_last_message_at ON chats(last_message_at);

-- Create unique constraint for session_id + chat_jid
CREATE UNIQUE INDEX IF NOT EXISTS idx_chats_session_chat_unique ON chats(session_id, chat_jid);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_chats_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER trigger_chats_updated_at
    BEFORE UPDATE ON chats
    FOR EACH ROW
    EXECUTE FUNCTION update_chats_updated_at();

-- Comments
COMMENT ON TABLE chats IS 'WhatsApp chats/conversations';
COMMENT ON COLUMN chats.id IS 'Unique chat identifier (UUID)';
COMMENT ON COLUMN chats.session_id IS 'Reference to the WhatsApp session';
COMMENT ON COLUMN chats.chat_jid IS 'WhatsApp chat JID';
COMMENT ON COLUMN chats.chat_type IS 'Type of chat: individual or group';
COMMENT ON COLUMN chats.is_group IS 'Whether this is a group chat';
COMMENT ON COLUMN chats.chatwoot_conversation_id IS 'Linked Chatwoot conversation ID';
COMMENT ON COLUMN chats.chatwoot_contact_id IS 'Linked Chatwoot contact ID';
