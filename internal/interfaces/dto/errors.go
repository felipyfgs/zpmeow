package dto

import (
	"net/http"
	"time"
)

// CommonErrorResponse representa uma resposta de erro padronizada
type CommonErrorResponse struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"Invalid request parameters"`
	Details string `json:"details,omitempty" example:"Additional error details"`
}

// StandardErrorResponse representa uma resposta de erro completa
type StandardErrorResponse struct {
	Success   bool                 `json:"success"`
	Code      int                  `json:"code"`
	Data      interface{}          `json:"data,omitempty"`
	Error     *CommonErrorResponse `json:"error,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
}

// Códigos de erro padronizados
const (
	// HTTP Status Codes
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusConflict            = http.StatusConflict
	StatusInternalServerError = http.StatusInternalServerError
	StatusNotImplemented      = http.StatusNotImplemented

	// Error Codes
	ErrorCodeInvalidRequest     = "INVALID_REQUEST"
	ErrorCodeValidationFailed   = "VALIDATION_FAILED"
	ErrorCodeUnauthorized       = "UNAUTHORIZED"
	ErrorCodeForbidden          = "FORBIDDEN"
	ErrorCodeNotFound           = "NOT_FOUND"
	ErrorCodeConflict           = "CONFLICT"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
	ErrorCodeNotImplemented     = "NOT_IMPLEMENTED"
	ErrorCodeSessionNotFound    = "SESSION_NOT_FOUND"
	ErrorCodeSessionInactive    = "SESSION_INACTIVE"
	ErrorCodeInvalidPhoneNumber = "INVALID_PHONE_NUMBER"
	ErrorCodeInvalidMessageID   = "INVALID_MESSAGE_ID"
	ErrorCodeInvalidURL         = "INVALID_URL"
)

// NewStandardErrorResponse cria uma resposta de erro padronizada
func NewStandardErrorResponse(httpCode int, errorCode, message, details string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Code:    httpCode,
		Error: &CommonErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

// NewValidationErrorResponse cria uma resposta de erro de validação
func NewValidationErrorResponse(details string) *StandardErrorResponse {
	return NewStandardErrorResponse(
		StatusBadRequest,
		ErrorCodeValidationFailed,
		"Request validation failed",
		details,
	)
}

// NewNotFoundErrorResponse cria uma resposta de erro de recurso não encontrado
func NewNotFoundErrorResponse(resource string) *StandardErrorResponse {
	return NewStandardErrorResponse(
		StatusNotFound,
		ErrorCodeNotFound,
		resource+" not found",
		"",
	)
}

// NewInternalErrorResponse cria uma resposta de erro interno
func NewInternalErrorResponse(details string) *StandardErrorResponse {
	return NewStandardErrorResponse(
		StatusInternalServerError,
		ErrorCodeInternalError,
		"Internal server error",
		details,
	)
}

// NewUnauthorizedErrorResponse cria uma resposta de erro de autorização
func NewUnauthorizedErrorResponse() *StandardErrorResponse {
	return NewStandardErrorResponse(
		StatusUnauthorized,
		ErrorCodeUnauthorized,
		"Unauthorized access",
		"Valid API key required",
	)
}

// NewConflictErrorResponse cria uma resposta de erro de conflito
func NewConflictErrorResponse(message, details string) *StandardErrorResponse {
	return NewStandardErrorResponse(
		StatusConflict,
		ErrorCodeConflict,
		message,
		details,
	)
}

// NewNotImplementedErrorResponse cria uma resposta de erro de funcionalidade não implementada
func NewNotImplementedErrorResponse(feature string) *StandardErrorResponse {
	return NewStandardErrorResponse(
		StatusNotImplemented,
		ErrorCodeNotImplemented,
		feature+" not yet implemented",
		"This feature is planned for future releases",
	)
}
