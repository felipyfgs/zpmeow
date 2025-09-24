package chatwoot

import (
	"fmt"
	"net/http"
	"strings"

	"zpmeow/internal/application/ports"
)

// ErrorHandler implementa a interface ChatwootErrorHandler
type ErrorHandler struct{}

// NewErrorHandler cria um novo manipulador de erros
func NewErrorHandler() ports.ChatwootErrorHandler {
	return &ErrorHandler{}
}

// HandleContactError trata erros relacionados a contatos
func (eh *ErrorHandler) HandleContactError(err error, phoneNumber string) error {
	if err == nil {
		return nil
	}

	// Verifica se é erro de contato duplicado
	if eh.isContactDuplicateError(err) {
		return &ContactDuplicateError{
			PhoneNumber: phoneNumber,
			OriginalErr: err,
		}
	}

	// Verifica se é erro de contato não encontrado
	if eh.isContactNotFoundError(err) {
		return &ContactNotFoundError{
			PhoneNumber: phoneNumber,
			OriginalErr: err,
		}
	}

	return eh.WrapError(err, "contact_operation", map[string]interface{}{
		"phone_number": phoneNumber,
	})
}

// HandleMessageError trata erros relacionados a mensagens
func (eh *ErrorHandler) HandleMessageError(err error, messageID string) error {
	if err == nil {
		return nil
	}

	// Verifica se é erro de mensagem não encontrada
	if eh.isMessageNotFoundError(err) {
		return &MessageNotFoundError{
			MessageID:   messageID,
			OriginalErr: err,
		}
	}

	return eh.WrapError(err, "message_operation", map[string]interface{}{
		"message_id": messageID,
	})
}

// HandleConversationError trata erros relacionados a conversas
func (eh *ErrorHandler) HandleConversationError(err error, conversationID int) error {
	if err == nil {
		return nil
	}

	// Verifica se é erro de conversa não encontrada
	if eh.isConversationNotFoundError(err) {
		return &ConversationNotFoundError{
			ConversationID: conversationID,
			OriginalErr:    err,
		}
	}

	return eh.WrapError(err, "conversation_operation", map[string]interface{}{
		"conversation_id": conversationID,
	})
}

// WrapError encapsula um erro com contexto adicional
func (eh *ErrorHandler) WrapError(err error, operation string, context map[string]interface{}) error {
	return &ChatwootError{
		Operation:   operation,
		Context:     context,
		OriginalErr: err,
	}
}

// Error type definitions
type ChatwootError struct {
	Operation   string                 `json:"operation"`
	Context     map[string]interface{} `json:"context"`
	OriginalErr error                  `json:"original_error"`
}

func (e *ChatwootError) Error() string {
	return fmt.Sprintf("chatwoot %s failed: %v (context: %+v)", e.Operation, e.OriginalErr, e.Context)
}

func (e *ChatwootError) Unwrap() error {
	return e.OriginalErr
}

type ContactDuplicateError struct {
	PhoneNumber string `json:"phone_number"`
	OriginalErr error  `json:"original_error"`
}

func (e *ContactDuplicateError) Error() string {
	return fmt.Sprintf("contact with phone number %s already exists: %v", e.PhoneNumber, e.OriginalErr)
}

func (e *ContactDuplicateError) Unwrap() error {
	return e.OriginalErr
}

type ContactNotFoundError struct {
	PhoneNumber string `json:"phone_number"`
	OriginalErr error  `json:"original_error"`
}

func (e *ContactNotFoundError) Error() string {
	return fmt.Sprintf("contact with phone number %s not found: %v", e.PhoneNumber, e.OriginalErr)
}

func (e *ContactNotFoundError) Unwrap() error {
	return e.OriginalErr
}

type MessageNotFoundError struct {
	MessageID   string `json:"message_id"`
	OriginalErr error  `json:"original_error"`
}

func (e *MessageNotFoundError) Error() string {
	return fmt.Sprintf("message with ID %s not found: %v", e.MessageID, e.OriginalErr)
}

func (e *MessageNotFoundError) Unwrap() error {
	return e.OriginalErr
}

type ConversationNotFoundError struct {
	ConversationID int   `json:"conversation_id"`
	OriginalErr    error `json:"original_error"`
}

func (e *ConversationNotFoundError) Error() string {
	return fmt.Sprintf("conversation with ID %d not found: %v", e.ConversationID, e.OriginalErr)
}

func (e *ConversationNotFoundError) Unwrap() error {
	return e.OriginalErr
}

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Endpoint   string `json:"endpoint"`
	Method     string `json:"method"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d on %s %s: %s", e.StatusCode, e.Method, e.Endpoint, e.Message)
}

// Error detection helpers
func (eh *ErrorHandler) isContactDuplicateError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "already been taken") ||
		strings.Contains(errStr, "identifier has already been taken") ||
		strings.Contains(errStr, "duplicate")
}

func (eh *ErrorHandler) isContactNotFoundError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "contact not found") ||
		strings.Contains(errStr, "resource could not be found")
}

func (eh *ErrorHandler) isMessageNotFoundError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "message not found") ||
		strings.Contains(errStr, "resource could not be found")
}

func (eh *ErrorHandler) isConversationNotFoundError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "conversation not found") ||
		strings.Contains(errStr, "resource could not be found")
}

// HTTP error helpers
func NewAPIError(statusCode int, message, endpoint, method string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Endpoint:   endpoint,
		Method:     method,
	}
}

func IsHTTPError(err error, statusCode int) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == statusCode
	}
	return false
}

func IsNotFoundError(err error) bool {
	return IsHTTPError(err, http.StatusNotFound)
}

func IsUnauthorizedError(err error) bool {
	return IsHTTPError(err, http.StatusUnauthorized)
}

func IsBadRequestError(err error) bool {
	return IsHTTPError(err, http.StatusBadRequest)
}

func IsServerError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode >= 500
	}
	return false
}

// Error helper for legacy compatibility
type ErrorHelper struct {
	handler ports.ChatwootErrorHandler
}

func NewErrorHelper() *ErrorHelper {
	return &ErrorHelper{
		handler: NewErrorHandler(),
	}
}

func (eh *ErrorHelper) WrapError(err error, message string) error {
	return eh.handler.WrapError(err, "legacy_operation", map[string]interface{}{
		"message": message,
	})
}

func (eh *ErrorHelper) HandleAPIError(statusCode int, body []byte, endpoint, method string) error {
	message := string(body)
	if len(message) == 0 {
		message = http.StatusText(statusCode)
	}
	return NewAPIError(statusCode, message, endpoint, method)
}
