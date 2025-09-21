package session

import (
	"fmt"
	"strings"

	"zpmeow/internal/domain/common"
)

type SessionID struct {
	common.ID
}

func NewSessionID(value string) (SessionID, error) {
	id, err := common.NewID(value)
	if err != nil {
		return SessionID{}, fmt.Errorf("invalid session ID: %w", err)
	}
	return SessionID{ID: id}, nil
}

func GenerateSessionID() SessionID {
	return SessionID{ID: common.GenerateID()}
}

type SessionName struct {
	common.Name
}

func NewSessionName(value string) (SessionName, error) {
	name, err := common.NewName(value, 3, 100)
	if err != nil {
		return SessionName{}, fmt.Errorf("invalid session name: %w", err)
	}
	return SessionName{Name: name}, nil
}

type ApiKey struct {
	value string
}

func NewApiKey(value string) (ApiKey, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return ApiKey{}, fmt.Errorf("API key cannot be empty")
	}

	if len(trimmed) < 10 {
		return ApiKey{}, fmt.Errorf("API key too short (minimum 10 characters)")
	}

	if len(trimmed) > 100 {
		return ApiKey{}, fmt.Errorf("API key too long (maximum 100 characters)")
	}

	return ApiKey{value: trimmed}, nil
}

func (a ApiKey) Value() string {
	return a.value
}

func (a ApiKey) String() string {
	if len(a.value) <= 8 {
		return "****"
	}
	return a.value[:4] + "****" + a.value[len(a.value)-4:]
}

func (a ApiKey) IsEmpty() bool {
	return a.value == ""
}

type DeviceJID struct {
	value string
}

func NewDeviceJID(value string) (DeviceJID, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return DeviceJID{}, nil
	}

	if !strings.Contains(trimmed, "@") {
		return DeviceJID{}, fmt.Errorf("invalid JID format: must contain @")
	}

	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return DeviceJID{}, fmt.Errorf("invalid JID format: user@server required")
	}

	return DeviceJID{value: trimmed}, nil
}

func (w DeviceJID) Value() string {
	return w.value
}

func (w DeviceJID) String() string {
	return w.value
}

func (w DeviceJID) IsEmpty() bool {
	return w.value == ""
}

type QRCode struct {
	value string
}

func NewQRCode(value string) (QRCode, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return QRCode{}, nil
	}

	if len(trimmed) > 10000 {
		return QRCode{}, fmt.Errorf("QR code too long")
	}

	return QRCode{value: trimmed}, nil
}

func (q QRCode) Value() string {
	return q.value
}

func (q QRCode) String() string {
	return q.value
}

func (q QRCode) IsEmpty() bool {
	return q.value == ""
}

func (q QRCode) IsDataURL() bool {
	return strings.HasPrefix(q.value, "data:")
}

type ProxyConfiguration struct {
	value string
}

func NewProxyConfiguration(value string) (ProxyConfiguration, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return ProxyConfiguration{}, nil
	}

	if len(trimmed) < 7 {
		return ProxyConfiguration{}, fmt.Errorf("proxy configuration too short")
	}

	if len(trimmed) > 500 {
		return ProxyConfiguration{}, fmt.Errorf("proxy configuration too long")
	}

	if !strings.Contains(trimmed, ":") {
		return ProxyConfiguration{}, fmt.Errorf("proxy configuration must contain scheme (:)")
	}

	parts := strings.Split(trimmed, ":")
	if len(parts) != 2 {
		return ProxyConfiguration{}, fmt.Errorf("invalid proxy configuration format")
	}

	scheme := strings.ToLower(parts[0])
	if scheme != "http" && scheme != "https" && scheme != "socks5" {
		return ProxyConfiguration{}, fmt.Errorf("unsupported proxy scheme: %s (supported: http, https, socks5)", scheme)
	}

	if parts[1] == "" {
		return ProxyConfiguration{}, fmt.Errorf("proxy configuration must have host")
	}

	return ProxyConfiguration{value: trimmed}, nil
}

func (p ProxyConfiguration) Value() string {
	return p.value
}

func (p ProxyConfiguration) String() string {
	return p.value
}

func (p ProxyConfiguration) IsEmpty() bool {
	return p.value == ""
}

func (p ProxyConfiguration) Scheme() string {
	if p.value == "" {
		return ""
	}
	parts := strings.Split(p.value, ":")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToLower(parts[0])
}
