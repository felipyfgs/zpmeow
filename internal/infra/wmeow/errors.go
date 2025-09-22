package wmeow

// Validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// Connection errors
type ConnectionError struct {
	SessionID string
	Operation string
	Cause     error
}

func (e ConnectionError) Error() string {
	return "connection error for session " + e.SessionID + " during " + e.Operation + ": " + e.Cause.Error()
}

func (e ConnectionError) Unwrap() error {
	return e.Cause
}

func NewConnectionError(sessionID, operation string, cause error) *ConnectionError {
	return &ConnectionError{
		SessionID: sessionID,
		Operation: operation,
		Cause:     cause,
	}
}
