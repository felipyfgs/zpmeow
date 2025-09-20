package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"zpmeow/internal/infra/logging"
	"zpmeow/internal/interfaces/dto"
)

type Handler interface {
	RegisterRoutes(router *gin.Engine)
}

type BaseHandler struct {
	logger logging.Logger
}

func NewBaseHandler(moduleName string) *BaseHandler {
	return &BaseHandler{
		logger: logging.GetLogger().Sub(moduleName),
	}
}

func (h *BaseHandler) ValidateRequest(req interface{}) error {
	if validator, ok := req.(interface{ Validate() error }); ok {
		return validator.Validate()
	}

	return nil
}

func (h *BaseHandler) BindAndValidate(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(req); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return h.ValidateRequest(req)
}

type HTTPHandler struct {
	*BaseHandler
}

func (h *BaseHandler) SendSuccessResponse(c *gin.Context, statusCode int, data interface{}) {
	response := dto.NewSuccessResponse(statusCode, data)
	c.JSON(statusCode, response)
}

func (h *BaseHandler) SendActionResponse(c *gin.Context, statusCode int, action string, data interface{}) {
	response := dto.NewActionResponse(statusCode, action, data)
	c.JSON(statusCode, response)
}

func (h *BaseHandler) SendErrorResponse(c *gin.Context, statusCode int, errorCode, message string, err error) {
	details := ""
	if err != nil {
		details = err.Error()
	}
	response := dto.NewErrorResponse(statusCode, errorCode, message, details)
	c.JSON(statusCode, response)
}

func (h *HTTPHandler) SendSuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	h.BaseHandler.SendSuccessResponse(c, statusCode, data)
}

func (h *HTTPHandler) SendErrorResponse(c *gin.Context, statusCode int, message string, err error) {
	h.BaseHandler.SendErrorResponse(c, statusCode, dto.ErrorCodeInternalError, message, err)
}

func (h *BaseHandler) SendValidationErrorResponse(c *gin.Context, err error) {
	h.SendErrorResponse(c, dto.StatusBadRequest, dto.ErrorCodeValidationFailed, "Validation error", err)
}

func (h *BaseHandler) SendInternalErrorResponse(c *gin.Context, err error) {
	h.SendErrorResponse(c, dto.StatusInternalServerError, dto.ErrorCodeInternalError, "Internal server error", err)
}

func (h *BaseHandler) SendNotFoundResponse(c *gin.Context, message string) {
	h.SendErrorResponse(c, dto.StatusNotFound, dto.ErrorCodeNotFound, message, nil)
}

func (h *BaseHandler) SendUnauthorizedResponse(c *gin.Context, message string) {
	h.SendErrorResponse(c, dto.StatusUnauthorized, dto.ErrorCodeUnauthorized, message, nil)
}

func (h *BaseHandler) SendForbiddenResponse(c *gin.Context, message string) {
	h.SendErrorResponse(c, dto.StatusForbidden, dto.ErrorCodeForbidden, message, nil)
}

func (h *BaseHandler) SendConflictResponse(c *gin.Context, message string, err error) {
	h.SendErrorResponse(c, dto.StatusConflict, dto.ErrorCodeConflict, message, err)
}

func (h *HTTPHandler) SendValidationErrorResponse(c *gin.Context, err error) {
	h.BaseHandler.SendValidationErrorResponse(c, err)
}

// Métodos de conveniência usando as novas estruturas padronizadas
func (h *BaseHandler) SendStandardErrorResponse(c *gin.Context, errorResponse *dto.StandardErrorResponse) {
	c.JSON(errorResponse.Code, errorResponse)
}

func (h *BaseHandler) SendValidationError(c *gin.Context, details string) {
	response := dto.NewValidationErrorResponse(details)
	h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendNotFoundError(c *gin.Context, resource string) {
	response := dto.NewNotFoundErrorResponse(resource)
	h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendInternalError(c *gin.Context, details string) {
	response := dto.NewInternalErrorResponse(details)
	h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendUnauthorizedError(c *gin.Context) {
	response := dto.NewUnauthorizedErrorResponse()
	h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendConflictError(c *gin.Context, message, details string) {
	response := dto.NewConflictErrorResponse(message, details)
	h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendNotImplementedError(c *gin.Context, feature string) {
	response := dto.NewNotImplementedErrorResponse(feature)
	h.SendStandardErrorResponse(c, response)
}

func (h *HTTPHandler) SendInternalErrorResponse(c *gin.Context, err error) {
	h.BaseHandler.SendInternalErrorResponse(c, err)
}

func (h *HTTPHandler) SendNotFoundResponse(c *gin.Context, message string) {
	h.BaseHandler.SendNotFoundResponse(c, message)
}

func (h *HTTPHandler) SendUnauthorizedResponse(c *gin.Context, message string) {
	h.BaseHandler.SendUnauthorizedResponse(c, message)
}

func (h *HTTPHandler) SendForbiddenResponse(c *gin.Context, message string) {
	h.BaseHandler.SendForbiddenResponse(c, message)
}

func (h *HTTPHandler) SendConflictResponse(c *gin.Context, message string, err error) {
	h.BaseHandler.SendConflictResponse(c, message, err)
}

func (h *BaseHandler) GetSessionIDFromPath(c *gin.Context) string {
	return c.Param("sessionId")
}

func (h *BaseHandler) GetQueryParam(c *gin.Context, key, defaultValue string) string {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func (h *BaseHandler) GetQueryParamInt(c *gin.Context, key string, defaultValue int) int {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

func (h *HTTPHandler) GetSessionIDFromPath(c *gin.Context) string {
	return h.BaseHandler.GetSessionIDFromPath(c)
}

func (h *HTTPHandler) GetQueryParam(c *gin.Context, key, defaultValue string) string {
	return h.BaseHandler.GetQueryParam(c, key, defaultValue)
}

func (h *HTTPHandler) GetQueryParamInt(c *gin.Context, key string, defaultValue int) int {
	return h.BaseHandler.GetQueryParamInt(c, key, defaultValue)
}

func (h *BaseHandler) BindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		h.logger.Errorf("JSON binding failed: %v", err)
		h.SendValidationErrorResponse(c, err)
		return err
	}
	return nil
}

func (h *BaseHandler) BindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		h.logger.Errorf("Query binding failed: %v", err)
		h.SendValidationErrorResponse(c, err)
		return err
	}
	return nil
}

func (h *BaseHandler) BindURI(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindUri(obj); err != nil {
		h.logger.Errorf("URI binding failed: %v", err)
		h.SendValidationErrorResponse(c, err)
		return err
	}
	return nil
}

func (h *HTTPHandler) BindJSON(c *gin.Context, obj interface{}) error {
	return h.BaseHandler.BindJSON(c, obj)
}

func (h *HTTPHandler) BindQuery(c *gin.Context, obj interface{}) error {
	return h.BaseHandler.BindQuery(c, obj)
}

func (h *HTTPHandler) BindURI(c *gin.Context, obj interface{}) error {
	return h.BaseHandler.BindURI(c, obj)
}
