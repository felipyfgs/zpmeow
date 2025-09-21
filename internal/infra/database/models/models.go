package models

import (
	"time"
)

type SessionModel struct {
	ID         string    `db:"id" json:"id"`                         // Primary key (UUID)
	Name       string    `db:"name" json:"name"`                     // Unique session name
	DeviceJID  string    `db:"device_jid" json:"device_jid"`         // WhatsApp device JID when connected
	Status     string    `db:"status" json:"status"`                 // Session status: disconnected, connecting, connected
	QRCode     string    `db:"qr_code" json:"qr_code"`               // QR code for WhatsApp pairing
	ProxyURL   string    `db:"proxy_url" json:"proxy_url"`           // Optional proxy URL for connection
	WebhookURL string    `db:"webhook_url" json:"webhook_url"`       // Webhook URL for events
	Events     string    `db:"webhook_events" json:"webhook_events"` // JSON array of subscribed events as string
	Connected  bool      `db:"connected" json:"connected"`           // Boolean flag for connection state
	ApiKey     string    `db:"apikey" json:"apikey"`                 // API key for session authentication
	CreatedAt  time.Time `db:"created_at" json:"created_at"`         // Record creation timestamp
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`         // Record last update timestamp
}

func (SessionModel) TableName() string {
	return "sessions"
}
