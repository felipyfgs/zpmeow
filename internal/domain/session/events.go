package session

import (
	"zpmeow/internal/domain/common"
)

const (
	SessionCreatedEventType              = "session.created"
	SessionConnectedEventType            = "session.connected"
	SessionDisconnectedEventType         = "session.disconnected"
	SessionAuthenticatedEventType        = "session.authenticated"
	SessionConfigurationChangedEventType = "session.configuration_changed"
	SessionDeletedEventType              = "session.deleted"
	SessionErrorEventType                = "session.error"
)

type SessionCreatedEvent struct {
	common.BaseDomainEvent
}

func NewSessionCreatedEvent(sessionID string, sessionName string) SessionCreatedEvent {
	data := map[string]interface{}{
		"session_id":   sessionID,
		"session_name": sessionName,
	}

	return SessionCreatedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(
			SessionCreatedEventType,
			sessionID,
			data,
		),
	}
}

type SessionConnectedEvent struct {
	common.BaseDomainEvent
}

func NewSessionConnectedEvent(sessionID string, DeviceJID string) SessionConnectedEvent {
	data := map[string]interface{}{
		"session_id": sessionID,
		"device_jid": DeviceJID,
	}

	return SessionConnectedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(
			SessionConnectedEventType,
			sessionID,
			data,
		),
	}
}

type SessionDisconnectedEvent struct {
	common.BaseDomainEvent
}

func NewSessionDisconnectedEvent(sessionID string, reason string) SessionDisconnectedEvent {
	data := map[string]interface{}{
		"session_id": sessionID,
		"reason":     reason,
	}

	return SessionDisconnectedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(
			SessionDisconnectedEventType,
			sessionID,
			data,
		),
	}
}

type SessionAuthenticatedEvent struct {
	common.BaseDomainEvent
}

func NewSessionAuthenticatedEvent(sessionID string, DeviceJID string) SessionAuthenticatedEvent {
	data := map[string]interface{}{
		"session_id": sessionID,
		"device_jid": DeviceJID,
	}

	return SessionAuthenticatedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(
			SessionAuthenticatedEventType,
			sessionID,
			data,
		),
	}
}

type SessionConfigurationChangedEvent struct {
	common.BaseDomainEvent
}

func NewSessionConfigurationChangedEvent(sessionID string, changes map[string]interface{}) SessionConfigurationChangedEvent {
	data := map[string]interface{}{
		"session_id": sessionID,
		"changes":    changes,
	}

	return SessionConfigurationChangedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(
			SessionConfigurationChangedEventType,
			sessionID,
			data,
		),
	}
}

type SessionDeletedEvent struct {
	common.BaseDomainEvent
}

func NewSessionDeletedEvent(sessionID string, sessionName string) SessionDeletedEvent {
	data := map[string]interface{}{
		"session_id":   sessionID,
		"session_name": sessionName,
	}

	return SessionDeletedEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(
			SessionDeletedEventType,
			sessionID,
			data,
		),
	}
}

type SessionErrorEvent struct {
	common.BaseDomainEvent
}

func NewSessionErrorEvent(sessionID string, errorMessage string, errorCode string) SessionErrorEvent {
	data := map[string]interface{}{
		"session_id":    sessionID,
		"error_message": errorMessage,
		"error_code":    errorCode,
	}

	return SessionErrorEvent{
		BaseDomainEvent: common.NewBaseDomainEvent(
			SessionErrorEventType,
			sessionID,
			data,
		),
	}
}
