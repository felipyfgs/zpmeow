package dto

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)


type UploadMediaRequest struct {
	MediaType string `json:"media_type" binding:"required" example:"image"`
	Data      string `json:"data" binding:"required" example:"base64_encoded_data"`
	FileName  string `json:"filename,omitempty" example:"image.jpg"`
	Caption   string `json:"caption,omitempty" example:"Check this out!"`
}

func (r UploadMediaRequest) Validate() error {
	if strings.TrimSpace(r.MediaType) == "" {
		return fmt.Errorf("media_type is required")
	}
	if strings.TrimSpace(r.Data) == "" {
		return fmt.Errorf("data is required")
	}

	validTypes := []string{"image", "audio", "video", "document", "sticker"}
	for _, validType := range validTypes {
		if r.MediaType == validType {
			return nil
		}
	}
	return fmt.Errorf("invalid media_type, must be one of: %s", strings.Join(validTypes, ", "))
}

type DownloadMediaRequest struct {
	MediaURL  string `json:"media_url" binding:"required" example:"https://example.com/media.jpg"`
	MessageID string `json:"message_id,omitempty" example:"msg_123"`
}

func (r DownloadMediaRequest) Validate() error {
	if strings.TrimSpace(r.MediaURL) == "" {
		return fmt.Errorf("media_url is required")
	}
	if !strings.HasPrefix(r.MediaURL, "http://") && !strings.HasPrefix(r.MediaURL, "https://") {
		return fmt.Errorf("media_url must be a valid HTTP or HTTPS URL")
	}
	return nil
}

type ConvertMediaRequest struct {
	TargetFormat string `json:"target_format" binding:"required" example:"jpeg"`
	Quality      int    `json:"quality,omitempty" example:"80"`
}

func (r ConvertMediaRequest) Validate() error {
	if strings.TrimSpace(r.TargetFormat) == "" {
		return fmt.Errorf("target_format is required")
	}
	if r.Quality < 0 || r.Quality > 100 {
		return fmt.Errorf("quality must be between 0 and 100")
	}
	return nil
}

type CompressMediaRequest struct {
	Quality int `json:"quality" binding:"required" example:"80"`
}

func (r CompressMediaRequest) Validate() error {
	if r.Quality < 1 || r.Quality > 100 {
		return fmt.Errorf("quality must be between 1 and 100")
	}
	return nil
}


type MediaErrorResponse struct {
	Code    string `json:"code" example:"MEDIA_UPLOAD_FAILED"`
	Message string `json:"message" example:"Failed to upload media"`
	Details string `json:"details" example:"Invalid media format"`
}

type MediaInfo struct {
	ID        string    `json:"id" example:"media_123"`
	URL       string    `json:"url" example:"https://example.com/media.jpg"`
	Type      string    `json:"type" example:"image"`
	MimeType  string    `json:"mime_type" example:"image/jpeg"`
	Size      int64     `json:"size" example:"1024000"`
	FileName  string    `json:"filename,omitempty" example:"image.jpg"`
	Caption   string    `json:"caption,omitempty" example:"Check this out!"`
	Width     int       `json:"width,omitempty" example:"1920"`
	Height    int       `json:"height,omitempty" example:"1080"`
	Duration  int       `json:"duration,omitempty" example:"30"`
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`
}

type MediaResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *MediaResponseData  `json:"data,omitempty"`
	Error   *MediaErrorResponse `json:"error,omitempty"`
}

type MediaResponseData struct {
	SessionID string      `json:"session_id,omitempty"`
	Action    string      `json:"action"`
	Status    string      `json:"status"`
	Message   string      `json:"message,omitempty"`
	Media     *MediaInfo  `json:"media,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type MediaListResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *MediaListData      `json:"data,omitempty"`
	Error   *MediaErrorResponse `json:"error,omitempty"`
}

type MediaListData struct {
	SessionID string      `json:"session_id,omitempty"`
	Media     []MediaInfo `json:"media"`
	Count     int         `json:"count"`
	Total     int         `json:"total"`
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
}

type MediaDownloadResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *MediaDownloadData  `json:"data,omitempty"`
	Error   *MediaErrorResponse `json:"error,omitempty"`
}

type MediaDownloadData struct {
	MediaID   string    `json:"media_id,omitempty"`
	URL       string    `json:"url"`
	Type      string    `json:"type"`
	MimeType  string    `json:"mime_type"`
	Size      int64     `json:"size"`
	FileName  string    `json:"filename,omitempty"`
	Data      []byte    `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type MediaConvertResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    *MediaConvertData   `json:"data,omitempty"`
	Error   *MediaErrorResponse `json:"error,omitempty"`
}

type MediaConvertData struct {
	OriginalURL   string    `json:"original_url"`
	ConvertedURL  string    `json:"converted_url"`
	FromFormat    string    `json:"from_format"`
	ToFormat      string    `json:"to_format"`
	Quality       int       `json:"quality,omitempty"`
	OriginalSize  int64     `json:"original_size"`
	ConvertedSize int64     `json:"converted_size"`
	Timestamp     time.Time `json:"timestamp"`
}


func NewMediaErrorResponse(code int, errorCode, message, details string) *MediaResponse {
	return &MediaResponse{
		Success: false,
		Code:    code,
		Error: &MediaErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewMediaSuccessResponse(sessionID, action, message string, data interface{}) *MediaResponse {
	responseData := &MediaResponseData{
		SessionID: sessionID,
		Action:    action,
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
	}

	if mediaInfo, ok := data.(*MediaInfo); ok {
		responseData.Media = mediaInfo
	}

	return &MediaResponse{
		Success: true,
		Code:    http.StatusOK,
		Data:    responseData,
	}
}

func NewMediaListSuccessResponse(sessionID string, media []MediaInfo, limit, offset, total int) *MediaListResponse {
	return &MediaListResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &MediaListData{
			SessionID: sessionID,
			Media:     media,
			Count:     len(media),
			Total:     total,
			Limit:     limit,
			Offset:    offset,
		},
	}
}

func NewMediaListErrorResponse(code int, errorCode, message, details string) *MediaListResponse {
	return &MediaListResponse{
		Success: false,
		Code:    code,
		Error: &MediaErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewMediaDownloadSuccessResponse(url, mediaType, mimeType, fileName string, size int64, data []byte) *MediaDownloadResponse {
	return &MediaDownloadResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &MediaDownloadData{
			URL:       url,
			Type:      mediaType,
			MimeType:  mimeType,
			Size:      size,
			FileName:  fileName,
			Data:      data,
			Timestamp: time.Now(),
		},
	}
}

func NewMediaDownloadErrorResponse(code int, errorCode, message, details string) *MediaDownloadResponse {
	return &MediaDownloadResponse{
		Success: false,
		Code:    code,
		Error: &MediaErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewMediaConvertSuccessResponse(originalURL, convertedURL, fromFormat, toFormat string, quality int, originalSize, convertedSize int64) *MediaConvertResponse {
	return &MediaConvertResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &MediaConvertData{
			OriginalURL:   originalURL,
			ConvertedURL:  convertedURL,
			FromFormat:    fromFormat,
			ToFormat:      toFormat,
			Quality:       quality,
			OriginalSize:  originalSize,
			ConvertedSize: convertedSize,
			Timestamp:     time.Now(),
		},
	}
}

func NewMediaConvertErrorResponse(code int, errorCode, message, details string) *MediaConvertResponse {
	return &MediaConvertResponse{
		Success: false,
		Code:    code,
		Error: &MediaErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

type GetMediaRequest struct {
	MediaID string `json:"media_id" binding:"required" example:"media123"`
}

func (r GetMediaRequest) Validate() error {
	if strings.TrimSpace(r.MediaID) == "" {
		return fmt.Errorf("media_id is required")
	}
	return nil
}
