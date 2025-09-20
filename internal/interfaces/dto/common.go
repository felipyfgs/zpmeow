package dto

import (
	"time"
)

type StandardResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorInfo struct {
	Code    string `json:"code" example:"INVALID_REQUEST"`
	Message string `json:"message" example:"Invalid request parameters"`
	Details string `json:"details,omitempty" example:"Additional error details"`
}

type PaginationInfo struct {
	Page       int `json:"page" example:"1"`
	PageSize   int `json:"page_size" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
}

type ActionData struct {
	Action    string    `json:"action" example:"create_session"`
	Status    string    `json:"status" example:"success"`
	Timestamp time.Time `json:"timestamp" example:"2023-01-01T00:00:00Z"`
}

func NewSuccessResponse(code int, data interface{}) *StandardResponse {
	return &StandardResponse{
		Success:   true,
		Code:      code,
		Data:      data,
		Timestamp: time.Now(),
	}
}

func NewErrorResponse(code int, errorCode, message, details string) *StandardResponse {
	return &StandardResponse{
		Success: false,
		Code:    code,
		Data:    nil,
		Error: &ErrorInfo{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now(),
	}
}

func NewActionResponse(code int, action string, data interface{}) *StandardResponse {
	actionData := ActionData{
		Action:    action,
		Status:    "success",
		Timestamp: time.Now(),
	}

	responseData := map[string]interface{}{
		"action":    actionData.Action,
		"status":    actionData.Status,
		"timestamp": actionData.Timestamp,
	}

	if data != nil {
		responseData["result"] = data
	}

	return &StandardResponse{
		Success:   true,
		Code:      code,
		Data:      responseData,
		Timestamp: time.Now(),
	}
}

// Constantes de status HTTP movidas para internal/interfaces/dto/errors.go

// Constantes movidas para internal/interfaces/dto/errors.go para evitar duplicação
