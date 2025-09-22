package session

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"zpmeow/internal/domain/common"
)

type WebhookEndpoint struct {
	url string
}

func NewWebhookEndpoint(url string) (WebhookEndpoint, error) {
	trimmed := strings.TrimSpace(url)
	if trimmed == "" {
		return WebhookEndpoint{}, fmt.Errorf("webhook URL cannot be empty")
	}

	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return WebhookEndpoint{}, fmt.Errorf("webhook URL must start with http:// or https://")
	}

	return WebhookEndpoint{url: trimmed}, nil
}

func (w WebhookEndpoint) URL() string {
	return w.url
}

func (w WebhookEndpoint) Value() string {
	return w.url
}

func (w WebhookEndpoint) IsEmpty() bool {
	return w.url == ""
}

type Status string

const (
	StatusDisconnected Status = "disconnected"
	StatusConnecting   Status = "connecting"
	StatusConnected    Status = "connected"
	StatusError        Status = "error"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusDisconnected, StatusConnecting, StatusConnected, StatusError:
		return true
	default:
		return false
	}
}

func (s Status) String() string {
	return string(s)
}

type Session struct {
	common.AggregateRoot

	id     SessionID
	name   SessionName
	status Status

	deviceJID DeviceJID
	qrCode    QRCode

	proxyConfig     ProxyConfiguration
	webhookEndpoint WebhookEndpoint
	webhookEvents   []string
	apiKey          ApiKey

	createdAt common.Timestamp
	updatedAt common.Timestamp
}

func NewSession(id, name string) (*Session, error) {
	sessionID, err := NewSessionID(id)
	if err != nil {
		return nil, err
	}

	sessionName, err := NewSessionName(name)
	if err != nil {
		return nil, err
	}

	waJID, _ := NewDeviceJID("")
	qrCode, _ := NewQRCode("")
	proxyConfig, _ := NewProxyConfiguration("")
	webhookEndpoint, _ := NewWebhookEndpoint("")

	now := common.Now()

	session := &Session{
		AggregateRoot:   common.NewAggregateRoot(sessionID.ID),
		id:              sessionID,
		name:            sessionName,
		status:          StatusDisconnected,
		deviceJID:       waJID,
		qrCode:          qrCode,
		proxyConfig:     proxyConfig,
		webhookEndpoint: webhookEndpoint,
		webhookEvents:   []string{},
		apiKey:          ApiKey{},
		createdAt:       now,
		updatedAt:       now,
	}

	if !session.id.IsEmpty() {
		event := NewSessionCreatedEvent(sessionID.Value(), sessionName.Value())
		session.AddEvent(event)
	}

	return session, nil
}

func (s *Session) SessionID() SessionID {
	return s.id
}

func (s *Session) Name() SessionName {
	return s.name
}

func (s *Session) Status() Status {
	return s.status
}

func (s *Session) WaJID() DeviceJID {
	return s.deviceJID
}

func (s *Session) QRCode() QRCode {
	return s.qrCode
}

func (s *Session) ProxyConfiguration() ProxyConfiguration {
	return s.proxyConfig
}

func (s *Session) WebhookEndpoint() WebhookEndpoint {
	return s.webhookEndpoint
}

func (s *Session) ApiKey() ApiKey {
	return s.apiKey
}

func (s *Session) CreatedAt() common.Timestamp {
	return s.createdAt
}

func (s *Session) UpdatedAt() common.Timestamp {
	return s.updatedAt
}

func (s *Session) IsConnected() bool {
	return s.status == StatusConnected
}

func (s *Session) IsDisconnected() bool {
	return s.status == StatusDisconnected
}

func (s *Session) IsConnecting() bool {
	return s.status == StatusConnecting
}

func (s *Session) HasError() bool {
	return s.status == StatusError
}

func (s *Session) CanConnect() bool {
	return s.IsDisconnected() || s.HasError() || s.IsConnecting()
}

func (s *Session) HasQRCode() bool {
	return !s.qrCode.IsEmpty()
}

func (s *Session) HasProxy() bool {
	return !s.proxyConfig.IsEmpty()
}

func (s *Session) IsAuthenticated() bool {
	return !s.deviceJID.IsEmpty()
}

func (s *Session) Connect() error {
	if !s.CanConnect() {
		return fmt.Errorf("session cannot connect from current status: %s", s.status)
	}

	s.status = StatusConnecting
	s.updateTimestamp()

	return nil
}

func (s *Session) Disconnect(reason string) error {
	if s.IsDisconnected() {
		return nil
	}

	s.status = StatusDisconnected
	s.updateTimestamp()

	s.qrCode, _ = NewQRCode("")

	event := NewSessionDisconnectedEvent(s.id.Value(), reason)
	s.AddEvent(event)

	return nil
}

func (s *Session) SetConnected() error {
	if !s.IsConnecting() {
		return fmt.Errorf("session must be connecting to be marked as connected")
	}

	s.status = StatusConnected
	s.updateTimestamp()

	event := NewSessionConnectedEvent(s.id.Value(), s.deviceJID.Value())
	s.AddEvent(event)

	return nil
}

func (s *Session) SetError(errorMessage string) {
	s.status = StatusError
	s.updateTimestamp()

	event := NewSessionErrorEvent(s.id.Value(), errorMessage, "connection_error")
	s.AddEvent(event)
}

func (s *Session) SetStatus(status Status) error {
	if err := ValidateSessionStatus(s.status, status); err != nil {
		return err
	}

	s.status = status
	s.updateTimestamp()
	return nil
}

func (s *Session) SetQRCode(qrCode string) error {
	qr, err := NewQRCode(qrCode)
	if err != nil {
		return err
	}

	s.qrCode = qr
	s.updateTimestamp()

	changes := map[string]interface{}{
		"qr_code_updated": true,
	}
	event := NewSessionConfigurationChangedEvent(s.id.Value(), changes)
	s.AddEvent(event)

	return nil
}

func (s *Session) Authenticate(jid string) error {
	deviceJID, err := NewDeviceJID(jid)
	if err != nil {
		return err
	}

	s.deviceJID = deviceJID
	s.updateTimestamp()

	s.qrCode, _ = NewQRCode("")

	event := NewSessionAuthenticatedEvent(s.id.Value(), jid)
	s.AddEvent(event)

	return nil
}

func (s *Session) SetProxyConfiguration(proxyConfig string) error {
	proxy, err := NewProxyConfiguration(proxyConfig)
	if err != nil {
		return err
	}

	s.proxyConfig = proxy
	s.updateTimestamp()

	changes := map[string]interface{}{
		"proxy_configuration": proxyConfig,
	}
	event := NewSessionConfigurationChangedEvent(s.id.Value(), changes)
	s.AddEvent(event)

	return nil
}

func (s *Session) SetWebhookEndpoint(webhookEndpoint string) error {
	webhook, err := NewWebhookEndpoint(webhookEndpoint)
	if err != nil {
		return err
	}

	s.webhookEndpoint = webhook
	s.updateTimestamp()

	changes := map[string]interface{}{
		"webhook_endpoint": webhookEndpoint,
	}
	event := NewSessionConfigurationChangedEvent(s.id.Value(), changes)
	s.AddEvent(event)

	return nil
}

func (s *Session) SetApiKey(apiKey string) error {
	key, err := NewApiKey(apiKey)
	if err != nil {
		return err
	}

	s.apiKey = key
	s.updateTimestamp()

	return nil
}

func (s *Session) ClearQRCode() {
	s.qrCode, _ = NewQRCode("")
	s.updateTimestamp()
}

func (s *Session) ClearProxy() {
	s.proxyConfig, _ = NewProxyConfiguration("")
	s.updateTimestamp()

	changes := map[string]interface{}{
		"proxy_cleared": true,
	}
	event := NewSessionConfigurationChangedEvent(s.id.Value(), changes)
	s.AddEvent(event)
}

func (s *Session) updateTimestamp() {
	s.updatedAt = common.Now()
}

func (s *Session) Validate() error {
	if s.id.IsEmpty() {
		return ErrInvalidSessionID
	}

	if s.name.IsEmpty() {
		return ErrInvalidSessionName
	}

	if !s.status.IsValid() {
		return ErrInvalidSessionStatus
	}

	return nil
}

func (s *Session) Delete() {
	event := NewSessionDeletedEvent(s.id.Value(), s.name.Value())
	s.AddEvent(event)
}

func (s *Session) HasApiKey() bool {
	return !s.apiKey.IsEmpty()
}

func (s *Session) GetDeviceJIDString() string {
	return s.deviceJID.Value()
}

func (s *Session) GetQRCodeString() string {
	return s.qrCode.Value()
}

func (s *Session) GetApiKeyString() string {
	return s.apiKey.Value()
}

func (s *Session) HasWebhook() bool {
	return !s.webhookEndpoint.IsEmpty()
}

func (s *Session) GetWebhookEndpointString() string {
	return s.webhookEndpoint.Value()
}

func (s *Session) GetWebhookEvents() []string {
	return s.webhookEvents
}

func (s *Session) SetWebhookEvents(events []string) {
	s.webhookEvents = events
	s.updateTimestamp()

	changes := map[string]interface{}{
		"webhook_events": events,
	}
	event := NewSessionConfigurationChangedEvent(s.id.Value(), changes)
	s.AddEvent(event)
}

func (s *Session) SetID(id string) error {
	newID, err := NewSessionID(id)
	if err != nil {
		return err
	}
	s.id = newID
	return nil
}

func (s *Session) MarkCreated() {
	event := NewSessionCreatedEvent(s.id.Value(), s.name.Value())
	s.AddEvent(event)
}

type sessionJSON struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Status          string    `json:"status"`
	DeviceJID       string    `json:"device_jid"`
	QRCode          string    `json:"qr_code"`
	ProxyConfig     string    `json:"proxy_config"`
	WebhookEndpoint string    `json:"webhook_endpoint"`
	WebhookEvents   []string  `json:"webhook_events"`
	ApiKey          string    `json:"api_key"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (s *Session) MarshalJSON() ([]byte, error) {
	return json.Marshal(sessionJSON{
		ID:              s.id.Value(),
		Name:            s.name.Value(),
		Status:          string(s.status),
		DeviceJID:       s.deviceJID.Value(),
		QRCode:          s.qrCode.Value(),
		ProxyConfig:     s.proxyConfig.Value(),
		WebhookEndpoint: s.webhookEndpoint.Value(),
		WebhookEvents:   s.webhookEvents,
		ApiKey:          s.apiKey.Value(),
		CreatedAt:       s.createdAt.Value(),
		UpdatedAt:       s.updatedAt.Value(),
	})
}

func (s *Session) UnmarshalJSON(data []byte) error {
	var sj sessionJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		return err
	}

	sessionID, err := NewSessionID(sj.ID)
	if err != nil {
		return err
	}

	sessionName, err := NewSessionName(sj.Name)
	if err != nil {
		return err
	}

	deviceJID, err := NewDeviceJID(sj.DeviceJID)
	if err != nil {
		return err
	}

	qrCode, err := NewQRCode(sj.QRCode)
	if err != nil {
		return err
	}

	proxyConfig, err := NewProxyConfiguration(sj.ProxyConfig)
	if err != nil {
		return err
	}

	webhookEndpoint, err := NewWebhookEndpoint(sj.WebhookEndpoint)
	if err != nil {
		webhookEndpoint = WebhookEndpoint{}
	}

	apiKey, err := NewApiKey(sj.ApiKey)
	if err != nil {
		apiKey = ApiKey{}
	}

	s.AggregateRoot = common.NewAggregateRoot(sessionID.ID)
	s.id = sessionID
	s.name = sessionName
	s.status = Status(sj.Status)
	s.deviceJID = deviceJID
	s.qrCode = qrCode
	s.proxyConfig = proxyConfig
	s.webhookEndpoint = webhookEndpoint
	s.webhookEvents = sj.WebhookEvents
	s.apiKey = apiKey
	s.createdAt = common.NewTimestamp(sj.CreatedAt)
	s.updatedAt = common.NewTimestamp(sj.UpdatedAt)

	return nil
}
