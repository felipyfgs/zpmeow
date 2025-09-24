package chatwoot

import "time"

// Cache TTL constants
const (
	DefaultCacheTTL      = 5 * time.Minute
	ContactCacheTTL      = 10 * time.Minute
	ConversationCacheTTL = 15 * time.Minute
)

// Cache keys
const (
	CacheKeyContact      = "contact:%s"
	CacheKeyConversation = "conversation:%d:%d"
	CacheKeyInbox        = "inbox:%d"
)

// File type constants
const (
	FileTypeAudio    = "audio"
	FileTypeImage    = "image"
	FileTypeVideo    = "video"
	FileTypeDocument = "document"
	FileTypeFile     = "file"
)

// Audio file extensions
var AudioExtensions = []string{".mp3", ".wav", ".ogg", ".aac", ".m4a", ".flac", ".wma"}

// Image file extensions
var ImageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".svg", ".tiff"}

// Video file extensions
var VideoExtensions = []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".webm", ".mkv", ".m4v"}

// Default MIME types
const (
	DefaultImageMimeType    = "image/jpeg"
	DefaultVideoMimeType    = "video/mp4"
	DefaultAudioMimeType    = "audio/ogg"
	DefaultDocumentMimeType = "application/octet-stream"
)

// Default file names
const (
	DefaultAudioFileName    = "audio.mp3"
	DefaultImageFileName    = "image.jpg"
	DefaultVideoFileName    = "video.mp4"
	DefaultDocumentFileName = "document.pdf"
)

// HTTP timeouts
const (
	DefaultHTTPTimeout     = 30 // seconds
	DefaultDownloadTimeout = 60 // seconds
)

// Chatwoot API endpoints
const (
	EndpointInboxes       = "/api/v1/accounts/%s/inboxes"
	EndpointContacts      = "/api/v1/accounts/%s/contacts"
	EndpointConversations = "/api/v1/accounts/%s/conversations"
	EndpointMessages      = "/api/v1/accounts/%s/conversations/%d/messages"
)

// Message types
const (
	MessageTypeIncoming = 0
	MessageTypeOutgoing = 1
)

// Conversation statuses
const (
	ConversationStatusOpenStr     = "open"
	ConversationStatusResolvedStr = "resolved"
	ConversationStatusPendingStr  = "pending"
)
