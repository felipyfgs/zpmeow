package chatwoot

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
)

// FileTypeDetector provides utilities for detecting file types
type FileTypeDetector struct{}

// NewFileTypeDetector creates a new file type detector
func NewFileTypeDetector() *FileTypeDetector {
	return &FileTypeDetector{}
}

// DetectFileType determines the actual file type based on URL, extension or MIME type
func (d *FileTypeDetector) DetectFileType(attachment Attachment) string {
	if d.isKnownSpecificType(attachment.FileType) {
		return attachment.FileType
	}

	fileName := d.extractFileName(attachment)
	extension := strings.ToLower(filepath.Ext(fileName))

	return d.getTypeByExtension(extension)
}

// isKnownSpecificType checks if the file type is already a known specific type
func (d *FileTypeDetector) isKnownSpecificType(fileType string) bool {
	return fileType == FileTypeAudio || fileType == FileTypeImage || fileType == FileTypeVideo
}

// extractFileName extracts filename from attachment
func (d *FileTypeDetector) extractFileName(attachment Attachment) string {
	if attachment.Fallback != "" {
		return attachment.Fallback
	}

	if attachment.DataURL != "" {
		if fileName := d.extractFileNameFromURL(attachment.DataURL); fileName != "" {
			return fileName
		}
	}

	return d.getDefaultFileName(attachment.FileType)
}

// extractFileNameFromURL extracts filename from URL
func (d *FileTypeDetector) extractFileNameFromURL(dataURL string) string {
	u, err := url.Parse(dataURL)
	if err != nil {
		return ""
	}

	fileName := filepath.Base(u.Path)
	if fileName == "" || fileName == "." || fileName == "/" {
		return ""
	}

	if decoded, err := url.QueryUnescape(fileName); err == nil {
		return decoded
	}

	return fileName
}

// getTypeByExtension returns file type based on extension
func (d *FileTypeDetector) getTypeByExtension(extension string) string {
	for _, ext := range AudioExtensions {
		if extension == ext {
			return FileTypeAudio
		}
	}

	for _, ext := range ImageExtensions {
		if extension == ext {
			return FileTypeImage
		}
	}

	for _, ext := range VideoExtensions {
		if extension == ext {
			return FileTypeVideo
		}
	}

	return FileTypeDocument
}

// getDefaultFileName returns default filename based on file type
func (d *FileTypeDetector) getDefaultFileName(fileType string) string {
	switch fileType {
	case FileTypeAudio:
		return DefaultAudioFileName
	case FileTypeImage:
		return DefaultImageFileName
	case FileTypeVideo:
		return DefaultVideoFileName
	default:
		return DefaultDocumentFileName
	}
}

// CacheKeyBuilder provides utilities for building cache keys
type CacheKeyBuilder struct{}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder() *CacheKeyBuilder {
	return &CacheKeyBuilder{}
}

// ContactKey builds cache key for contact
func (b *CacheKeyBuilder) ContactKey(phoneNumber string) string {
	return fmt.Sprintf(CacheKeyContact, phoneNumber)
}

// ConversationKey builds cache key for conversation
func (b *CacheKeyBuilder) ConversationKey(contactID, inboxID int) string {
	return fmt.Sprintf(CacheKeyConversation, contactID, inboxID)
}

// InboxKey builds cache key for inbox
func (b *CacheKeyBuilder) InboxKey(inboxID int) string {
	return fmt.Sprintf(CacheKeyInbox, inboxID)
}

// MimeTypeHelper provides utilities for MIME type handling
type MimeTypeHelper struct{}

// NewMimeTypeHelper creates a new MIME type helper
func NewMimeTypeHelper() *MimeTypeHelper {
	return &MimeTypeHelper{}
}

// GetDefaultMimeType returns default MIME type for file type
func (h *MimeTypeHelper) GetDefaultMimeType(fileType string) string {
	switch fileType {
	case FileTypeAudio:
		return DefaultAudioMimeType
	case FileTypeImage:
		return DefaultImageMimeType
	case FileTypeVideo:
		return DefaultVideoMimeType
	default:
		return DefaultDocumentMimeType
	}
}

// IsValidMimeType checks if MIME type is valid for file type
func (h *MimeTypeHelper) IsValidMimeType(fileType, mimeType string) bool {
	if mimeType == "" {
		return false
	}

	switch fileType {
	case FileTypeAudio:
		return strings.HasPrefix(mimeType, "audio/")
	case FileTypeImage:
		return strings.HasPrefix(mimeType, "image/")
	case FileTypeVideo:
		return strings.HasPrefix(mimeType, "video/")
	default:
		return true
	}
}

// PhoneNumberHelper provides utilities for phone number handling
type PhoneNumberHelper struct{}

// NewPhoneNumberHelper creates a new phone number helper
func NewPhoneNumberHelper() *PhoneNumberHelper {
	return &PhoneNumberHelper{}
}

// ExtractPhoneNumber extracts phone number from contact identifier
func (h *PhoneNumberHelper) ExtractPhoneNumber(identifier string) string {
	if strings.Contains(identifier, "@") {
		return strings.Split(identifier, "@")[0]
	}
	return identifier
}

// FormatPhoneNumber formats phone number for WhatsApp
func (h *PhoneNumberHelper) FormatPhoneNumber(phone string) string {
	phone = strings.TrimPrefix(phone, "+")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	return phone
}

// ErrorHelper provides utilities for error handling
type ErrorHelper struct{}

// NewErrorHelper creates a new error helper
func NewErrorHelper() *ErrorHelper {
	return &ErrorHelper{}
}

// WrapError wraps an error with context
func (h *ErrorHelper) WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// NewError creates a new error with context
func (h *ErrorHelper) NewError(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}
