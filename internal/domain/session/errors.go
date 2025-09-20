package session

import "errors"

var (
	ErrInvalidSessionID         = errors.New("invalid session ID")
	ErrInvalidSessionName       = errors.New("invalid session name")
	ErrSessionNameTooShort      = errors.New("session name too short")
	ErrSessionNameTooLong       = errors.New("session name too long")
	ErrInvalidSessionNameChar   = errors.New("session name contains invalid characters")
	ErrInvalidSessionNameFormat = errors.New("invalid session name format")
	ErrInvalidSessionStatus     = errors.New("invalid session status")
	ErrInvalidProxyURL          = errors.New("invalid proxy URL")

	ErrSessionAlreadyExists         = errors.New("session already exists")
	ErrSessionNotFound              = errors.New("session not found")
	ErrSessionAlreadyConnected      = errors.New("session is already connected")
	ErrSessionCannotConnect         = errors.New("session cannot be connected in current state")
	ErrSessionNotConnected          = errors.New("session is not connected")
	ErrSessionCannotDisconnect      = errors.New("session cannot be disconnected in current state")
	ErrSessionCannotDelete          = errors.New("session cannot be deleted in current state")
	ErrInvalidSession               = errors.New("invalid session")
	ErrCannotDeleteConnectedSession = errors.New("cannot delete connected session")

	ErrDeviceAlreadyInUse                    = errors.New("device is already in use by another session")
	ErrSessionCannotBeConnectedWithoutDevice = errors.New("session cannot be connected without a device JID")
	ErrMultipleSessionsWithSameDevice        = errors.New("multiple sessions cannot use the same device")
)

type DomainError struct {
	Code    string
	Message string
	Cause   error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

func NewDomainError(message string) *DomainError {
	return &DomainError{
		Message: message,
	}
}

func NewDomainErrorWithCause(message string, cause error) *DomainError {
	return &DomainError{
		Message: message,
		Cause:   cause,
	}
}
