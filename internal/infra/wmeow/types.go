package wmeow

import (
	"fmt"
	"time"
	"zpmeow/internal/application/ports"
	"go.mau.fi/whatsmeow"
)

// SessionConfiguration holds configuration for a WhatsApp session
type SessionConfiguration struct {
	SessionID   string
	PhoneNumber string
	Status      string
	QRCode      string
	Connected   bool
	Webhook     string
	DeviceJID   string
}

// sessionConfiguration is the internal type used by session management
type sessionConfiguration struct {
	deviceJID string
}

// MessageValidator interface alias for compatibility
type MessageValidator interface {
	ValidateClient(client interface{}) error
	ValidateRecipient(to string) error
	ValidateTextContent(text string) error
	ValidateMediaData(data []byte) error
	ValidateMessageInput(client interface{}, to string) error
}

// MessageBuilder interface alias - actual implementation in helper_messaging.go

// Additional types needed by service files
type MessageInfo struct {
	ID        string    `json:"id"`
	FromMe    bool      `json:"from_me"`
	Timestamp time.Time `json:"timestamp"`
	ChatJID   string    `json:"chat_jid"`
	SenderJID string    `json:"sender_jid,omitempty"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`
}

type UserInfo struct {
	JID       string `json:"jid"`
	Phone     string `json:"phone"`
	Status    string `json:"status"`
	PictureID string `json:"picture_id,omitempty"`
}

type MediaInfo struct {
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
	Filename string `json:"filename,omitempty"`
}

type PrivacySettingInfo struct {
	Category string `json:"category"`
	Value    string `json:"value"`
}

// messageValidatorWrapper adapts the existing messageValidator to the interface
type messageValidatorWrapper struct {
	validator *messageValidator
}

func (w *messageValidatorWrapper) ValidateClient(client interface{}) error {
	if whatsmeowClient, ok := client.(*whatsmeow.Client); ok {
		return w.validator.ValidateClient(whatsmeowClient)
	}
	return fmt.Errorf("invalid client type")
}

func (w *messageValidatorWrapper) ValidateRecipient(to string) error {
	return w.validator.ValidateRecipient(to)
}

func (w *messageValidatorWrapper) ValidateTextContent(text string) error {
	return w.validator.ValidateTextContent(text)
}

func (w *messageValidatorWrapper) ValidateMediaData(data []byte) error {
	return w.validator.ValidateMediaData(data)
}

func (w *messageValidatorWrapper) ValidateMessageInput(client interface{}, to string) error {
	if whatsmeowClient, ok := client.(*whatsmeow.Client); ok {
		return w.validator.ValidateMessageInput(whatsmeowClient, to)
	}
	return fmt.Errorf("invalid client type")
}

// messageBuilderWrapper adapts the existing MessageBuilder to the interface
type messageBuilderWrapper struct {
	builder *MessageBuilder
}

func (w *messageBuilderWrapper) BuildTextMessage(text string) (interface{}, error) {
	return w.builder.BuildTextMessage(text)
}

func (w *messageBuilderWrapper) BuildImageMessage(data []byte, caption string) (interface{}, error) {
	return w.builder.BuildImageMessage(data, caption)
}

func (w *messageBuilderWrapper) BuildAudioMessage(data []byte, mimeType string, ptt bool) (interface{}, error) {
	return w.builder.BuildAudioMessage(data, mimeType, ptt)
}

func (w *messageBuilderWrapper) BuildVideoMessage(data []byte, caption, mimeType string) (interface{}, error) {
	return w.builder.BuildVideoMessage(data, caption, mimeType)
}

func (w *messageBuilderWrapper) BuildDocumentMessage(data []byte, filename, mimeType string) (interface{}, error) {
	return w.builder.BuildDocumentMessage(data, filename, mimeType)
}

func (w *messageBuilderWrapper) BuildStickerMessage(data []byte, mimeType string) (interface{}, error) {
	return w.builder.BuildStickerMessage(data, mimeType)
}

func (w *messageBuilderWrapper) BuildContactsMessage(contacts []ports.ContactData) (interface{}, error) {
	return w.builder.BuildContactsMessage(contacts)
}

func (w *messageBuilderWrapper) BuildLocationMessage(latitude, longitude float64, name, address string) (interface{}, error) {
	return w.builder.BuildLocationMessage(latitude, longitude, name, address)
}
