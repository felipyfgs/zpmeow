package dto

import (
	"time"
)

type UploadMediaRequest struct {
	SessionID string `json:"session_id" binding:"required" example:"default"`
	MediaType string `json:"media_type" binding:"required" example:"image"`
	FileName  string `json:"file_name,omitempty" example:"image.jpg"`
}

type GetMediaRequest struct {
	MediaID string `json:"media_id" binding:"required" example:"media_123456789"`
}

type MediaInfo struct {
	MediaID    string    `json:"media_id" example:"media_123456789"`
	SessionID  string    `json:"session_id" example:"default"`
	FileName   string    `json:"file_name" example:"image.jpg"`
	MimeType   string    `json:"mime_type" example:"image/jpeg"`
	Size       int64     `json:"size" example:"1024000"`
	MediaType  string    `json:"media_type" example:"image"`
	Status     string    `json:"status" example:"ready"`
	UploadedAt time.Time `json:"uploaded_at" example:"2023-01-01T00:00:00Z"`
}

type MediaUploadProgress struct {
	MediaID       string  `json:"media_id" example:"media_123456789"`
	Progress      float64 `json:"progress" example:"75.5"`
	Status        string  `json:"status" example:"uploading"`
	BytesUploaded int64   `json:"bytes_uploaded" example:"768000"`
	TotalBytes    int64   `json:"total_bytes" example:"1024000"`
}

type MediaDownloadInfo struct {
	MediaID     string    `json:"media_id" example:"media_123456789"`
	DownloadURL string    `json:"download_url" example:"https://storage.example.com/media_123456789"`
	ExpiresAt   time.Time `json:"expires_at" example:"2023-01-01T01:00:00Z"`
}

type MediaResponse struct {
	Success bool                `json:"success"`
	Code    int                 `json:"code"`
	Data    MediaData           `json:"data"`
	Error   *MediaErrorResponse `json:"error,omitempty"`
}

type MediaData struct {
	MediaID   string               `json:"media_id,omitempty" example:"media_123456789"`
	Action    string               `json:"action" example:"upload"`
	Status    string               `json:"status" example:"success"`
	Timestamp time.Time            `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	MediaInfo *MediaInfo           `json:"media_info,omitempty"`
	Download  *MediaDownloadInfo   `json:"download,omitempty"`
	Progress  *MediaUploadProgress `json:"progress,omitempty"`
}

type MediaErrorResponse struct {
	Code    string `json:"code" example:"INVALID_MEDIA_ID"`
	Message string `json:"message" example:"Invalid media ID format"`
	Details string `json:"details,omitempty" example:"Media ID must be alphanumeric"`
}

func NewMediaSuccessResponse(mediaID, action string, mediaInfo *MediaInfo) *MediaResponse {
	return &MediaResponse{
		Success: true,
		Code:    200,
		Data: MediaData{
			MediaID:   mediaID,
			Action:    action,
			Status:    "success",
			Timestamp: time.Now(),
			MediaInfo: mediaInfo,
		},
	}
}

func NewMediaErrorResponse(code int, errorCode, message, details string) *MediaResponse {
	return &MediaResponse{
		Success: false,
		Code:    code,
		Data: MediaData{
			Status:    "error",
			Timestamp: time.Now(),
		},
		Error: &MediaErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}
