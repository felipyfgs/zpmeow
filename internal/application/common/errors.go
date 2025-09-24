package common

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrValidationFailed = errors.New("validation failed")
	ErrMissingRequired  = errors.New("missing required field")

	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionAlreadyExists = errors.New("session already exists")
	ErrSessionNotConnected  = errors.New("session not connected")
	ErrSessionInUse         = errors.New("session is in use")

	ErrOperationFailed        = errors.New("operation failed")
	ErrConcurrentModification = errors.New("concurrent modification detected")
	ErrResourceUnavailable    = errors.New("resource unavailable")
)

type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}

type BusinessRuleError struct {
	Rule    string
	Message string
}

func (e BusinessRuleError) Error() string {
	return fmt.Sprintf("business rule violation '%s': %s", e.Rule, e.Message)
}

func NewBusinessRuleError(rule, message string) *BusinessRuleError {
	return &BusinessRuleError{
		Rule:    rule,
		Message: message,
	}
}

type ApplicationError struct {
	Code    string
	Message string
	Cause   error
}

func (e ApplicationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e ApplicationError) Unwrap() error {
	return e.Cause
}

func NewApplicationError(code, message string, cause error) *ApplicationError {
	return &ApplicationError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

func IsBusinessRuleError(err error) bool {
	var businessErr *BusinessRuleError
	return errors.As(err, &businessErr)
}

func IsApplicationError(err error) bool {
	var appErr *ApplicationError
	return errors.As(err, &appErr)
}
