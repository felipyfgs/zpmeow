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
	Events     string    `db:"webhook_events" json:"webhook_events"` // JSON array as string
	Connected  bool      `db:"connected" json:"connected"`
	ApiKey     string    `db:"apikey" json:"apikey"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

func (SessionModel) TableName() string {
	return "sessions"
}

type MessageModel struct {
	ID        string    `db:"id" json:"id"`
	SessionID string    `db:"session_id" json:"session_id"`
	ChatJID   string    `db:"chat_jid" json:"chat_jid"`
	FromJID   string    `db:"from_jid" json:"from_jid"`
	Content   string    `db:"content" json:"content"`
	Type      string    `db:"type" json:"type"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
	IsFromMe  bool      `db:"is_from_me" json:"is_from_me"`
	IsRead    bool      `db:"is_read" json:"is_read"`
	MediaURL  string    `db:"media_url" json:"media_url"`
	Caption   string    `db:"caption" json:"caption"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func (MessageModel) TableName() string {
	return "messages"
}

type ContactModel struct {
	ID           string    `db:"id" json:"id"`
	SessionID    string    `db:"session_id" json:"session_id"`
	JID          string    `db:"jid" json:"jid"`
	Name         string    `db:"name" json:"name"`
	Notify       string    `db:"notify" json:"notify"`
	PushName     string    `db:"push_name" json:"push_name"`
	BusinessName string    `db:"business_name" json:"business_name"`
	Phone        string    `db:"phone" json:"phone"`
	Organization string    `db:"organization" json:"organization"`
	Email        string    `db:"email" json:"email"`
	IsBlocked    bool      `db:"is_blocked" json:"is_blocked"`
	IsMuted      bool      `db:"is_muted" json:"is_muted"`
	IsContact    bool      `db:"is_contact" json:"is_contact"`
	Avatar       string    `db:"avatar" json:"avatar"`
	Status       string    `db:"status" json:"status"`
	LastSeen     time.Time `db:"last_seen" json:"last_seen"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

func (ContactModel) TableName() string {
	return "contacts"
}

type GroupModel struct {
	ID           string    `db:"id" json:"id"`
	SessionID    string    `db:"session_id" json:"session_id"`
	JID          string    `db:"jid" json:"jid"`
	Name         string    `db:"name" json:"name"`
	Description  string    `db:"description" json:"description"`
	Participants string    `db:"participants" json:"participants"` // JSON array as string
	Admins       string    `db:"admins" json:"admins"`             // JSON array as string
	Owner        string    `db:"owner" json:"owner"`
	IsAnnounce   bool      `db:"is_announce" json:"is_announce"`
	IsLocked     bool      `db:"is_locked" json:"is_locked"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

func (GroupModel) TableName() string {
	return "groups"
}

type ChatModel struct {
	ID            string    `db:"id" json:"id"`
	SessionID     string    `db:"session_id" json:"session_id"`
	JID           string    `db:"jid" json:"jid"`
	Name          string    `db:"name" json:"name"`
	LastMessage   string    `db:"last_message" json:"last_message"`
	LastMessageAt time.Time `db:"last_message_at" json:"last_message_at"`
	UnreadCount   int       `db:"unread_count" json:"unread_count"`
	IsGroup       bool      `db:"is_group" json:"is_group"`
	IsMuted       bool      `db:"is_muted" json:"is_muted"`
	IsArchived    bool      `db:"is_archived" json:"is_archived"`
	IsBlocked     bool      `db:"is_blocked" json:"is_blocked"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

func (ChatModel) TableName() string {
	return "chats"
}

type NewsletterModel struct {
	ID              string    `db:"id" json:"id"`
	SessionID       string    `db:"session_id" json:"session_id"`
	JID             string    `db:"jid" json:"jid"`
	Name            string    `db:"name" json:"name"`
	Description     string    `db:"description" json:"description"`
	SubscriberCount int       `db:"subscriber_count" json:"subscriber_count"`
	IsSubscribed    bool      `db:"is_subscribed" json:"is_subscribed"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

func (NewsletterModel) TableName() string {
	return "newsletters"
}

type WebhookLogModel struct {
	ID           string    `db:"id" json:"id"`
	SessionID    string    `db:"session_id" json:"session_id"`
	WebhookURL   string    `db:"webhook_url" json:"webhook_url"`
	Event        string    `db:"event" json:"event"`
	Payload      string    `db:"payload" json:"payload"` // JSON as string
	StatusCode   int       `db:"status_code" json:"status_code"`
	ResponseBody string    `db:"response_body" json:"response_body"`
	AttemptCount int       `db:"attempt_count" json:"attempt_count"`
	Success      bool      `db:"success" json:"success"`
	Error        string    `db:"error" json:"error"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

func (WebhookLogModel) TableName() string {
	return "webhook_logs"
}
