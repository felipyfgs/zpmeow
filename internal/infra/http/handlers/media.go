package handlers

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gin-gonic/gin"
)

type MediaHandler struct {
	sessionService *application.SessionApp
	wmeowService   wmeow.WameowService
}

func NewMediaHandler(sessionService *application.SessionApp, wmeowService wmeow.WameowService) *MediaHandler {
	return &MediaHandler{
		sessionService: sessionService,
		wmeowService:   wmeowService,
	}
}

func (h *MediaHandler) UploadMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.UploadMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"INVALID_DATA",
			"Invalid base64 data",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	mediaURL, err := h.wmeowService.UploadMedia(ctx, sessionID, data, req.MediaType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"UPLOAD_FAILED",
			"Failed to upload media",
			err.Error(),
		))
		return
	}

	responseData := map[string]interface{}{
		"media_url":  mediaURL,
		"media_type": req.MediaType,
		"size":       len(data),
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "upload", "Media uploaded successfully", responseData))
}

func (h *MediaHandler) GetMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")

	var req dto.GetMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	mediaInfo, err := h.wmeowService.GetMediaInfo(ctx, sessionID, req.MediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"GET_MEDIA_FAILED",
			"Failed to get media info",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "get", "Media retrieved successfully", mediaInfo))
}

func (h *MediaHandler) DownloadMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")
	mediaID := c.Param("mediaId")

	if mediaID == "" {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	data, mimeType, err := h.wmeowService.DownloadMedia(ctx, sessionID, mediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"DOWNLOAD_FAILED",
			"Failed to download media",
			err.Error(),
		))
		return
	}

	c.Header("Content-Type", mimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", mediaID))
	c.Header("Content-Length", fmt.Sprintf("%d", len(data)))

	c.Data(http.StatusOK, mimeType, data)
}

func (h *MediaHandler) DeleteMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")
	mediaID := c.Param("mediaId")

	if mediaID == "" {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	err := h.wmeowService.DeleteMedia(ctx, sessionID, mediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"DELETE_FAILED",
			"Failed to delete media",
			err.Error(),
		))
		return
	}

	responseData := map[string]interface{}{
		"media_id": mediaID,
		"deleted":  true,
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "delete", "Media deleted successfully", responseData))
}

func (h *MediaHandler) ListMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")

	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"INVALID_LIMIT",
			"Invalid limit parameter",
			err.Error(),
		))
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"INVALID_OFFSET",
			"Invalid offset parameter",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	mediaList, err := h.wmeowService.ListMedia(ctx, sessionID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"LIST_FAILED",
			"Failed to list media",
			err.Error(),
		))
		return
	}

	responseData := map[string]interface{}{
		"media":  mediaList,
		"count":  len(mediaList),
		"limit":  limit,
		"offset": offset,
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "list", "Media listed successfully", responseData))
}

func (h *MediaHandler) GetMediaProgress(c *gin.Context) {
	sessionID := c.Param("sessionId")
	mediaID := c.Param("mediaId")

	if mediaID == "" {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	progress, err := h.wmeowService.GetMediaProgress(ctx, sessionID, mediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"GET_PROGRESS_FAILED",
			"Failed to get media progress",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "progress", "Media progress retrieved successfully", progress))
}

func (h *MediaHandler) ConvertMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")
	mediaID := c.Param("mediaId")

	var req dto.ConvertMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	convertedMediaID, err := h.wmeowService.ConvertMedia(ctx, sessionID, mediaID, req.TargetFormat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"CONVERT_FAILED",
			"Failed to convert media",
			err.Error(),
		))
		return
	}

	responseData := map[string]interface{}{
		"original_media_id":  mediaID,
		"converted_media_id": convertedMediaID,
		"target_format":      req.TargetFormat,
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "convert", "Media converted successfully", responseData))
}

func (h *MediaHandler) CompressMedia(c *gin.Context) {
	sessionID := c.Param("sessionId")
	mediaID := c.Param("mediaId")

	var req dto.CompressMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
		return
	}

	ctx := c.Request.Context()
	compressedMediaID, err := h.wmeowService.CompressMedia(ctx, sessionID, mediaID, req.Quality)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"COMPRESS_FAILED",
			"Failed to compress media",
			err.Error(),
		))
		return
	}

	responseData := map[string]interface{}{
		"original_media_id":   mediaID,
		"compressed_media_id": compressedMediaID,
		"quality":             req.Quality,
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "compress", "Media compressed successfully", responseData))
}

func (h *MediaHandler) GetMediaMetadata(c *gin.Context) {
	sessionID := c.Param("sessionId")
	mediaID := c.Param("mediaId")

	if mediaID == "" {
		c.JSON(http.StatusBadRequest, dto.NewMediaErrorResponse(
			http.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
		return
	}

	ctx := c.Request.Context()
	metadata, err := h.wmeowService.GetMediaMetadata(ctx, sessionID, mediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.NewMediaErrorResponse(
			http.StatusInternalServerError,
			"GET_METADATA_FAILED",
			"Failed to get media metadata",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, dto.NewMediaSuccessResponse(sessionID, "metadata", "Media metadata retrieved successfully", metadata))
}
