package wmeow

import (
	"errors"
	"strings"

	"go.mau.fi/whatsmeow"
	waTypes "go.mau.fi/whatsmeow/types"
)

type messageValidator struct{}

func NewMessageValidator() *messageValidator {
	return &messageValidator{}
}

func (v *messageValidator) ValidateClient(client *whatsmeow.Client) error {
	if client == nil {
		return NewValidationError("client", "cannot be nil")
	}
	return nil
}

func (v *messageValidator) ValidateRecipient(to string) error {
	if strings.TrimSpace(to) == "" {
		return NewValidationError("recipient", "cannot be empty")
	}
	return nil
}

func (v *messageValidator) ValidateTextContent(text string) error {
	if strings.TrimSpace(text) == "" {
		return NewValidationError("text", "cannot be empty")
	}
	return nil
}

func (v *messageValidator) ValidateMediaData(data []byte) error {
	if len(data) == 0 {
		return NewValidationError("media_data", "cannot be empty")
	}
	return nil
}

func (v *messageValidator) ValidateMessageInput(client *whatsmeow.Client, to string) error {
	if err := v.ValidateClient(client); err != nil {
		return err
	}
	return v.ValidateRecipient(to)
}

type phoneParser struct{}

func NewPhoneParser() *phoneParser {
	return &phoneParser{}
}

func (p *phoneParser) ParseToJID(phone string) (waTypes.JID, error) {
	normalized, err := p.NormalizePhoneNumber(phone)
	if err != nil {
		return waTypes.EmptyJID, err
	}

	if err := p.ValidatePhoneNumber(normalized); err != nil {
		return waTypes.EmptyJID, err
	}

	return waTypes.NewJID(normalized, waTypes.DefaultUserServer), nil
}

func (p *phoneParser) NormalizePhoneNumber(phone string) (string, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return "", NewValidationError("phone", "cannot be empty")
	}

	if phone[0] == '+' {
		phone = phone[1:]
	}

	var digits strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}

	normalized := digits.String()
	if normalized == "" {
		return "", NewValidationError("phone", "must contain digits")
	}

	return normalized, nil
}

func (p *phoneParser) ValidatePhoneNumber(phone string) error {
	if len(phone) < 7 || len(phone) > 15 {
		return NewValidationError("phone", "must be between 7 and 15 digits")
	}

	if phone[0] == '0' {
		return NewValidationError("phone", "should not start with 0")
	}

	return nil
}

func ValidateClientAndStore(client *whatsmeow.Client, sessionID string) error {
	if client == nil {
		return NewValidationError("client", "WhatsApp client is nil for session "+sessionID)
	}

	if client.Store == nil {
		return NewValidationError("store", "WhatsApp client store is nil for session "+sessionID)
	}

	return nil
}

func ValidateSessionID(sessionID string) error {
	if strings.TrimSpace(sessionID) == "" {
		return NewValidationError("session_id", "cannot be empty")
	}
	return nil
}

func IsDeviceRegistered(client *whatsmeow.Client) bool {
	return client != nil && client.Store != nil && client.Store.ID != nil
}

func ValidateNonEmpty(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return NewValidationError(fieldName, "cannot be empty")
	}
	return nil
}

func ValidateNonNil(value interface{}, fieldName string) error {
	if value == nil {
		return NewValidationError(fieldName, "cannot be nil")
	}
	return nil
}

func IsValidationError(err error) bool {
	var validationErr *ValidationError
	return errors.As(err, &validationErr)
}

func IsConnectionError(err error) bool {
	var connErr *ConnectionError
	return errors.As(err, &connErr)
}
