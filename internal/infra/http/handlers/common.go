package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"zpmeow/internal/infra/http/dto"
	"zpmeow/internal/infra/logging"

	"github.com/gofiber/fiber/v2"
)

type Handler interface {
	RegisterRoutes(app *fiber.App)
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

func (h *BaseHandler) BindAndValidate(c *fiber.Ctx, req interface{}) error {
	if err := c.BodyParser(req); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return h.ValidateRequest(req)
}

type HTTPHandler struct {
	*BaseHandler
}

func (h *BaseHandler) SendSuccessResponse(c *fiber.Ctx, statusCode int, data interface{}) error {
	response := dto.NewSuccessResponse(statusCode, data)
	return c.Status(statusCode).JSON(response)
}

func (h *BaseHandler) SendActionResponse(c *fiber.Ctx, statusCode int, action string, data interface{}) error {
	response := dto.NewActionResponse(statusCode, action, data)
	return c.Status(statusCode).JSON(response)
}

func (h *BaseHandler) SendErrorResponse(c *fiber.Ctx, statusCode int, errorCode, message string, err error) error {
	details := ""
	if err != nil {
		details = err.Error()
	}
	response := dto.NewErrorResponse(statusCode, errorCode, message, details)
	return c.Status(statusCode).JSON(response)
}

func (h *HTTPHandler) SendSuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}) error {
	return h.BaseHandler.SendSuccessResponse(c, statusCode, data)
}

func (h *HTTPHandler) SendErrorResponse(c *fiber.Ctx, statusCode int, message string, err error) error {
	return h.BaseHandler.SendErrorResponse(c, statusCode, dto.ErrorCodeInternalError, message, err)
}

func (h *BaseHandler) SendValidationErrorResponse(c *fiber.Ctx, err error) error {
	return h.SendErrorResponse(c, dto.StatusBadRequest, dto.ErrorCodeValidationFailed, "Validation error", err)
}

func (h *BaseHandler) SendInternalErrorResponse(c *fiber.Ctx, err error) error {
	return h.SendErrorResponse(c, dto.StatusInternalServerError, dto.ErrorCodeInternalError, "Internal server error", err)
}

func (h *BaseHandler) SendNotFoundResponse(c *fiber.Ctx, message string) error {
	return h.SendErrorResponse(c, dto.StatusNotFound, dto.ErrorCodeNotFound, message, nil)
}

func (h *BaseHandler) SendUnauthorizedResponse(c *fiber.Ctx, message string) error {
	return h.SendErrorResponse(c, dto.StatusUnauthorized, dto.ErrorCodeUnauthorized, message, nil)
}

func (h *BaseHandler) SendForbiddenResponse(c *fiber.Ctx, message string) error {
	return h.SendErrorResponse(c, dto.StatusForbidden, dto.ErrorCodeForbidden, message, nil)
}

func (h *BaseHandler) SendConflictResponse(c *fiber.Ctx, message string, err error) error {
	return h.SendErrorResponse(c, dto.StatusConflict, dto.ErrorCodeConflict, message, err)
}

func (h *HTTPHandler) SendValidationErrorResponse(c *fiber.Ctx, err error) error {
	return h.BaseHandler.SendValidationErrorResponse(c, err)
}

func (h *BaseHandler) SendStandardErrorResponse(c *fiber.Ctx, errorResponse *dto.StandardErrorResponse) error {
	return c.Status(errorResponse.Code).JSON(errorResponse)
}

func (h *BaseHandler) SendValidationError(c *fiber.Ctx, details string) error {
	response := dto.NewValidationErrorResponse(details)
	return h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendNotFoundError(c *fiber.Ctx, resource string) error {
	response := dto.NewNotFoundErrorResponse(resource)
	return h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendInternalError(c *fiber.Ctx, details string) error {
	response := dto.NewInternalErrorResponse(details)
	return h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendUnauthorizedError(c *fiber.Ctx) error {
	response := dto.NewUnauthorizedErrorResponse()
	return h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendConflictError(c *fiber.Ctx, message, details string) error {
	response := dto.NewConflictErrorResponse(message, details)
	return h.SendStandardErrorResponse(c, response)
}

func (h *BaseHandler) SendNotImplementedError(c *fiber.Ctx, feature string) error {
	response := dto.NewNotImplementedErrorResponse(feature)
	return h.SendStandardErrorResponse(c, response)
}

func (h *HTTPHandler) SendInternalErrorResponse(c *fiber.Ctx, err error) error {
	return h.BaseHandler.SendInternalErrorResponse(c, err)
}

func (h *HTTPHandler) SendNotFoundResponse(c *fiber.Ctx, message string) error {
	return h.BaseHandler.SendNotFoundResponse(c, message)
}

func (h *HTTPHandler) SendUnauthorizedResponse(c *fiber.Ctx, message string) error {
	return h.BaseHandler.SendUnauthorizedResponse(c, message)
}

func (h *HTTPHandler) SendForbiddenResponse(c *fiber.Ctx, message string) error {
	return h.BaseHandler.SendForbiddenResponse(c, message)
}

func (h *HTTPHandler) SendConflictResponse(c *fiber.Ctx, message string, err error) error {
	return h.BaseHandler.SendConflictResponse(c, message, err)
}

func (h *BaseHandler) GetSessionIDFromPath(c *fiber.Ctx) string {
	return c.Params("sessionId")
}

func (h *BaseHandler) GetQueryParam(c *fiber.Ctx, key, defaultValue string) string {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return defaultValue
	}
	return value
}

func (h *BaseHandler) GetQueryParamInt(c *fiber.Ctx, key string, defaultValue int) int {
	value := strings.TrimSpace(c.Query(key))
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

func (h *HTTPHandler) GetSessionIDFromPath(c *fiber.Ctx) string {
	return h.BaseHandler.GetSessionIDFromPath(c)
}

func (h *HTTPHandler) GetQueryParam(c *fiber.Ctx, key, defaultValue string) string {
	return h.BaseHandler.GetQueryParam(c, key, defaultValue)
}

func (h *HTTPHandler) GetQueryParamInt(c *fiber.Ctx, key string, defaultValue int) int {
	return h.BaseHandler.GetQueryParamInt(c, key, defaultValue)
}

func (h *BaseHandler) BindJSON(c *fiber.Ctx, obj interface{}) error {
	if err := c.BodyParser(obj); err != nil {
		h.logger.Errorf("JSON binding failed: %v", err)
		h.SendValidationErrorResponse(c, err)
		return err
	}
	return nil
}

func (h *BaseHandler) BindQuery(c *fiber.Ctx, obj interface{}) error {
	if err := c.QueryParser(obj); err != nil {
		h.logger.Errorf("Query binding failed: %v", err)
		h.SendValidationErrorResponse(c, err)
		return err
	}
	return nil
}

func (h *BaseHandler) BindURI(c *fiber.Ctx, obj interface{}) error {
	if err := c.ParamsParser(obj); err != nil {
		h.logger.Errorf("URI binding failed: %v", err)
		h.SendValidationErrorResponse(c, err)
		return err
	}
	return nil
}

func (h *HTTPHandler) BindJSON(c *fiber.Ctx, obj interface{}) error {
	return h.BaseHandler.BindJSON(c, obj)
}

func (h *HTTPHandler) BindQuery(c *fiber.Ctx, obj interface{}) error {
	return h.BaseHandler.BindQuery(c, obj)
}

func (h *HTTPHandler) BindURI(c *fiber.Ctx, obj interface{}) error {
	return h.BaseHandler.BindURI(c, obj)
}
