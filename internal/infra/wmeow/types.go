package wmeow

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
