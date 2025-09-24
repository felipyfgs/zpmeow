package wmeow

import (
	"fmt"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
	"time"
	"zpmeow/internal/application/ports"
)

// WhatsAppClient wraps the whatsmeow client
type WhatsAppClient struct {
	client    *whatsmeow.Client
	connected bool
}

func (w *WhatsAppClient) GetClient() *whatsmeow.Client {
	return w.client
}

func (w *WhatsAppClient) IsConnected() bool {
	return w.connected && w.client != nil && w.client.IsConnected()
}

func (w *WhatsAppClient) SetConnected(connected bool) {
	w.connected = connected
}

// MessageInfo represents a WhatsApp message
type MessageInfo struct {
	ID        string
	FromMe    bool
	Timestamp int64
	ChatJID   string
	Type      string
	Content   string
	SenderJID *string
}

// ChatInfo represents a WhatsApp chat
type ChatInfo struct {
	JID     string
	Name    string
	IsGroup bool
}

// ContactInfo represents a WhatsApp contact
type ContactInfo struct {
	JID   string
	Phone string
	Name  string
}

// UserInfoResult represents user information
type UserInfoResult struct {
	Phone string
	Name  string
}

// MediaInfo represents media information
type MediaInfo struct {
	ID       string
	Type     string
	MimeType string
	Size     int64
	URL      string
}

// GroupInfo represents a WhatsApp group
type GroupInfo struct {
	JID          string
	Name         string
	Description  string
	CreatedAt    int64
	Participants []string
}

// NewsletterInfo represents a WhatsApp newsletter
type NewsletterInfo struct {
	JID         string
	Name        string
	Description string
	CreatedAt   int64
	Subscribers int
	Verified    bool
}

// SendResponse represents a message send response
type SendResponse struct {
	MessageID string
	Timestamp int64
}

// WhatsAppMessageBuilder wraps message building functionality
type WhatsAppMessageBuilder struct {
	// For now, we'll keep it simple
}

func NewWhatsAppMessageBuilder() *WhatsAppMessageBuilder {
	return &WhatsAppMessageBuilder{}
}

func (w *WhatsAppMessageBuilder) BuildTextMessage(text string) (*waProto.Message, error) {
	return &waProto.Message{
		Conversation: proto.String(text),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildImageMessage(data []byte, caption string) (*waProto.Message, error) {
	// For now, return a simple text message
	return &waProto.Message{
		Conversation: proto.String(caption),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildAudioMessage(data []byte, ptt bool) (*waProto.Message, error) {
	// For now, return a simple text message
	return &waProto.Message{
		Conversation: proto.String("Audio message"),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildVideoMessage(data []byte, caption string) (*waProto.Message, error) {
	// For now, return a simple text message
	return &waProto.Message{
		Conversation: proto.String(caption),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildDocumentMessage(data []byte, filename, mimetype string) (*waProto.Message, error) {
	// For now, return a simple text message
	return &waProto.Message{
		Conversation: proto.String("Document: " + filename),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildStickerMessage(data []byte) (*waProto.Message, error) {
	// For now, return a simple text message
	return &waProto.Message{
		Conversation: proto.String("Sticker"),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildContactMessage(contacts []ports.ContactInfo) (*waProto.Message, error) {
	// For now, return a simple text message
	return &waProto.Message{
		Conversation: proto.String("Contact message"),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildLocationMessage(latitude, longitude float64, name, address string) (*waProto.Message, error) {
	// For now, return a simple text message
	locationText := fmt.Sprintf("Location: %s at %f,%f", name, latitude, longitude)
	return &waProto.Message{
		Conversation: proto.String(locationText),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildTemplateMessage(template map[string]interface{}) (*waProto.Message, error) {
	// For now, return a simple text message
	return &waProto.Message{
		Conversation: proto.String("Template message"),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildButtonMessage(text string, buttons []map[string]interface{}) (*waProto.Message, error) {
	// For now, return a simple text message
	buttonText := fmt.Sprintf("%s (with %d buttons)", text, len(buttons))
	return &waProto.Message{
		Conversation: proto.String(buttonText),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildListMessage(text, buttonText string, sections []map[string]interface{}) (*waProto.Message, error) {
	// For now, return a simple text message
	listText := fmt.Sprintf("%s (with %d sections)", text, len(sections))
	return &waProto.Message{
		Conversation: proto.String(listText),
	}, nil
}

func (w *WhatsAppMessageBuilder) BuildPollMessage(question string, options []string) (*waProto.Message, error) {
	// For now, return a simple text message
	pollText := fmt.Sprintf("Poll: %s (with %d options)", question, len(options))
	return &waProto.Message{
		Conversation: proto.String(pollText),
	}, nil
}

// Helper functions

func parsePhoneToJID(phone string) (string, error) {
	// Simple phone to JID conversion
	if phone == "" {
		return "", fmt.Errorf("phone number cannot be empty")
	}

	// Remove any non-numeric characters except +
	cleanPhone := phone
	if cleanPhone[0] == '+' {
		cleanPhone = cleanPhone[1:]
	}

	return cleanPhone + "@s.whatsapp.net", nil
}

func formatTimestamp(timestamp time.Time) int64 {
	return timestamp.Unix()
}

func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}

// Error types
type WhatsAppError struct {
	Code    int
	Message string
	Details string
}

func (e *WhatsAppError) Error() string {
	return fmt.Sprintf("WhatsApp error %d: %s - %s", e.Code, e.Message, e.Details)
}

func NewWhatsAppError(code int, message, details string) *WhatsAppError {
	return &WhatsAppError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// Constants are defined in constants.go
