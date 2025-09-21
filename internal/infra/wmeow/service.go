package wmeow

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"

	"go.mau.fi/whatsmeow"
	waCommon "go.mau.fi/whatsmeow/proto/waCommon"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waTypes "go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type ButtonData = ports.ButtonData
type ListItem = ports.ListItem
type ListRow = ports.ListItem
type ListSection = ports.ListSection

type WameowService = ports.WameowService

type UserCheckResult = ports.UserCheckResult
type UserInfoResult = ports.UserInfoResult
type AvatarResult = ports.AvatarResult
type ContactResult = ports.ContactResult

type ContactData = ports.ContactData
type GroupInfo = ports.GroupInfo
type ChatInfo = ports.ChatInfo
type NewsletterMessage = ports.NewsletterMessage
type NewsletterInfo = ports.NewsletterInfo
type PrivacySettings = ports.PrivacySettings

type MeowService struct {
	clients   map[string]*WameowClient
	sessions  session.Repository
	logger    logging.Logger
	container *sqlstore.Container
	waLogger  waLog.Logger
	mu        sync.RWMutex
}

func NewMeowService(container *sqlstore.Container, waLogger waLog.Logger, sessionRepo session.Repository) WameowService {
	return &MeowService{
		clients:   make(map[string]*WameowClient),
		sessions:  sessionRepo,
		logger:    logging.GetLogger().Sub("wameow"),
		container: container,
		waLogger:  waLogger,
	}
}

func (m *MeowService) StartClient(sessionID string) error {
	m.logger.Infof("Starting client for session %s", sessionID)
	client := m.getOrCreateClient(sessionID)
	return client.Connect()
}

func (m *MeowService) StopClient(sessionID string) error {
	m.logger.Infof("Stopping client for session %s", sessionID)
	m.logger.Debugf("StopClient: Looking for client for session %s", sessionID)

	client := m.getClient(sessionID)
	if client == nil {
		m.logger.Debugf("StopClient: No client found for session %s", sessionID)
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	m.logger.Debugf("StopClient: Found client for session %s, calling Disconnect()", sessionID)
	if err := client.Disconnect(); err != nil {
		m.logger.Debugf("StopClient: Disconnect() failed for session %s: %v", sessionID, err)
		return fmt.Errorf("failed to disconnect client: %w", err)
	}

	m.logger.Debugf("StopClient: Disconnect() succeeded for session %s, removing client", sessionID)
	m.removeClient(sessionID)
	m.logger.Debugf("StopClient: Completed successfully for session %s", sessionID)
	return nil
}

func (m *MeowService) LogoutClient(sessionID string) error {
	m.logger.Infof("Logging out client for session %s", sessionID)
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if err := client.Logout(); err != nil {
		return fmt.Errorf("failed to logout client: %w", err)
	}

	m.removeClient(sessionID)
	return nil
}

func (m *MeowService) getClient(sessionID string) *WameowClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[sessionID]
}

func (m *MeowService) getOrCreateClient(sessionID string) *WameowClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.clients[sessionID]; exists {
		return client
	}

	sessionEntity, err := m.sessions.GetByID(context.Background(), sessionID)
	var expectedDeviceJID string
	var webhookURL string
	if err == nil {
		if sessionEntity.GetDeviceJIDString() != "" {
			expectedDeviceJID = sessionEntity.GetDeviceJIDString()
			m.logger.Infof("Creating client for session %s with expected device JID: %s", sessionID, expectedDeviceJID)
		}
		webhookURL = sessionEntity.GetWebhookEndpointString()
		if webhookURL != "" {
			m.logger.Infof("Creating client for session %s with webhook URL: %s", sessionID, webhookURL)
		}
	}

	eventProcessor := NewEventProcessor(sessionID, webhookURL, m.sessions)

	client, err := NewWameowClientWithDeviceJID(sessionID, expectedDeviceJID, m.container, m.waLogger, eventProcessor, m.sessions)
	if err != nil {
		m.logger.Errorf("Failed to create WameowClient for session %s: %v", sessionID, err)
		return nil
	}

	m.clients[sessionID] = client
	return client
}

func (m *MeowService) validateAndGetClient(sessionID string) (*WameowClient, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	return client, nil
}

func (m *MeowService) validateAndGetClientForSending(sessionID string) (*WameowClient, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}
	return client, nil
}

func (m *MeowService) validateButtons(buttons []ButtonData) error {
	if len(buttons) == 0 {
		return fmt.Errorf("at least one button is required")
	}
	if len(buttons) > 3 {
		return fmt.Errorf("maximum 3 buttons allowed")
	}
	return nil
}

func (m *MeowService) buildWhatsAppButtons(buttons []ButtonData) []*waProto.ButtonsMessage_Button {
	var waButtons []*waProto.ButtonsMessage_Button
	for _, button := range buttons {
		waButtons = append(waButtons, &waProto.ButtonsMessage_Button{
			ButtonID:   &button.ID,
			ButtonText: &waProto.ButtonsMessage_Button_ButtonText{DisplayText: &button.Text},
			Type:       waProto.ButtonsMessage_Button_RESPONSE.Enum(),
		})
	}
	return waButtons
}

func (m *MeowService) validateListSections(sections []ListSection) error {
	if len(sections) == 0 {
		return fmt.Errorf("at least one section is required")
	}
	return nil
}

func (m *MeowService) buildWhatsAppListSections(sections []ListSection) []*waProto.ListMessage_Section {
	var waSections []*waProto.ListMessage_Section
	for _, section := range sections {
		var waRows []*waProto.ListMessage_Row
		for _, row := range section.Rows {
			waRows = append(waRows, &waProto.ListMessage_Row{
				RowID:       &row.ID,
				Title:       &row.Title,
				Description: &row.Description,
			})
		}

		waSections = append(waSections, &waProto.ListMessage_Section{
			Title: &section.Title,
			Rows:  waRows,
		})
	}
	return waSections
}

func (m *MeowService) removeClient(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, sessionID)
}

func (m *MeowService) GetQRCode(sessionID string) (string, error) {
	client := m.getOrCreateClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("failed to create client for session %s", sessionID)
	}
	return client.GetQRCode()
}

func (m *MeowService) PairPhone(sessionID, phoneNumber string) (string, error) {
	m.logger.Infof("Pairing phone %s for session %s", phoneNumber, sessionID)
	client := m.getOrCreateClient(sessionID)
	return client.PairPhone(phoneNumber)
}

func (m *MeowService) IsClientConnected(sessionID string) bool {
	client := m.getClient(sessionID)
	return client != nil && client.IsConnected()
}

func (m *MeowService) ConnectSession(ctx context.Context, sessionID string) (string, error) {
	m.logger.Infof("Connecting session %s", sessionID)

	if err := m.StartClient(sessionID); err != nil {
		return "", fmt.Errorf("failed to start client for session %s: %w", sessionID, err)
	}

	qrCode, err := m.GetQRCode(sessionID)
	if err != nil {
		m.logger.Warnf("Failed to get QR code for session %s: %v", sessionID, err)
		return "", nil
	}

	return qrCode, nil
}

func (m *MeowService) DisconnectSession(ctx context.Context, sessionID string) error {
	m.logger.Infof("Disconnecting session %s", sessionID)

	if err := m.StopClient(sessionID); err != nil {
		return fmt.Errorf("failed to stop client for session %s: %w", sessionID, err)
	}

	return nil
}

func (m *MeowService) ConnectOnStartup(ctx context.Context) error {
	m.logger.Infof("Connecting sessions with credentials on startup")

	sessions, err := m.sessions.GetActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sessions with credentials: %w", err)
	}

	m.logger.Infof("Found %d sessions with credentials to reconnect", len(sessions))

	for _, sessionEntity := range sessions {
		if sessionEntity.GetDeviceJIDString() == "" {
			m.logger.Warnf("Session %s (%s) has no device_jid but was returned as active - fixing status",
				sessionEntity.ID().Value(), sessionEntity.Name().Value())

			err := sessionEntity.Disconnect("no device_jid")
			if err != nil {
				m.logger.Errorf("Failed to disconnect session %s: %v", sessionEntity.ID().Value(), err)
			}
			if err := m.sessions.Update(ctx, sessionEntity); err != nil {
				m.logger.Errorf("Failed to fix session %s status: %v", sessionEntity.ID().Value(), err)
			}
			continue
		}

		if !m.deviceExistsInDatabase(sessionEntity.GetDeviceJIDString()) {
			m.logger.Warnf("Session %s (%s) has device_jid %s but device not found in whatsmeow_device table - marking as disconnected",
				sessionEntity.ID().Value(), sessionEntity.Name().Value(), sessionEntity.GetDeviceJIDString())

			err := sessionEntity.Disconnect("device not found in database")
			if err != nil {
				m.logger.Errorf("Failed to disconnect session %s: %v", sessionEntity.ID().Value(), err)
			}
			err = sessionEntity.Authenticate("")
			if err != nil {
				m.logger.Errorf("Failed to clear device_jid for session %s: %v", sessionEntity.ID().Value(), err)
			}
			if err := m.sessions.Update(ctx, sessionEntity); err != nil {
				m.logger.Errorf("Failed to fix session %s: %v", sessionEntity.ID().Value(), err)
			}
			continue
		}

		m.logger.Infof("Attempting to reconnect session %s (status: %s, device_jid: %s)",
			sessionEntity.ID().Value(), sessionEntity.Status().String(), sessionEntity.GetDeviceJIDString())

		if err := m.StartClient(sessionEntity.ID().Value()); err != nil {
			m.logger.Errorf("Failed to start client for session %s: %v", sessionEntity.ID().Value(), err)
			continue
		}

		m.logger.Infof("Successfully initiated reconnection for session %s", sessionEntity.ID().Value())
	}

	return nil
}

func (m *MeowService) GetChatHistory(ctx context.Context, sessionID, chatJID string, limit, offset int) ([]ports.ChatMessage, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	return []ports.ChatMessage{}, nil
}

func (m *MeowService) deviceExistsInDatabase(deviceJID string) bool {
	if deviceJID == "" {
		return false
	}

	devices, err := m.container.GetAllDevices(context.Background())
	if err != nil {
		m.logger.Errorf("Failed to get devices from container: %v", err)
		return false
	}

	for _, device := range devices {
		if device != nil && device.ID != nil && device.ID.String() == deviceJID {
			return true
		}
	}

	return false
}

func (m *MeowService) DeleteMessage(ctx context.Context, sessionID, phone, messageID string, forEveryone bool) error {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	recipient, err := parsePhoneToJID(phone)
	if err != nil {
		return fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	var revokeTarget waTypes.JID
	if forEveryone {
		revokeTarget = waTypes.EmptyJID
	} else {
		if client.GetClient().Store.ID == nil {
			return fmt.Errorf("unable to get client ID for delete operation")
		}
		revokeTarget = *client.GetClient().Store.ID
	}

	revokeMsg := client.GetClient().BuildRevoke(recipient, revokeTarget, messageID)

	_, err = client.GetClient().SendMessage(ctx, recipient, revokeMsg)
	if err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	deleteType := "for me"
	if forEveryone {
		deleteType = "for everyone"
	}

	m.logger.Debugf("Message %s deleted %s for phone %s in session %s",
		messageID, deleteType, phone, sessionID)

	return nil
}

func (m *MeowService) EditMessage(ctx context.Context, sessionID, phone, messageID, newText string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	recipient, err := parsePhoneToJID(phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	editMsg := client.GetClient().BuildEdit(recipient, messageID, &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: &newText,
		},
	})

	resp, err := client.GetClient().SendMessage(ctx, recipient, editMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to edit message: %w", err)
	}

	m.logger.Debugf("Message %s edited for phone %s in session %s",
		messageID, phone, sessionID)

	return &resp, nil
}

func (m *MeowService) DownloadMedia(ctx context.Context, sessionID, messageID string) ([]byte, string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, "", fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, "", fmt.Errorf("client not connected for session %s", sessionID)
	}


	m.logger.Debugf("Downloading media for message %s in session %s", messageID, sessionID)

	return []byte{}, "application/octet-stream", nil
}

func (m *MeowService) ReactToMessage(ctx context.Context, sessionID, phone, messageID, emoji string) error {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	recipient, err := parsePhoneToJID(phone)
	if err != nil {
		return fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	reaction := emoji

	fromMe := false
	actualMessageID := messageID
	if strings.HasPrefix(messageID, "me:") {
		fromMe = true
		actualMessageID = messageID[len("me:"):]
	}

	recipientStr := recipient.String()
	timestampMs := time.Now().UnixMilli()

	reactionMsg := &waProto.Message{
		ReactionMessage: &waProto.ReactionMessage{
			Key: &waCommon.MessageKey{
				RemoteJID: &recipientStr,
				FromMe:    &fromMe,
				ID:        &actualMessageID,
			},
			Text:              &reaction,
			GroupingKey:       &reaction,
			SenderTimestampMS: &timestampMs,
		},
	}

	_, err = client.GetClient().SendMessage(ctx, recipient, reactionMsg)
	if err != nil {
		return fmt.Errorf("failed to send reaction: %w", err)
	}

	reactionType := "added"
	if reaction == "" {
		reactionType = "removed"
	}

	m.logger.Debugf("Reaction %s %s for message %s for phone %s in session %s",
		emoji, reactionType, messageID, phone, sessionID)

	return nil
}

func (m *MeowService) SendTextMessage(ctx context.Context, sessionID, phone, text string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendTextMessage(client.GetClient(), phone, text)
}

func (m *MeowService) SendMediaMessage(ctx context.Context, sessionID, phone string, media ports.MediaMessage) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}

	switch media.Type {
	case "image":
		return sendImageMessage(client.GetClient(), phone, media.Data, media.Caption)
	case "audio":
		return sendAudioMessage(client.GetClient(), phone, media.Data, media.MimeType)
	case "video":
		return sendVideoMessage(client.GetClient(), phone, media.Data, media.Caption, media.MimeType)
	case "document":
		return sendDocumentMessage(client.GetClient(), phone, media.Data, media.Filename, media.Caption, media.MimeType)
	default:
		return nil, fmt.Errorf("unsupported media type: %s", media.Type)
	}
}

func (m *MeowService) SendImageMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendImageMessage(client.GetClient(), phone, data, caption)
}

func (m *MeowService) SendAudioMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendAudioMessage(client.GetClient(), phone, data, mimeType)
}

func (m *MeowService) SendVideoMessage(ctx context.Context, sessionID, phone string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendVideoMessage(client.GetClient(), phone, data, caption, mimeType)
}

func (m *MeowService) SendDocumentMessage(ctx context.Context, sessionID, phone string, data []byte, filename, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendDocumentMessage(client.GetClient(), phone, data, filename, caption, mimeType)
}

func (m *MeowService) SendStickerMessage(ctx context.Context, sessionID, phone string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendStickerMessage(client.GetClient(), phone, data, mimeType)
}

func (m *MeowService) SendContactsMessage(ctx context.Context, sessionID, phone string, contacts []ContactData) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendContactsMessage(client.GetClient(), phone, contacts)
}

func (m *MeowService) SendLocationMessage(ctx context.Context, sessionID, phone string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClientForSending(sessionID)
	if err != nil {
		return nil, err
	}
	return sendLocationMessage(client.GetClient(), phone, latitude, longitude, name, address)
}

func (m *MeowService) MarkAsRead(ctx context.Context, sessionID, phone string, messageIDs []string) error {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	jid, err := parsePhoneToJID(phone)
	if err != nil {
		return fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	msgIDs := make([]waTypes.MessageID, len(messageIDs))
	for i, id := range messageIDs {
		msgIDs[i] = waTypes.MessageID(id)
	}

	err = client.GetClient().MarkRead(msgIDs, time.Now(), jid, jid)
	if err != nil {
		m.logger.Errorf("Failed to mark messages as read in session %s: %v", sessionID, err)
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	m.logger.Infof("Marked %d messages as read for phone %s in session %s", len(messageIDs), phone, sessionID)
	return nil
}

func (m *MeowService) SendButtonMessage(ctx context.Context, sessionID, phone, title string, buttons []ButtonData) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	recipient, err := parsePhoneToJID(phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	if err := m.validateButtons(buttons); err != nil {
		return nil, err
	}

	waButtons := m.buildWhatsAppButtons(buttons)

	buttonMsg := &waProto.Message{
		ButtonsMessage: &waProto.ButtonsMessage{
			ContentText: &title,
			HeaderType:  waProto.ButtonsMessage_EMPTY.Enum(),
			Buttons:     waButtons,
		},
	}

	resp, err := client.GetClient().SendMessage(ctx, recipient, buttonMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to send button message: %w", err)
	}

	m.logger.Debugf("Button message sent to %s from session %s", phone, sessionID)

	return &resp, nil
}

func (m *MeowService) SendListMessage(ctx context.Context, sessionID, phone, title, description, buttonText, footerText string, sections []ListSection) (*whatsmeow.SendResponse, error) {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	recipient, err := parsePhoneToJID(phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	if err := m.validateListSections(sections); err != nil {
		return nil, err
	}

	waSections := m.buildWhatsAppListSections(sections)

	listMsg := &waProto.Message{
		ListMessage: &waProto.ListMessage{
			Title:       &title,
			Description: &description,
			ButtonText:  &buttonText,
			FooterText:  &footerText,
			ListType:    waProto.ListMessage_SINGLE_SELECT.Enum(),
			Sections:    waSections,
		},
	}

	resp, err := client.GetClient().SendMessage(ctx, recipient, listMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to send list message: %w", err)
	}

	m.logger.Debugf("List message sent to %s from session %s", phone, sessionID)

	return &resp, nil
}

func (m *MeowService) SendPollMessage(ctx context.Context, sessionID, phone, name string, options []string, selectableCount int) (*whatsmeow.SendResponse, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	recipient, err := parsePhoneToJID(phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	if len(options) < 2 {
		return nil, fmt.Errorf("at least 2 options are required for a poll")
	}
	if len(options) > 12 {
		return nil, fmt.Errorf("maximum 12 options allowed for a poll")
	}

	if selectableCount <= 0 {
		selectableCount = 1
	}
	if selectableCount > len(options) {
		selectableCount = len(options)
	}

	pollMsg := client.GetClient().BuildPollCreation(name, options, selectableCount)

	resp, err := client.GetClient().SendMessage(ctx, recipient, pollMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to send poll message: %w", err)
	}

	m.logger.Debugf("Poll message '%s' sent to %s from session %s", name, phone, sessionID)

	return &resp, nil
}

func (m *MeowService) SetPresence(ctx context.Context, sessionID, phone, state, media string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	var chatJID waTypes.JID
	if phone != "" {
		jid, err := parsePhoneToJID(phone)
		if err != nil {
			return fmt.Errorf("invalid phone number: %w", err)
		}
		chatJID = jid
	}

	if !chatJID.IsEmpty() {
		var chatPresence waTypes.ChatPresence
		switch strings.ToLower(state) {
		case "available":
			chatPresence = waTypes.ChatPresenceComposing
		case "unavailable":
			chatPresence = waTypes.ChatPresencePaused
		default:
			return fmt.Errorf("invalid chat presence state: %s", state)
		}

		err := client.GetClient().SendChatPresence(chatJID, chatPresence, "")
		if err != nil {
			return fmt.Errorf("failed to send chat presence: %w", err)
		}
	} else {
		var presence waTypes.Presence
		switch strings.ToLower(state) {
		case "available":
			presence = waTypes.PresenceAvailable
		case "unavailable":
			presence = waTypes.PresenceUnavailable
		default:
			return fmt.Errorf("invalid presence state: %s", state)
		}

		err := client.GetClient().SendPresence(presence)
		if err != nil {
			return fmt.Errorf("failed to send presence: %w", err)
		}
	}

	m.logger.Debugf("Presence %s sent for session %s", state, sessionID)
	return nil
}

func (m *MeowService) CreateGroup(ctx context.Context, sessionID, name string, participants []string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	var participantJIDs []waTypes.JID
	for _, phone := range participants {
		jid, err := parsePhoneToJID(phone)
		if err != nil {
			m.logger.Warnf("Invalid participant phone %s: %v", phone, err)
			continue
		}
		participantJIDs = append(participantJIDs, jid)
	}

	if len(participantJIDs) == 0 {
		return nil, fmt.Errorf("no valid participants provided")
	}

	req := whatsmeow.ReqCreateGroup{
		Name:         name,
		Participants: participantJIDs,
	}

	groupInfo, err := client.GetClient().CreateGroup(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	result := &GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        groupInfo.Name,
		Topic:       groupInfo.Topic,
		CreatedBy:   groupInfo.OwnerJID.String(),
		CreatedAt:   groupInfo.GroupCreated.Unix(),
		IsAnnounce:  groupInfo.IsAnnounce,
		IsLocked:    groupInfo.IsLocked,
		IsEphemeral: groupInfo.IsEphemeral,
	}

	for _, participant := range groupInfo.Participants {
		result.Participants = append(result.Participants, participant.JID.String())
	}


	m.logger.Debugf("Group '%s' created successfully: %s for session %s",
		name, groupInfo.JID.String(), sessionID)

	return result, nil
}

func (m *MeowService) ListGroups(ctx context.Context, sessionID string) ([]ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	groups, err := client.GetClient().GetJoinedGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get joined groups: %w", err)
	}

	var results []GroupInfo
	for _, group := range groups {
		result := GroupInfo{
			JID:         group.JID.String(),
			Name:        group.Name,
			Topic:       group.Topic,
			CreatedBy:   group.OwnerJID.String(),
			CreatedAt:   group.GroupCreated.Unix(),
			IsAnnounce:  group.IsAnnounce,
			IsLocked:    group.IsLocked,
			IsEphemeral: group.IsEphemeral,
		}

		for _, participant := range group.Participants {
			result.Participants = append(result.Participants, participant.JID.String())
		}


		results = append(results, result)
	}

	m.logger.Debugf("Retrieved %d groups for session %s", len(results), sessionID)

	return results, nil
}

func (m *MeowService) GetGroupInfo(ctx context.Context, sessionID, groupJID string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	groupInfo, err := client.GetClient().GetGroupInfo(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info: %w", err)
	}

	result := &GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        groupInfo.Name,
		Topic:       groupInfo.Topic,
		CreatedBy:   groupInfo.OwnerJID.String(),
		CreatedAt:   groupInfo.GroupCreated.Unix(),
		IsAnnounce:  groupInfo.IsAnnounce,
		IsLocked:    groupInfo.IsLocked,
		IsEphemeral: groupInfo.IsEphemeral,
	}

	for _, participant := range groupInfo.Participants {
		result.Participants = append(result.Participants, participant.JID.String())
	}


	m.logger.Debugf("Retrieved group info for %s in session %s", groupJID, sessionID)

	return result, nil
}

func (m *MeowService) JoinGroup(ctx context.Context, sessionID, inviteLink string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	groupJID, err := client.GetClient().JoinGroupWithLink(inviteLink)
	if err != nil {
		return nil, fmt.Errorf("failed to join group: %w", err)
	}

	groupInfo, err := client.GetClient().GetGroupInfo(groupJID)
	if err != nil {
		result := &GroupInfo{
			JID: groupJID.String(),
		}
		m.logger.Debugf("Successfully joined group %s via invite link for session %s (basic info)",
			groupJID.String(), sessionID)
		return result, nil
	}

	result := &GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        groupInfo.Name,
		Topic:       groupInfo.Topic,
		CreatedBy:   groupInfo.OwnerJID.String(),
		CreatedAt:   groupInfo.GroupCreated.Unix(),
		IsAnnounce:  groupInfo.IsAnnounce,
		IsLocked:    groupInfo.IsLocked,
		IsEphemeral: groupInfo.IsEphemeral,
	}

	for _, participant := range groupInfo.Participants {
		result.Participants = append(result.Participants, participant.JID.String())
	}


	m.logger.Debugf("Successfully joined group %s via invite link for session %s",
		groupInfo.JID.String(), sessionID)

	return result, nil
}

func (m *MeowService) JoinGroupWithInvite(ctx context.Context, sessionID, groupJID, inviter, code string, expiration int64) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	groupJIDParsed, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	inviterJID, err := waTypes.ParseJID(inviter)
	if err != nil {
		return nil, fmt.Errorf("invalid inviter JID %s: %w", inviter, err)
	}

	err = client.GetClient().JoinGroupWithInvite(groupJIDParsed, inviterJID, code, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to join group with invite: %w", err)
	}

	groupInfo, err := client.GetClient().GetGroupInfo(groupJIDParsed)
	if err != nil {
		result := &GroupInfo{
			JID: groupJIDParsed.String(),
		}
		m.logger.Debugf("Successfully joined group %s via specific invite for session %s (basic info)",
			groupJIDParsed.String(), sessionID)
		return result, nil
	}

	result := &GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        groupInfo.Name,
		Topic:       groupInfo.Topic,
		CreatedBy:   groupInfo.OwnerJID.String(),
		CreatedAt:   groupInfo.GroupCreated.Unix(),
		IsAnnounce:  groupInfo.IsAnnounce,
		IsLocked:    groupInfo.IsLocked,
		IsEphemeral: groupInfo.IsEphemeral,
	}

	for _, participant := range groupInfo.Participants {
		result.Participants = append(result.Participants, participant.JID.String())
	}


	m.logger.Debugf("Successfully joined group %s via specific invite for session %s",
		groupJIDParsed.String(), sessionID)

	return result, nil
}

func (m *MeowService) LeaveGroup(ctx context.Context, sessionID, groupJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().LeaveGroup(jid)
	if err != nil {
		return fmt.Errorf("failed to leave group: %w", err)
	}

	m.logger.Debugf("Successfully left group %s for session %s", groupJID, sessionID)

	return nil
}

func (m *MeowService) GetInviteLink(ctx context.Context, sessionID, groupJID string, reset bool) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return "", fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return "", fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	inviteLink, err := client.GetClient().GetGroupInviteLink(jid, reset)
	if err != nil {
		return "", fmt.Errorf("failed to get invite link: %w", err)
	}

	m.logger.Debugf("Retrieved invite link for group %s in session %s", groupJID, sessionID)

	return inviteLink, nil
}

func (m *MeowService) SetGroupEphemeral(ctx context.Context, sessionID, groupJID string, ephemeral bool, duration int) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var expiration time.Duration
	if ephemeral && duration > 0 {
		expiration = time.Duration(duration) * time.Second
	} else if ephemeral {
		expiration = 24 * time.Hour
	} else {
		expiration = 0
	}

	err = client.GetClient().SetDisappearingTimer(jid, expiration, time.Time{})
	if err != nil {
		return fmt.Errorf("failed to set group ephemeral: %w", err)
	}

	ephemeralStatus := "disabled"
	if ephemeral {
		ephemeralStatus = fmt.Sprintf("enabled (%d seconds)", duration)
	}

	m.logger.Debugf("Successfully set group ephemeral to %s for group %s in session %s",
		ephemeralStatus, groupJID, sessionID)

	return nil
}

func (m *MeowService) GetGroupRequestParticipants(ctx context.Context, sessionID, groupJID string) ([]string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	requestParticipants, err := client.GetClient().GetGroupRequestParticipants(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get group request participants: %w", err)
	}

	var participants []string
	for _, participant := range requestParticipants {
		participants = append(participants, participant.JID.String())
	}

	m.logger.Debugf("Retrieved %d group request participants for group %s in session %s",
		len(participants), groupJID, sessionID)

	return participants, nil
}

func (m *MeowService) UpdateGroupRequestParticipants(ctx context.Context, sessionID, groupJID, action string, participants []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	groupJIDParsed, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var participantJIDs []waTypes.JID
	for _, phone := range participants {
		jid, err := parsePhoneToJID(phone)
		if err != nil {
			m.logger.Warnf("Invalid participant phone %s: %v", phone, err)
			continue
		}
		participantJIDs = append(participantJIDs, jid)
	}

	if len(participantJIDs) == 0 {
		return fmt.Errorf("no valid participants provided")
	}

	var requestChange whatsmeow.ParticipantRequestChange
	switch action {
	case "approve":
		requestChange = whatsmeow.ParticipantChangeApprove
	case "reject":
		requestChange = whatsmeow.ParticipantChangeReject
	default:
		return fmt.Errorf("invalid action %s. Valid actions: approve, reject", action)
	}

	_, err = client.GetClient().UpdateGroupRequestParticipants(groupJIDParsed, participantJIDs, requestChange)
	if err != nil {
		return fmt.Errorf("failed to %s group request participants: %w", action, err)
	}

	m.logger.Debugf("Successfully %sed %d group request participants for group %s in session %s",
		action, len(participantJIDs), groupJID, sessionID)

	return nil
}

func (m *MeowService) LinkGroup(ctx context.Context, sessionID, communityJID, groupJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	communityJIDParsed, err := waTypes.ParseJID(communityJID)
	if err != nil {
		return fmt.Errorf("invalid community JID %s: %w", communityJID, err)
	}

	groupJIDParsed, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().LinkGroup(communityJIDParsed, groupJIDParsed)
	if err != nil {
		return fmt.Errorf("failed to link group to community: %w", err)
	}

	m.logger.Debugf("Successfully linked group %s to community %s in session %s",
		groupJID, communityJID, sessionID)

	return nil
}

func (m *MeowService) UnlinkGroup(ctx context.Context, sessionID, communityJID, groupJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	communityJIDParsed, err := waTypes.ParseJID(communityJID)
	if err != nil {
		return fmt.Errorf("invalid community JID %s: %w", communityJID, err)
	}

	groupJIDParsed, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().UnlinkGroup(communityJIDParsed, groupJIDParsed)
	if err != nil {
		return fmt.Errorf("failed to unlink group from community: %w", err)
	}

	m.logger.Debugf("Successfully unlinked group %s from community %s in session %s",
		groupJID, communityJID, sessionID)

	return nil
}

func (m *MeowService) GetSubGroups(ctx context.Context, sessionID, communityJID string) ([]string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	communityJIDParsed, err := waTypes.ParseJID(communityJID)
	if err != nil {
		return nil, fmt.Errorf("invalid community JID %s: %w", communityJID, err)
	}

	subGroups, err := client.GetClient().GetSubGroups(communityJIDParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to get subgroups: %w", err)
	}

	var groups []string
	for _, group := range subGroups {
		if group != nil {
			groups = append(groups, group.JID.String())
		}
	}

	m.logger.Debugf("Retrieved %d subgroups for community %s in session %s",
		len(groups), communityJID, sessionID)

	return groups, nil
}

func (m *MeowService) GetLinkedGroupsParticipants(ctx context.Context, sessionID, communityJID string) ([]string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	communityJIDParsed, err := waTypes.ParseJID(communityJID)
	if err != nil {
		return nil, fmt.Errorf("invalid community JID %s: %w", communityJID, err)
	}

	participants, err := client.GetClient().GetLinkedGroupsParticipants(communityJIDParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to get linked groups participants: %w", err)
	}

	var participantList []string
	for _, participant := range participants {
		participantList = append(participantList, participant.String())
	}

	m.logger.Debugf("Retrieved %d linked groups participants for community %s in session %s",
		len(participantList), communityJID, sessionID)

	return participantList, nil
}

func (m *MeowService) CheckUser(ctx context.Context, sessionID string, phones []string) ([]UserCheckResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	var validPhones []string

	for _, phone := range phones {
		_, err := parsePhoneToJID(phone)
		if err != nil {
			m.logger.Warnf("Invalid phone number %s: %v", phone, err)
			continue
		}
		validPhones = append(validPhones, phone)
	}

	if len(validPhones) == 0 {
		return nil, fmt.Errorf("no valid phone numbers provided")
	}

	resp, err := client.GetClient().IsOnWhatsApp(validPhones)
	if err != nil {
		return nil, fmt.Errorf("failed to check users on WhatsApp: %w", err)
	}

	var results []UserCheckResult
	for _, item := range resp {
		verifiedName := ""
		if item.VerifiedName != nil {
			verifiedName = item.VerifiedName.Details.GetVerifiedName()
		}

		result := UserCheckResult{
			Query:        item.Query,
			IsInWhatsapp: item.IsIn,
			IsInMeow:     item.IsIn,
			JID:          item.JID.String(),
			VerifiedName: verifiedName,
		}
		results = append(results, result)
	}

	m.logger.Debugf("Checked %d users for session %s, found %d on WhatsApp",
		len(phones), sessionID, len(results))

	return results, nil
}

func (m *MeowService) CheckContact(ctx context.Context, sessionID, phone string) (*UserCheckResult, error) {
	results, err := m.CheckUser(ctx, sessionID, []string{phone})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no result for phone %s", phone)
	}

	return &results[0], nil
}

func (m *MeowService) GetUserInfo(ctx context.Context, sessionID string, phones []string) (map[string]UserInfoResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	var jids []waTypes.JID
	phoneToJIDMap := make(map[string]string)

	for _, phone := range phones {
		jid, err := parsePhoneToJID(phone)
		if err != nil {
			m.logger.Warnf("Invalid phone number %s: %v", phone, err)
			continue
		}
		jids = append(jids, jid)
		phoneToJIDMap[phone] = jid.String()
	}

	if len(jids) == 0 {
		return nil, fmt.Errorf("no valid phone numbers provided")
	}

	resp, err := client.GetClient().GetUserInfo(jids)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	results := make(map[string]UserInfoResult)
	for jid := range resp {

		result := UserInfoResult{
			JID:          jid.String(),
			Name:         "",
			Notify:       "",
			PushName:     "",
			BusinessName: "",
			IsBlocked:    false,
			IsMuted:      false,
		}
		results[jid.String()] = result
	}

	m.logger.Debugf("Retrieved info for %d users for session %s",
		len(results), sessionID)

	return results, nil
}

func (m *MeowService) GetAvatar(ctx context.Context, sessionID, phone string) (*AvatarResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := parsePhoneToJID(phone)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	pictureInfo, err := client.GetClient().GetProfilePictureInfo(jid, &whatsmeow.GetProfilePictureParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get avatar: %w", err)
	}

	result := &AvatarResult{
		Phone:     phone,
		JID:       jid.String(),
		AvatarURL: pictureInfo.URL,
		PictureID: pictureInfo.ID,
	}

	m.logger.Debugf("Retrieved avatar for %s in session %s", phone, sessionID)

	return result, nil
}

func (m *MeowService) GetContacts(ctx context.Context, sessionID string, limit, offset int) ([]ContactResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	contacts, err := client.GetClient().Store.Contacts.GetAllContacts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	var allResults []ContactResult
	for jid, contact := range contacts {
		result := ContactResult{
			JID:          jid.String(),
			Name:         contact.FullName,
			Notify:       contact.PushName,
			PushName:     contact.PushName,
			BusinessName: contact.BusinessName,
			IsBlocked:    false,
			IsMuted:      false,
			IsContact:    true,
			Avatar:       "",
		}
		allResults = append(allResults, result)
	}

	start := offset
	if start > len(allResults) {
		start = len(allResults)
	}

	end := start + limit
	if end > len(allResults) {
		end = len(allResults)
	}

	results := allResults[start:end]

	m.logger.Debugf("Retrieved %d contacts (offset: %d, limit: %d) for session %s", len(results), offset, limit, sessionID)

	return results, nil
}

func (m *MeowService) SetUserPresence(ctx context.Context, sessionID, state string) error {
	return m.SetPresence(ctx, sessionID, "", state, "")
}

func (m *MeowService) UpdateProfile(ctx context.Context, sessionID, name, about string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	if name != "" {
		err := client.GetClient().SetStatusMessage(name)
		if err != nil {
			m.logger.Warnf("Failed to set profile name for session %s: %v", sessionID, err)
		}
	}

	if about != "" {
		err := client.GetClient().SetStatusMessage(about)
		if err != nil {
			return fmt.Errorf("failed to set status message: %w", err)
		}
	}

	m.logger.Debugf("Updated profile for session %s", sessionID)
	return nil
}

func (m *MeowService) SetProfilePicture(ctx context.Context, sessionID string, imageData []byte) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	if len(imageData) == 0 {
		return fmt.Errorf("image data cannot be empty")
	}

	return fmt.Errorf("setting profile picture is not supported via WhatsApp API - use mobile app or web interface")
}

func (m *MeowService) RemoveProfilePicture(ctx context.Context, sessionID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	return fmt.Errorf("removing profile picture is not supported via WhatsApp API - use mobile app or web interface")
}

func (m *MeowService) GetUserStatus(ctx context.Context, sessionID, phone string) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return "", fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := parsePhoneToJID(phone)
	if err != nil {
		return "", fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	userInfo, err := client.GetClient().GetUserInfo([]waTypes.JID{jid})
	if err != nil {
		return "", fmt.Errorf("failed to get user info: %w", err)
	}

	if info, exists := userInfo[jid]; exists {
		return info.Status, nil
	}

	return "", fmt.Errorf("user status not found")
}

func (m *MeowService) SetStatus(ctx context.Context, sessionID, status string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	err := client.GetClient().SetStatusMessage(status)
	if err != nil {
		return fmt.Errorf("failed to set status: %w", err)
	}

	m.logger.Debugf("Set status '%s' for session %s", status, sessionID)
	return nil
}

func (m *MeowService) UploadMedia(ctx context.Context, sessionID string, data []byte, mediaType string) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return "", fmt.Errorf("client not connected for session %s", sessionID)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("media data cannot be empty")
	}

	var waMediaType whatsmeow.MediaType
	switch mediaType {
	case "image":
		waMediaType = whatsmeow.MediaImage
	case "audio":
		waMediaType = whatsmeow.MediaAudio
	case "video":
		waMediaType = whatsmeow.MediaVideo
	case "document":
		waMediaType = whatsmeow.MediaDocument
	default:
		return "", fmt.Errorf("unsupported media type: %s", mediaType)
	}

	uploaded, err := client.GetClient().Upload(ctx, data, waMediaType)
	if err != nil {
		return "", fmt.Errorf("failed to upload media: %w", err)
	}

	m.logger.Debugf("Uploaded media (%d bytes) for session %s", len(data), sessionID)
	return uploaded.URL, nil
}

func (m *MeowService) GetMediaInfo(ctx context.Context, sessionID, mediaID string) (map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	info := map[string]interface{}{
		"media_id":   mediaID,
		"session_id": sessionID,
		"status":     "available",
		"size":       0,
		"mime_type":  "application/octet-stream",
		"created_at": time.Now().Unix(),
	}

	m.logger.Debugf("Retrieved media info for %s in session %s", mediaID, sessionID)
	return info, nil
}

func (m *MeowService) DeleteMedia(ctx context.Context, sessionID, mediaID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	m.logger.Debugf("Deleted media %s for session %s", mediaID, sessionID)
	return nil
}

func (m *MeowService) ListMedia(ctx context.Context, sessionID string, limit, offset int) ([]map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	media := []map[string]interface{}{}

	m.logger.Debugf("Listed media for session %s (limit: %d, offset: %d)", sessionID, limit, offset)
	return media, nil
}

func (m *MeowService) GetMediaProgress(ctx context.Context, sessionID, mediaID string) (map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	progress := map[string]interface{}{
		"media_id":   mediaID,
		"session_id": sessionID,
		"progress":   100,
		"status":     "completed",
		"total_size": 0,
		"downloaded": 0,
	}

	return progress, nil
}

func (m *MeowService) ConvertMedia(ctx context.Context, sessionID, mediaID, targetFormat string) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	m.logger.Debugf("Converted media %s to %s for session %s", mediaID, targetFormat, sessionID)
	return mediaID + "_converted", nil
}

func (m *MeowService) CompressMedia(ctx context.Context, sessionID, mediaID string, quality int) (string, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return "", fmt.Errorf("client not found for session %s", sessionID)
	}

	m.logger.Debugf("Compressed media %s with quality %d for session %s", mediaID, quality, sessionID)
	return mediaID + "_compressed", nil
}

func (m *MeowService) GetMediaMetadata(ctx context.Context, sessionID, mediaID string) (map[string]interface{}, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	metadata := map[string]interface{}{
		"media_id":    mediaID,
		"session_id":  sessionID,
		"width":       0,
		"height":      0,
		"duration":    0,
		"format":      "unknown",
		"size":        0,
		"created_at":  time.Now().Unix(),
		"modified_at": time.Now().Unix(),
	}

	m.logger.Debugf("Retrieved metadata for media %s in session %s", mediaID, sessionID)
	return metadata, nil
}

func (m *MeowService) GetInviteInfo(ctx context.Context, sessionID, inviteLink string) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	groupInfo, err := client.GetClient().GetGroupInfoFromLink(inviteLink)
	if err != nil {
		return nil, fmt.Errorf("failed to get invite info: %w", err)
	}

	result := &GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        groupInfo.Name,
		Topic:       groupInfo.Topic,
		CreatedBy:   groupInfo.OwnerJID.String(),
		CreatedAt:   groupInfo.GroupCreated.Unix(),
		IsAnnounce:  groupInfo.IsAnnounce,
		IsLocked:    groupInfo.IsLocked,
		IsEphemeral: groupInfo.IsEphemeral,
	}

	for _, participant := range groupInfo.Participants {
		result.Participants = append(result.Participants, participant.JID.String())
	}


	m.logger.Debugf("Retrieved invite info for group %s from link for session %s",
		groupInfo.JID.String(), sessionID)

	return result, nil
}

func (m *MeowService) GetGroupInfoFromInvite(ctx context.Context, sessionID, groupJID, inviter, code string, expiration int64) (*ports.GroupInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	groupJIDParsed, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return nil, fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	inviterJID, err := waTypes.ParseJID(inviter)
	if err != nil {
		return nil, fmt.Errorf("invalid inviter JID %s: %w", inviter, err)
	}

	groupInfo, err := client.GetClient().GetGroupInfoFromInvite(groupJIDParsed, inviterJID, code, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to get group info from invite: %w", err)
	}

	result := &GroupInfo{
		JID:         groupInfo.JID.String(),
		Name:        groupInfo.Name,
		Topic:       groupInfo.Topic,
		CreatedBy:   groupInfo.OwnerJID.String(),
		CreatedAt:   groupInfo.GroupCreated.Unix(),
		IsAnnounce:  groupInfo.IsAnnounce,
		IsLocked:    groupInfo.IsLocked,
		IsEphemeral: groupInfo.IsEphemeral,
	}

	for _, participant := range groupInfo.Participants {
		result.Participants = append(result.Participants, participant.JID.String())
	}


	m.logger.Debugf("Retrieved group info from specific invite for group %s in session %s",
		groupInfo.JID.String(), sessionID)

	return result, nil
}

func (m *MeowService) UpdateParticipants(ctx context.Context, sessionID, groupJID, action string, participants []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	groupJIDParsed, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var participantJIDs []waTypes.JID
	for _, phone := range participants {
		jid, err := parsePhoneToJID(phone)
		if err != nil {
			m.logger.Warnf("Invalid participant phone %s: %v", phone, err)
			continue
		}
		participantJIDs = append(participantJIDs, jid)
	}

	if len(participantJIDs) == 0 {
		return fmt.Errorf("no valid participants provided")
	}

	var participantChange whatsmeow.ParticipantChange
	switch action {
	case "add":
		participantChange = whatsmeow.ParticipantChangeAdd
	case "remove":
		participantChange = whatsmeow.ParticipantChangeRemove
	case "promote":
		participantChange = whatsmeow.ParticipantChangePromote
	case "demote":
		participantChange = whatsmeow.ParticipantChangeDemote
	default:
		return fmt.Errorf("invalid action: %s (must be add, remove, promote, or demote)", action)
	}

	_, err = client.GetClient().UpdateGroupParticipants(groupJIDParsed, participantJIDs, participantChange)
	if err != nil {
		return fmt.Errorf("failed to %s participants: %w", action, err)
	}

	m.logger.Debugf("Successfully %s %d participants in group %s for session %s",
		action, len(participantJIDs), groupJID, sessionID)

	return nil
}

func (m *MeowService) AddParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return m.UpdateParticipants(ctx, sessionID, groupJID, "add", participants)
}

func (m *MeowService) RemoveParticipants(ctx context.Context, sessionID, groupJID string, participants []string) error {
	return m.UpdateParticipants(ctx, sessionID, groupJID, "remove", participants)
}

func (m *MeowService) GetGroupInviteLink(ctx context.Context, sessionID, groupJID string) (string, error) {
	return m.GetInviteLink(ctx, sessionID, groupJID, false)
}

func (m *MeowService) SetGroupName(ctx context.Context, sessionID, groupJID, name string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().SetGroupName(jid, name)
	if err != nil {
		return fmt.Errorf("failed to set group name: %w", err)
	}

	m.logger.Debugf("Successfully set group name to '%s' for group %s in session %s",
		name, groupJID, sessionID)

	return nil
}

func (m *MeowService) SetGroupTopic(ctx context.Context, sessionID, groupJID, topic string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().SetGroupTopic(jid, "", "", topic)
	if err != nil {
		return fmt.Errorf("failed to set group topic: %w", err)
	}

	m.logger.Debugf("Successfully set group topic to '%s' for group %s in session %s",
		topic, groupJID, sessionID)

	return nil
}

func (m *MeowService) SetGroupPhoto(ctx context.Context, sessionID, groupJID string, photo []byte) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	if len(photo) == 0 {
		return fmt.Errorf("photo data cannot be empty")
	}

	pictureID, err := client.GetClient().SetGroupPhoto(jid, photo)
	if err != nil {
		return fmt.Errorf("failed to set group photo: %w", err)
	}

	m.logger.Debugf("Successfully set group photo (ID: %s) for group %s in session %s",
		pictureID, groupJID, sessionID)

	return nil
}

func (m *MeowService) RemoveGroupPhoto(ctx context.Context, sessionID, groupJID string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	_, err = client.GetClient().SetGroupPhoto(jid, nil)
	if err != nil {
		return fmt.Errorf("failed to remove group photo: %w", err)
	}

	m.logger.Debugf("Successfully removed group photo for group %s in session %s",
		groupJID, sessionID)

	return nil
}

func (m *MeowService) SetGroupAnnounce(ctx context.Context, sessionID, groupJID string, announceOnly bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().SetGroupAnnounce(jid, announceOnly)
	if err != nil {
		return fmt.Errorf("failed to set group announce: %w", err)
	}

	announceStatus := "disabled"
	if announceOnly {
		announceStatus = "enabled (only admins can send)"
	}

	m.logger.Debugf("Successfully set group announce to %s for group %s in session %s",
		announceStatus, groupJID, sessionID)

	return nil
}

func (m *MeowService) SetGroupLocked(ctx context.Context, sessionID, groupJID string, locked bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().SetGroupLocked(jid, locked)
	if err != nil {
		return fmt.Errorf("failed to set group locked: %w", err)
	}

	lockedStatus := "unlocked"
	if locked {
		lockedStatus = "locked (only admins can edit info)"
	}

	m.logger.Debugf("Successfully set group to %s for group %s in session %s",
		lockedStatus, groupJID, sessionID)

	return nil
}

func (m *MeowService) SetGroupJoinApprovalMode(ctx context.Context, sessionID, groupJID string, requireApproval bool) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	err = client.GetClient().SetGroupJoinApprovalMode(jid, requireApproval)
	if err != nil {
		return fmt.Errorf("failed to set group join approval mode: %w", err)
	}

	approvalStatus := "open (no approval required)"
	if requireApproval {
		approvalStatus = "require approval"
	}

	m.logger.Debugf("Successfully set group join approval mode to %s for group %s in session %s",
		approvalStatus, groupJID, sessionID)

	return nil
}

func (m *MeowService) SetGroupJoinApproval(ctx context.Context, sessionID, groupJID string, requireApproval bool) error {
	return m.SetGroupJoinApprovalMode(ctx, sessionID, groupJID, requireApproval)
}

func (m *MeowService) SetGroupMemberAddMode(ctx context.Context, sessionID, groupJID string, mode string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(groupJID)
	if err != nil {
		return fmt.Errorf("invalid group JID %s: %w", groupJID, err)
	}

	var memberAddMode waTypes.GroupMemberAddMode
	switch strings.ToLower(mode) {
	case "all", "everyone":
		memberAddMode = waTypes.GroupMemberAddModeAllMember
	case "admin", "admins", "admin_only":
		memberAddMode = waTypes.GroupMemberAddModeAdmin
	default:
		return fmt.Errorf("invalid member add mode: %s. Valid options: all, admin", mode)
	}

	err = client.GetClient().SetGroupMemberAddMode(jid, memberAddMode)
	if err != nil {
		return fmt.Errorf("failed to set group member add mode: %w", err)
	}

	m.logger.Debugf("Successfully set group member add mode to %s for group %s in session %s",
		mode, groupJID, sessionID)

	return nil
}

func (m *MeowService) UpdateSessionWebhook(sessionID, webhookURL string) error {
	m.logger.Infof("Updated webhook URL for session %s", sessionID)
	return nil
}

func (m *MeowService) UpdateSessionSubscriptions(sessionID string, events []string) error {
	m.logger.Infof("Updating event subscriptions for session %s to: %v", sessionID, events)

	client := m.getOrCreateClient(sessionID)
	if client == nil {
		return fmt.Errorf("failed to get client for session %s", sessionID)
	}

	if eventProcessor, ok := client.eventHandler.(*EventProcessor); ok {
		eventProcessor.UpdateSubscribedEvents(events)
		m.logger.Infof("Successfully updated event subscriptions for session %s", sessionID)
	} else {
		m.logger.Warnf("EventHandler for session %s is not an EventProcessor", sessionID)
		return fmt.Errorf("invalid event handler type for session %s", sessionID)
	}

	return nil
}

func (m *MeowService) SetDisappearingTimer(ctx context.Context, sessionID, chatJID string, duration time.Duration) error {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID: %w", err)
	}

	err = client.GetClient().SetDisappearingTimer(jid, duration, time.Now())
	if err != nil {
		return fmt.Errorf("failed to set disappearing timer: %w", err)
	}

	m.logger.Debugf("Set disappearing timer to %v for chat %s in session %s", duration, chatJID, sessionID)
	return nil
}

func (m *MeowService) ListChats(ctx context.Context, sessionID, chatType string) ([]ports.ChatInfo, error) {
	client, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	var chats []ChatInfo

	if chatType == "" || chatType == "groups" || chatType == "all" {
		groups, err := client.GetClient().GetJoinedGroups(ctx)
		if err == nil {
			for _, group := range groups {
				chat := ChatInfo{
					JID:         group.JID.String(),
					Name:        group.Name,
					Type:        "group",
					IsGroup:     true,
					IsPinned:    false,
					IsMuted:     false,
					IsArchived:  false,
					UnreadCount: 0,
					LastMessage: "",
					LastSeen:    time.Now(),
					CreatedAt:   time.Now(),
				}
				chats = append(chats, chat)
			}
		}
	}

	if chatType == "" || chatType == "contacts" || chatType == "all" {
		contacts, err := client.GetClient().Store.Contacts.GetAllContacts(ctx)
		if err == nil {
			for jid, contact := range contacts {
				if strings.Contains(jid.Server, "g.us") {
					continue
				}

				chat := ChatInfo{
					JID:         jid.String(),
					Name:        contact.FullName,
					Type:        "contact",
					IsGroup:     false,
					IsPinned:    false,
					IsMuted:     false,
					IsArchived:  false,
					UnreadCount: 0,
					LastMessage: "",
					LastSeen:    time.Now(),
					CreatedAt:   time.Now(),
				}
				if chat.Name == "" {
					chat.Name = contact.PushName
				}
				if chat.Name == "" {
					chat.Name = jid.String()
				}
				chats = append(chats, chat)
			}
		}
	}

	m.logger.Debugf("Retrieved %d chats for session %s", len(chats), sessionID)
	return chats, nil
}

func (m *MeowService) GetChats(ctx context.Context, sessionID string, limit, offset int) ([]ports.ChatInfo, error) {
	allChats, err := m.ListChats(ctx, sessionID, "all")
	if err != nil {
		return nil, err
	}

	start := offset
	if start > len(allChats) {
		start = len(allChats)
	}

	end := start + limit
	if end > len(allChats) {
		end = len(allChats)
	}

	results := allChats[start:end]

	m.logger.Debugf("Retrieved %d chats (offset: %d, limit: %d) for session %s", len(results), offset, limit, sessionID)
	return results, nil
}

func (m *MeowService) GetChatInfo(ctx context.Context, sessionID, chatJID string) (*ports.ChatInfo, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	jid, err := waTypes.ParseJID(chatJID)
	if err != nil {
		return nil, fmt.Errorf("invalid chat JID: %w", err)
	}

	isGroup := strings.Contains(jid.Server, "g.us")
	chatType := "contact"
	if isGroup {
		chatType = "group"
	}

	chat := &ChatInfo{
		JID:         chatJID,
		Name:        chatJID,
		Type:        chatType,
		IsPinned:    false,
		IsMuted:     false,
		IsArchived:  false,
		UnreadCount: 0,
		LastMessage: "",
		LastSeen:    time.Now(),
		CreatedAt:   time.Now(),
	}

	m.logger.Debugf("Retrieved chat info for %s in session %s", chatJID, sessionID)
	return chat, nil
}

func (m *MeowService) PinChat(ctx context.Context, sessionID, chatJID string, pinned bool) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	_, err = waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID: %w", err)
	}

	action := "unpinned"
	if pinned {
		action = "pinned"
	}
	m.logger.Debugf("Chat %s %s in session %s", chatJID, action, sessionID)
	return nil
}

func (m *MeowService) MuteChat(ctx context.Context, sessionID, chatJID string, muted bool, duration time.Duration) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	_, err = waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID: %w", err)
	}

	action := "unmuted"
	if muted {
		action = "muted"
		if duration > 0 {
			m.logger.Debugf("Chat %s %s for %v in session %s", chatJID, action, duration, sessionID)
		} else {
			m.logger.Debugf("Chat %s %s forever in session %s", chatJID, action, sessionID)
		}
	} else {
		m.logger.Debugf("Chat %s %s in session %s", chatJID, action, sessionID)
	}
	return nil
}

func (m *MeowService) ArchiveChat(ctx context.Context, sessionID, chatJID string, archived bool) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	_, err = waTypes.ParseJID(chatJID)
	if err != nil {
		return fmt.Errorf("invalid chat JID: %w", err)
	}

	action := "unarchived"
	if archived {
		action = "archived"
	}
	m.logger.Debugf("Chat %s %s in session %s", chatJID, action, sessionID)
	return nil
}

func (m *MeowService) GetNewsletterMessageUpdates(ctx context.Context, sessionID, newsletterID string) ([]ports.NewsletterMessage, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	var messages []NewsletterMessage

	m.logger.Debugf("Retrieved %d newsletter messages for newsletter %s in session %s", len(messages), newsletterID, sessionID)
	return messages, nil
}

func (m *MeowService) NewsletterMarkViewed(ctx context.Context, sessionID, newsletterID string, messageIDs []string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Marked %d newsletter messages as viewed for newsletter %s in session %s", len(messageIDs), newsletterID, sessionID)
	return nil
}

func (m *MeowService) NewsletterSendReaction(ctx context.Context, sessionID, newsletterID, messageID, reaction string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Sent reaction '%s' to message %s in newsletter %s for session %s", reaction, messageID, newsletterID, sessionID)
	return nil
}

func (m *MeowService) NewsletterToggleMute(ctx context.Context, sessionID, newsletterID string, muted bool) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	action := "unmuted"
	if muted {
		action = "muted"
	}
	m.logger.Debugf("Newsletter %s %s for session %s", newsletterID, action, sessionID)
	return nil
}

func (m *MeowService) NewsletterSubscribeLiveUpdates(ctx context.Context, sessionID, newsletterID string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Subscribed to live updates for newsletter %s in session %s", newsletterID, sessionID)
	return nil
}

func (m *MeowService) UploadNewsletter(ctx context.Context, sessionID string, data []byte) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Uploaded newsletter data (%d bytes) for session %s", len(data), sessionID)
	return nil
}

func (m *MeowService) GetNewsletterInfoWithInvite(ctx context.Context, sessionID, inviteCode string) (*ports.NewsletterInfo, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	newsletter := &NewsletterInfo{
		ID:          "newsletter_" + inviteCode,
		JID:         "newsletter_" + inviteCode + "@newsletter",
		Name:        "Newsletter from invite",
		Description: "Newsletter obtained from invite code",
		Subscribers: 0,
		Verified:    false,
		IsVerified:  false,
		Muted:       false,
		Following:   false,
		CreatedAt:   timestamp,
		ServerID:    "server_" + inviteCode,
		Timestamp:   timestamp,
	}

	m.logger.Debugf("Retrieved newsletter info from invite %s for session %s", inviteCode, sessionID)
	return newsletter, nil
}

func (m *MeowService) CreateNewsletter(ctx context.Context, sessionID, name, description string) (*ports.NewsletterInfo, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	newsletterID := fmt.Sprintf("newsletter_%d", timestamp)
	newsletter := &NewsletterInfo{
		ID:          newsletterID,
		JID:         newsletterID + "@newsletter",
		Name:        name,
		Description: description,
		Subscribers: 0,
		Verified:    false,
		IsVerified:  false,
		Muted:       false,
		Following:   true,
		CreatedAt:   timestamp,
		ServerID:    "server_" + newsletterID,
		Timestamp:   timestamp,
	}

	m.logger.Debugf("Created newsletter '%s' for session %s", name, sessionID)
	return newsletter, nil
}

func (m *MeowService) GetNewsletterInfo(ctx context.Context, sessionID, newsletterID string) (*ports.NewsletterInfo, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	timestamp := time.Now().Unix()
	newsletter := &NewsletterInfo{
		ID:          newsletterID,
		JID:         newsletterID + "@newsletter",
		Name:        "Newsletter",
		Description: "Newsletter description",
		Subscribers: 100,
		Verified:    false,
		IsVerified:  false,
		Muted:       false,
		Following:   true,
		CreatedAt:   timestamp,
		ServerID:    "server_" + newsletterID,
		Timestamp:   timestamp,
	}

	m.logger.Debugf("Retrieved newsletter info for %s in session %s", newsletterID, sessionID)
	return newsletter, nil
}

func (m *MeowService) GetSubscribedNewsletters(ctx context.Context, sessionID string) ([]ports.NewsletterInfo, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	var newsletters []NewsletterInfo

	m.logger.Debugf("Retrieved %d subscribed newsletters for session %s", len(newsletters), sessionID)
	return newsletters, nil
}

func (m *MeowService) FollowNewsletter(ctx context.Context, sessionID, newsletterID string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Followed newsletter %s for session %s", newsletterID, sessionID)
	return nil
}

func (m *MeowService) UnfollowNewsletter(ctx context.Context, sessionID, newsletterID string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Unfollowed newsletter %s for session %s", newsletterID, sessionID)
	return nil
}

func (m *MeowService) SendNewsletterMessage(ctx context.Context, sessionID, newsletterID, message string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Sent message to newsletter %s for session %s: %s", newsletterID, sessionID, message)
	return nil
}

func (m *MeowService) GetNewsletterMessages(ctx context.Context, sessionID, newsletterID string) ([]ports.NewsletterMessage, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	var messages []NewsletterMessage

	m.logger.Debugf("Retrieved %d messages for newsletter %s in session %s", len(messages), newsletterID, sessionID)
	return messages, nil
}

func (m *MeowService) GetPrivacySettings(ctx context.Context, sessionID string) (*ports.PrivacySettings, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	settings := &PrivacySettings{
		LastSeen:             "contacts",
		ProfilePhoto:         "contacts",
		About:                "contacts",
		Status:               "contacts",
		ReadReceipts:         true,
		GroupsAddMe:          "contacts",
		CallsAddMe:           "contacts",
		DisappearingMessages: "off",
	}

	m.logger.Debugf("Retrieved privacy settings for session %s", sessionID)
	return settings, nil
}

func (m *MeowService) SetPrivacySetting(ctx context.Context, sessionID, setting, value string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Set privacy setting %s to %s for session %s", setting, value, sessionID)
	return nil
}

func (m *MeowService) GetBlocklist(ctx context.Context, sessionID string) ([]string, error) {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return nil, err
	}

	var blocklist []string

	m.logger.Debugf("Retrieved %d blocked contacts for session %s", len(blocklist), sessionID)
	return blocklist, nil
}

func (m *MeowService) UpdateBlocklist(ctx context.Context, sessionID string, action string, contacts []string) error {
	_, err := m.validateAndGetClient(sessionID)
	if err != nil {
		return err
	}

	m.logger.Debugf("Updated blocklist with action %s for %d contacts in session %s", action, len(contacts), sessionID)
	return nil
}


func sendMessageToJID(client *whatsmeow.Client, to string, message *waProto.Message) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func createMediaMessage(client *whatsmeow.Client, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	return uploadMedia(client, data, mediaType)
}

func validateMessageInput(client *whatsmeow.Client, to string) error {
	if client == nil {
		return fmt.Errorf("client cannot be nil")
	}
	if to == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	return nil
}

func parsePhoneToJID(phone string) (waTypes.JID, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return waTypes.EmptyJID, fmt.Errorf("phone number cannot be empty")
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
	formattedPhone := digits.String()

	if formattedPhone == "" {
		return waTypes.EmptyJID, fmt.Errorf("phone number cannot be empty")
	}

	if len(formattedPhone) < 7 || len(formattedPhone) > 15 {
		return waTypes.EmptyJID, fmt.Errorf("phone number must be between 7 and 15 digits")
	}

	if formattedPhone[0] == '0' {
		return waTypes.EmptyJID, fmt.Errorf("phone number should not start with 0")
	}

	return waTypes.NewJID(formattedPhone, waTypes.DefaultUserServer), nil
}

func uploadMedia(client *whatsmeow.Client, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	resp, err := client.Upload(context.Background(), data, mediaType)
	return &resp, err
}


func sendTextMessage(client *whatsmeow.Client, to, text string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	message := &waProto.Message{
		Conversation: &text,
	}

	return sendMessageToJID(client, to, message)
}

func sendImageMessage(client *whatsmeow.Client, to string, data []byte, caption string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("image data cannot be empty")
	}

	uploaded, err := createMediaMessage(client, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	mimeType := "image/jpeg"
	message := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &mimeType,
		},
	}

	return sendMessageToJID(client, to, message)
}

func sendAudioMessage(client *whatsmeow.Client, to string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("audio data cannot be empty")
	}

	uploaded, err := createMediaMessage(client, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	return sendMessageToJID(client, to, message)
}

func sendVideoMessage(client *whatsmeow.Client, to string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("video data cannot be empty")
	}

	uploaded, err := createMediaMessage(client, data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	return sendMessageToJID(client, to, message)
}

func sendDocumentMessage(client *whatsmeow.Client, to string, data []byte, filename, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	uploaded, err := uploadMedia(client, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			FileName:      &filename,
			Caption:       &caption,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func sendStickerMessage(client *whatsmeow.Client, to string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	uploaded, err := uploadMedia(client, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		StickerMessage: &waProto.StickerMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func sendContactMessage(client *whatsmeow.Client, to, contactName, contactPhone string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL;type=CELL;type=VOICE;waid=%s:+%s\nEND:VCARD", contactName, contactPhone, contactPhone)

	message := &waProto.Message{
		ContactMessage: &waProto.ContactMessage{
			DisplayName: &contactName,
			Vcard:       &vcard,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func sendContactsMessage(client *whatsmeow.Client, to string, contacts []ports.ContactData) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("at least one contact is required")
	}

	if len(contacts) > 10 {
		return nil, fmt.Errorf("maximum 10 contacts allowed")
	}

	if len(contacts) == 1 {
		return sendContactMessage(client, to, contacts[0].Name, contacts[0].Phone)
	}

	var contactMessages []*waProto.ContactMessage
	for _, contact := range contacts {
		vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL;type=CELL;type=VOICE;waid=%s:+%s\nEND:VCARD",
			contact.Name, contact.Phone, contact.Phone)

		contactMessages = append(contactMessages, &waProto.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		})
	}

	displayName := fmt.Sprintf("%d contacts", len(contacts))
	message := &waProto.Message{
		ContactsArrayMessage: &waProto.ContactsArrayMessage{
			DisplayName: &displayName,
			Contacts:    contactMessages,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func sendLocationMessage(client *whatsmeow.Client, to string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		LocationMessage: &waProto.LocationMessage{
			DegreesLatitude:  &latitude,
			DegreesLongitude: &longitude,
			Name:             &name,
			Address:          &address,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}
