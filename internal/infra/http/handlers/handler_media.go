package handlers

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"zpmeow/internal/application"
	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
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

func (h *MediaHandler) UploadMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.UploadMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	data, err := base64.StdEncoding.DecodeString(req.Data)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_DATA",
			"Invalid base64 data",
			err.Error(),
		))
	}

	ctx := c.Context()
	mediaURL, err := h.wmeowService.UploadMedia(ctx, sessionID, data, req.MediaType)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"UPLOAD_FAILED",
			"Failed to upload media",
			err.Error(),
		))
	}

	responseData := map[string]interface{}{
		"media_url":  mediaURL,
		"media_type": req.MediaType,
		"size":       len(data),
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "upload", "Media uploaded successfully", responseData))
}

func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	var req dto.GetMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	mediaInfo, err := h.wmeowService.GetMediaInfo(ctx, sessionID, req.MediaID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"GET_MEDIA_FAILED",
			"Failed to get media info",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "get", "Media retrieved successfully", mediaInfo))
}

func (h *MediaHandler) DownloadMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	mediaID := c.Params("mediaId")

	if mediaID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
	}

	ctx := c.Context()
	data, mimeType, err := h.wmeowService.DownloadMedia(ctx, sessionID, mediaID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"DOWNLOAD_FAILED",
			"Failed to download media",
			err.Error(),
		))
	}

	c.Set("Content-Type", mimeType)
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", mediaID))
	c.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	return c.Status(fiber.StatusOK).Send(data)
}

func (h *MediaHandler) DeleteMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	mediaID := c.Params("mediaId")

	if mediaID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
	}

	ctx := c.Context()
	err := h.wmeowService.DeleteMedia(ctx, sessionID, mediaID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"DELETE_FAILED",
			"Failed to delete media",
			err.Error(),
		))
	}

	responseData := map[string]interface{}{
		"media_id": mediaID,
		"deleted":  true,
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "delete", "Media deleted successfully", responseData))
}

func (h *MediaHandler) ListMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")

	limitStr := c.Query("limit", "50")
	offsetStr := c.Query("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_LIMIT",
			"Invalid limit parameter",
			err.Error(),
		))
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_OFFSET",
			"Invalid offset parameter",
			err.Error(),
		))
	}

	ctx := c.Context()
	mediaList, err := h.wmeowService.ListMedia(ctx, sessionID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"LIST_FAILED",
			"Failed to list media",
			err.Error(),
		))
	}

	responseData := map[string]interface{}{
		"media":  mediaList,
		"count":  len(mediaList),
		"limit":  limit,
		"offset": offset,
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "list", "Media listed successfully", responseData))
}

func (h *MediaHandler) GetMediaProgress(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	mediaID := c.Params("mediaId")

	if mediaID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
	}

	ctx := c.Context()
	progress, err := h.wmeowService.GetMediaProgress(ctx, sessionID, mediaID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"GET_PROGRESS_FAILED",
			"Failed to get media progress",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "progress", "Media progress retrieved successfully", progress))
}

func (h *MediaHandler) ConvertMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	mediaID := c.Params("mediaId")

	var req dto.ConvertMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	convertedMediaID, err := h.wmeowService.ConvertMedia(ctx, sessionID, mediaID, req.TargetFormat)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"CONVERT_FAILED",
			"Failed to convert media",
			err.Error(),
		))
	}

	responseData := map[string]interface{}{
		"original_media_id":  mediaID,
		"converted_media_id": convertedMediaID,
		"target_format":      req.TargetFormat,
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "convert", "Media converted successfully", responseData))
}

func (h *MediaHandler) CompressMedia(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	mediaID := c.Params("mediaId")

	var req dto.CompressMediaRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"INVALID_REQUEST",
			"Invalid request format",
			err.Error(),
		))
	}

	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"VALIDATION_ERROR",
			"Request validation failed",
			err.Error(),
		))
	}

	ctx := c.Context()
	compressedMediaID, err := h.wmeowService.CompressMedia(ctx, sessionID, mediaID, req.Quality)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"COMPRESS_FAILED",
			"Failed to compress media",
			err.Error(),
		))
	}

	responseData := map[string]interface{}{
		"original_media_id":   mediaID,
		"compressed_media_id": compressedMediaID,
		"quality":             req.Quality,
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "compress", "Media compressed successfully", responseData))
}

func (h *MediaHandler) GetMediaMetadata(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	mediaID := c.Params("mediaId")

	if mediaID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(dto.NewMediaErrorResponse(
			fiber.StatusBadRequest,
			"MISSING_MEDIA_ID",
			"Media ID is required",
			"",
		))
	}

	ctx := c.Context()
	metadata, err := h.wmeowService.GetMediaMetadata(ctx, sessionID, mediaID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(dto.NewMediaErrorResponse(
			fiber.StatusInternalServerError,
			"GET_METADATA_FAILED",
			"Failed to get media metadata",
			err.Error(),
		))
	}

	return c.Status(fiber.StatusOK).JSON(dto.NewMediaSuccessResponse(sessionID, "metadata", "Media metadata retrieved successfully", metadata))
}
