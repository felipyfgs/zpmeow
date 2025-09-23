-- Create messages table
CREATE TABLE IF NOT EXISTS messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    whatsapp_message_id VARCHAR(255) NOT NULL,
    message_type VARCHAR(50) NOT NULL, -- 'text', 'image', 'video', 'audio', 'document', 'sticker', 'location', 'contact', 'system'
    content TEXT,
    media_url VARCHAR(500),
    media_mime_type VARCHAR(100),
    media_size BIGINT,
    media_filename VARCHAR(255),
    thumbnail_url VARCHAR(500),
    sender_jid VARCHAR(255) NOT NULL,
    sender_name VARCHAR(255),
    is_from_me BOOLEAN NOT NULL DEFAULT FALSE,
    is_forwarded BOOLEAN DEFAULT FALSE,
    is_broadcast BOOLEAN DEFAULT FALSE,
    quoted_message_id UUID REFERENCES messages(id),
    quoted_content TEXT,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'sent', 'delivered', 'read', 'failed'
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    edit_timestamp TIMESTAMP WITH TIME ZONE,
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    reaction VARCHAR(10), -- emoji reaction
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id);
CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id);
CREATE INDEX IF NOT EXISTS idx_messages_whatsapp_message_id ON messages(whatsapp_message_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_jid ON messages(sender_jid);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
CREATE INDEX IF NOT EXISTS idx_messages_status ON messages(status);
CREATE INDEX IF NOT EXISTS idx_messages_message_type ON messages(message_type);
CREATE INDEX IF NOT EXISTS idx_messages_is_from_me ON messages(is_from_me);
CREATE INDEX IF NOT EXISTS idx_messages_quoted_message_id ON messages(quoted_message_id);

-- Create unique constraint for session_id + whatsapp_message_id
CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_session_whatsapp_unique ON messages(session_id, whatsapp_message_id);

-- Create trigger function for updated_at
CREATE OR REPLACE FUNCTION update_messages_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger
CREATE TRIGGER trigger_messages_updated_at
    BEFORE UPDATE ON messages
    FOR EACH ROW
    EXECUTE FUNCTION update_messages_updated_at();

-- Comments
COMMENT ON TABLE messages IS 'WhatsApp messages';
COMMENT ON COLUMN messages.id IS 'Unique message identifier (UUID)';
COMMENT ON COLUMN messages.chat_id IS 'Reference to the chat (UUID)';
COMMENT ON COLUMN messages.whatsapp_message_id IS 'WhatsApp message ID';
COMMENT ON COLUMN messages.message_type IS 'Type of message content';
COMMENT ON COLUMN messages.sender_jid IS 'WhatsApp JID of the sender';
COMMENT ON COLUMN messages.is_from_me IS 'Whether message was sent by us';
COMMENT ON COLUMN messages.quoted_message_id IS 'Reference to quoted message (UUID)';
COMMENT ON COLUMN messages.status IS 'Message delivery status';
