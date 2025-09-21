package models

import (
	"time"
)

type SessionModel struct {
	ID         string    `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	DeviceJID  string    `db:"device_jid" json:"device_jid"`
	Status     string    `db:"status" json:"status"`
	QRCode     string    `db:"qr_code" json:"qr_code"`
	ProxyURL   string    `db:"proxy_url" json:"proxy_url"`
	WebhookURL string    `db:"webhook_url" json:"webhook_url"`
	Events     string    `db:"webhook_events" json:"webhook_events"`
	Connected  bool      `db:"connected" json:"connected"`
	ApiKey     string    `db:"apikey" json:"apikey"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

func (SessionModel) TableName() string {
	return "sessions"
}
