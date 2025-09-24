package wmeow

import (
	"context"
	"fmt"

	"zpmeow/internal/application/ports"

	"go.mau.fi/whatsmeow"
	waTypes "go.mau.fi/whatsmeow/types"
)

// MediaManager methods - upload/download de mídia

func (m *MeowService) UploadMedia(ctx context.Context, sessionID string, data []byte, mimeType string) (*ports.MediaUploadResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	// Determine media type from MIME type
	var mediaType whatsmeow.MediaType
	switch {
	case mimeType == "image/jpeg" || mimeType == "image/png" || mimeType == "image/gif":
		mediaType = whatsmeow.MediaImage
	case mimeType == "video/mp4" || mimeType == "video/3gpp" || mimeType == "video/quicktime":
		mediaType = whatsmeow.MediaVideo
	case mimeType == "audio/aac" || mimeType == "audio/mp4" || mimeType == "audio/mpeg" || mimeType == "audio/ogg":
		mediaType = whatsmeow.MediaAudio
	case mimeType == "application/pdf" || mimeType == "text/plain":
		mediaType = whatsmeow.MediaDocument
	default:
		mediaType = whatsmeow.MediaDocument
	}

	uploaded, err := client.GetClient().Upload(ctx, data, mediaType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload media: %w", err)
	}

	result := &ports.MediaUploadResult{
		URL:           uploaded.URL,
		MediaKey:      uploaded.MediaKey,
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    uploaded.FileLength,
		MimeType:      mimeType,
	}

	m.logger.Debugf("Uploaded media (%s) for session %s", mimeType, sessionID)
	return result, nil
}

func (m *MeowService) DownloadMedia(ctx context.Context, sessionID, messageID string) ([]byte, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	// This is a simplified implementation
	// In a real implementation, you would need to:
	// 1. Find the message by ID
	// 2. Extract media info from the message
	// 3. Download the media using the extracted info

	m.logger.Warnf("DownloadMedia not fully implemented for message %s in session %s", messageID, sessionID)
	return nil, fmt.Errorf("download media not fully implemented")
}

func (m *MeowService) GetMediaInfo(ctx context.Context, sessionID, messageID string) (*ports.MediaInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	// This is a simplified implementation
	// In a real implementation, you would need to:
	// 1. Find the message by ID
	// 2. Extract media info from the message
	// 3. Return the media information

	m.logger.Warnf("GetMediaInfo not fully implemented for message %s in session %s", messageID, sessionID)
	return nil, fmt.Errorf("get media info not fully implemented")
}
