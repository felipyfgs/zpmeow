package wmeow

import "time"

// Media size limits (in bytes)
const (
	MaxImageSize    = 16 * 1024 * 1024  // 16MB
	MaxVideoSize    = 64 * 1024 * 1024  // 64MB
	MaxAudioSize    = 16 * 1024 * 1024  // 16MB
	MaxDocumentSize = 100 * 1024 * 1024 // 100MB
	MaxStickerSize  = 500 * 1024        // 500KB
)

// Message limits
const (
	MaxTextMessageLength = 4096
	MaxCaptionLength     = 1024
	MaxFileNameLength    = 255
	MaxButtonText        = 20
	MaxListRowTitle      = 24
	MaxListRowDesc       = 72
)

// Button and list limits
const (
	MaxButtons      = 3
	MaxListSections = 10
	MaxListRows     = 10
	MaxPollOptions  = 12
)

// Phone number validation
const (
	MinPhoneLength = 10
	MaxPhoneLength = 15
)

// Session validation
const (
	MinSessionIDLength = 8
)

// Timeouts
const (
	DefaultConnectionTimeout = 30 * time.Second
	DefaultOperationTimeout  = 60 * time.Second
	DefaultQRCodeTimeout     = 120 * time.Second
	DefaultPairTimeout       = 60 * time.Second
)

// Retry configuration
const (
	DefaultMaxRetries    = 3
	DefaultRetryDelay    = 1 * time.Second
	DefaultBackoffFactor = 2.0
)

// WhatsApp presence states
const (
	PresenceAvailable   = "available"
	PresenceUnavailable = "unavailable"
	PresenceComposing   = "composing"
	PresenceRecording   = "recording"
	PresencePaused      = "paused"
)

// WhatsApp media types
const (
	MediaTypeImage    = "image"
	MediaTypeVideo    = "video"
	MediaTypeAudio    = "audio"
	MediaTypeDocument = "document"
	MediaTypeSticker  = "sticker"
)

// Group actions
const (
	GroupActionAdd     = "add"
	GroupActionRemove  = "remove"
	GroupActionPromote = "promote"
	GroupActionDemote  = "demote"
	GroupActionApprove = "approve"
	GroupActionReject  = "reject"
)

// Privacy settings
const (
	PrivacyEveryone   = "all"
	PrivacyContacts   = "contacts"
	PrivacyNobody     = "none"
	PrivacyMyContacts = "contact_blacklist"
)

// Chat types
const (
	ChatTypeAll        = "all"
	ChatTypePersonal   = "personal"
	ChatTypeGroup      = "group"
	ChatTypeBroadcast  = "broadcast"
	ChatTypeNewsletter = "newsletter"
)

// Message types for WhatsApp
const (
	MessageTypeText     = "text"
	MessageTypeImage    = "image"
	MessageTypeVideo    = "video"
	MessageTypeAudio    = "audio"
	MessageTypeDocument = "document"
	MessageTypeSticker  = "sticker"
	MessageTypeLocation = "location"
	MessageTypeContact  = "contact"
	MessageTypeButton   = "button"
	MessageTypeList     = "list"
	MessageTypePoll     = "poll"
	MessageTypeReaction = "reaction"
)

// Error messages
const (
	ErrSessionNotFound     = "session not found"
	ErrClientNotConnected  = "client not connected"
	ErrInvalidPhoneNumber  = "invalid phone number"
	ErrInvalidSessionID    = "invalid session ID"
	ErrMessageTooLong      = "message too long"
	ErrMediaTooLarge       = "media file too large"
	ErrInvalidMediaType    = "invalid media type"
	ErrTooManyButtons      = "too many buttons"
	ErrTooManyListSections = "too many list sections"
	ErrTooManyListRows     = "too many list rows"
)

// Default values
const (
	DefaultQRCodeSize = 256
	DefaultLimit      = 50
	DefaultOffset     = 0
)

// Cache keys for WMeow
const (
	CacheKeySession  = "session:%s"
	CacheKeyContact  = "contact:%s:%s"
	CacheKeyGroup    = "group:%s:%s"
	CacheKeyMedia    = "media:%s:%s"
	CacheKeyQRCode   = "qr:%s"
	CacheKeyPresence = "presence:%s:%s"
)

// Event types
const (
	EventMessage          = "message"
	EventMessageUpdate    = "message.update"
	EventMessageDelete    = "message.delete"
	EventPresence         = "presence"
	EventContact          = "contact"
	EventGroup            = "group"
	EventGroupParticipant = "group.participant"
	EventCall             = "call"
	EventConnection       = "connection"
	EventQRCode           = "qr"
	EventPair             = "pair"
)

// Connection states
const (
	StateDisconnected = "disconnected"
	StateConnecting   = "connecting"
	StateConnected    = "connected"
	StateReconnecting = "reconnecting"
	StateLoggedOut    = "logged_out"
)

// File extensions by type
var (
	ImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	VideoExtensions = []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv"}
	AudioExtensions = []string{".mp3", ".wav", ".ogg", ".aac", ".m4a", ".flac", ".opus"}
	DocExtensions   = []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt"}
)

// MIME types
var (
	ImageMimeTypes = []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/webp"}
	VideoMimeTypes = []string{"video/mp4", "video/avi", "video/quicktime", "video/webm"}
	AudioMimeTypes = []string{"audio/mpeg", "audio/wav", "audio/ogg", "audio/aac", "audio/opus"}
	DocMimeTypes   = []string{"application/pdf", "application/msword", "text/plain"}
)
