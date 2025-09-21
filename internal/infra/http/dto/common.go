package dto

import (
	"net/http"
	"time"
)

const (
	StatusOK                  = http.StatusOK
	StatusCreated             = http.StatusCreated
	StatusBadRequest          = http.StatusBadRequest
	StatusUnauthorized        = http.StatusUnauthorized
	StatusForbidden           = http.StatusForbidden
	StatusNotFound            = http.StatusNotFound
	StatusConflict            = http.StatusConflict
	StatusInternalServerError = http.StatusInternalServerError
)

const (
	ErrorCodeValidationFailed = "VALIDATION_FAILED"
	ErrorCodeInternalError    = "INTERNAL_ERROR"
	ErrorCodeNotFound         = "NOT_FOUND"
	ErrorCodeUnauthorized     = "UNAUTHORIZED"
	ErrorCodeForbidden        = "FORBIDDEN"
	ErrorCodeConflict         = "CONFLICT"
)

type BaseResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ActionResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Action    string      `json:"action"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func NewSuccessResponse(code int, data interface{}) *BaseResponse {
	return &BaseResponse{
		Success:   true,
		Code:      code,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func NewActionResponse(code int, action string, data interface{}) *ActionResponse {
	return &ActionResponse{
		Success:   true,
		Code:      code,
		Action:    action,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func NewErrorResponse(code int, errorCode, message, details string) *BaseResponse {
	return &BaseResponse{
		Success: false,
		Code:    code,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

func NewActionErrorResponse(code int, action, errorCode, message, details string) *ActionResponse {
	return &ActionResponse{
		Success: false,
		Code:    code,
		Action:  action,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

type Validator interface {
	Validate() error
}

type HealthResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Timestamp time.Time         `json:"timestamp" example:"2023-01-01T12:00:00Z"`
	Version   string            `json:"version,omitempty" example:"1.0.0"`
	Services  map[string]string `json:"services,omitempty"`
}

func NewHealthResponse(status, version string, services map[string]string) *HealthResponse {
	return &HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   version,
		Services:  services,
	}
}

type StandardErrorResponse struct {
	Success   bool      `json:"success"`
	Code      int       `json:"code"`
	Error     ErrorInfo `json:"error"`
	Timestamp time.Time `json:"timestamp"`
}

func NewValidationErrorResponse(details string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Code:    StatusBadRequest,
		Error: ErrorInfo{
			Code:    ErrorCodeValidationFailed,
			Message: "Validation failed",
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

func NewNotFoundErrorResponse(resource string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Code:    StatusNotFound,
		Error: ErrorInfo{
			Code:    ErrorCodeNotFound,
			Message: "Resource not found",
			Details: resource + " not found",
		},
		Timestamp: time.Now(),
	}
}

func NewInternalErrorResponse(details string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Code:    StatusInternalServerError,
		Error: ErrorInfo{
			Code:    ErrorCodeInternalError,
			Message: "Internal server error",
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

func NewUnauthorizedErrorResponse() *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Code:    StatusUnauthorized,
		Error: ErrorInfo{
			Code:    ErrorCodeUnauthorized,
			Message: "Unauthorized access",
			Details: "Authentication required",
		},
		Timestamp: time.Now(),
	}
}

func NewConflictErrorResponse(message, details string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Code:    StatusConflict,
		Error: ErrorInfo{
			Code:    ErrorCodeConflict,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

func NewNotImplementedErrorResponse(feature string) *StandardErrorResponse {
	return &StandardErrorResponse{
		Success: false,
		Code:    http.StatusNotImplemented,
		Error: ErrorInfo{
			Code:    "NOT_IMPLEMENTED",
			Message: "Feature not implemented",
			Details: feature + " is not implemented yet",
		},
		Timestamp: time.Now(),
	}
}

type StandardResponse = BaseResponse
