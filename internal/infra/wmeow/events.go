package wmeow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"
	"zpmeow/internal/infra/webhooks"

	"go.mau.fi/whatsmeow/types/events"
)

type EventProcessor struct {
	sessionID        string
	webhookURL       string
	sessionManager   *sessionManager
	logger           logging.Logger
	subscribedEvents []string

	receiptMutex   sync.Mutex
	receiptCount   int
	lastReceiptLog time.Time
}

var eventTypeMapping = map[string]string{
	"*events.Message":              "Message",
	"*events.UndecryptableMessage": "UndecryptableMessage",
	"*events.Receipt":              "Receipt",
	"*events.MediaRetry":           "MediaRetry",
	"*events.MediaRetryError":      "MediaRetryError",

	"*events.GroupInfo":   "GroupInfo",
	"*events.JoinedGroup": "JoinedGroup",

	"*events.Contact":         "Contact",
	"*events.Picture":         "Picture",
	"*events.BusinessName":    "BusinessName",
	"*events.PushName":        "PushName",
	"*events.PushNameSetting": "PushNameSetting",

	"*events.Archive":        "Archive",
	"*events.Pin":            "Pin",
	"*events.Mute":           "Mute",
	"*events.Star":           "Star",
	"*events.DeleteChat":     "DeleteChat",
	"*events.ClearChat":      "ClearChat",
	"*events.DeleteForMe":    "DeleteForMe",
	"*events.MarkChatAsRead": "MarkChatAsRead",

	"*events.Blocklist":       "Blocklist",
	"*events.BlocklistChange": "BlocklistChange",

	"*events.LabelAssociationChat":    "LabelAssociationChat",
	"*events.LabelAssociationMessage": "LabelAssociationMessage",
	"*events.LabelEdit":               "LabelEdit",

	"*events.Connected":         "Connected",
	"*events.Disconnected":      "Disconnected",
	"*events.ConnectFailure":    "ConnectFailure",
	"*events.KeepAliveRestored": "KeepAliveRestored",
	"*events.KeepAliveTimeout":  "KeepAliveTimeout",
	"*events.LoggedOut":         "LoggedOut",
	"*events.ClientOutdated":    "ClientOutdated",
	"*events.TemporaryBan":      "TemporaryBan",
	"*events.StreamError":       "StreamError",
	"*events.StreamReplaced":    "StreamReplaced",

	"*events.PairSuccess":                 "PairSuccess",
	"*events.PairError":                   "PairError",
	"*events.QR":                          "QR",
	"*events.QRScannedWithoutMultidevice": "QRScannedWithoutMultidevice",

	"*events.PrivacySettings":       "PrivacySettings",
	"*events.UserAbout":             "UserAbout",
	"*events.UnarchiveChatsSetting": "UnarchiveChatsSetting",
	"*events.UserStatusMute":        "UserStatusMute",

	"*events.AppState":             "AppState",
	"*events.AppStateSyncComplete": "AppStateSyncComplete",
	"*events.HistorySync":          "HistorySync",
	"*events.OfflineSyncCompleted": "OfflineSyncCompleted",
	"*events.OfflineSyncPreview":   "OfflineSyncPreview",

	"*events.CallOffer":        "CallOffer",
	"*events.CallAccept":       "CallAccept",
	"*events.CallTerminate":    "CallTerminate",
	"*events.CallOfferNotice":  "CallOfferNotice",
	"*events.CallRelayLatency": "CallRelayLatency",
	"*events.CallPreAccept":    "CallPreAccept",
	"*events.CallReject":       "CallReject",
	"*events.CallTransport":    "CallTransport",
	"*events.UnknownCallEvent": "UnknownCallEvent",

	"*events.Presence":     "Presence",
	"*events.ChatPresence": "ChatPresence",

	"*events.IdentityChange":  "IdentityChange",
	"*events.CATRefreshError": "CATRefreshError",

	"*events.NewsletterJoin":        "NewsletterJoin",
	"*events.NewsletterLeave":       "NewsletterLeave",
	"*events.NewsletterMuteChange":  "NewsletterMuteChange",
	"*events.NewsletterLiveUpdate":  "NewsletterLiveUpdate",
	"*events.NewsletterMessageMeta": "NewsletterMessageMeta",

	"*events.FBMessage": "FBMessage",

	"*events.ManualLoginReconnect": "ManualLoginReconnect",
}

var eventHandlers = map[string]func(*EventProcessor, interface{}){
	"*events.Message": (*EventProcessor).handleMessage,
	"*events.Receipt": (*EventProcessor).handleReceipt,

	"*events.Connected":    (*EventProcessor).handleConnected,
	"*events.Disconnected": (*EventProcessor).handleDisconnected,
	"*events.LoggedOut":    (*EventProcessor).handleLoggedOut,

	"*events.QR":          (*EventProcessor).handleQR,
	"*events.PairSuccess": (*EventProcessor).handlePairSuccess,
	"*events.PairError":   (*EventProcessor).handlePairError,

	"*events.Presence":     (*EventProcessor).handlePresence,
	"*events.ChatPresence": (*EventProcessor).handleChatPresence,
}

func NewEventProcessor(sessionID, webhookURL string, sessionRepo session.Repository) *EventProcessor {
	logger := logging.GetLogger().Sub("events").Sub(sessionID)
	ep := &EventProcessor{
		sessionID:        sessionID,
		webhookURL:       webhookURL,
		sessionManager:   NewSessionManager(sessionRepo, logger),
		logger:           logger,
		subscribedEvents: []string{},
	}

	ep.loadSubscribedEvents()

	return ep
}

func (ep *EventProcessor) shouldProcessEvent(eventType string) bool {
	if len(ep.subscribedEvents) == 0 {
		return true // Process all events if none specified
	}

	for _, subscribedEvent := range ep.subscribedEvents {
		if subscribedEvent == "All" || subscribedEvent == eventType {
			return true
		}
	}
	return false
}

func (ep *EventProcessor) logEventWithThrottling(eventType string, details string) {
	if eventType == "Receipt" {
		ep.receiptMutex.Lock()
		ep.receiptCount++
		now := time.Now()

		if now.Sub(ep.lastReceiptLog) > 30*time.Second {
			ep.logger.Debugf("Processed %d receipt events in last 30s", ep.receiptCount)
			ep.receiptCount = 0
			ep.lastReceiptLog = now
		}
		ep.receiptMutex.Unlock()
		return
	}

	ep.logger.Debugf("Event %s: %s", eventType, details)
}

func (ep *EventProcessor) loadSubscribedEvents() {
	sessionEntity, err := ep.sessionManager.GetSession(ep.sessionID)
	if err != nil {
		ep.logger.Warnf("Failed to load session for events: %v", err)
		ep.subscribedEvents = []string{"All"}
		ep.logger.Infof("Loaded default subscribed events: %v", ep.subscribedEvents)
		return
	}

	events := sessionEntity.GetWebhookEvents()
	if len(events) > 0 {
		ep.subscribedEvents = events
	} else {
		ep.subscribedEvents = []string{"All"}
	}

	ep.logger.Infof("Loaded subscribed events: %v", ep.subscribedEvents)
}

func (ep *EventProcessor) UpdateSubscribedEvents(events []string) {
	ep.subscribedEvents = events
	ep.logger.Infof("Updated subscribed events: %v", ep.subscribedEvents)
}

func (ep *EventProcessor) UpdateWebhookURL(webhookURL string) {
	ep.webhookURL = webhookURL
	ep.logger.Infof("Updated webhook URL: %s", webhookURL)
}

func (ep *EventProcessor) HandleEvent(evt interface{}) {
	eventType := fmt.Sprintf("%T", evt)

	systemEventType, exists := eventTypeMapping[eventType]
	if !exists {
		if !isCommonUnmappedEvent(eventType) {
			ep.logger.Debugf("Unmapped event: %s", eventType)
		}
		return
	}

	if !ep.shouldProcessEvent(systemEventType) {
		return
	}

	ep.logEventWithThrottling(systemEventType, fmt.Sprintf("Processing %s", systemEventType))

	if handler, exists := eventHandlers[eventType]; exists {
		handler(ep, evt)
	} else {
		ep.sendGenericEvent(systemEventType, evt)
	}
}

func (ep *EventProcessor) handleMessage(evt interface{}) {
	msg := evt.(*events.Message)
	ep.logger.Infof("Message received from %s in session %s", msg.Info.Sender, ep.sessionID)

	if ep.webhookURL == "" {
		ep.logger.Warnf("No webhook URL configured for session %s, skipping Message event", ep.sessionID)
		return
	}

	normalizedMsg := ep.normalizeMessage(msg)

	webhookPayload := map[string]interface{}{
		"event":     "Message",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      normalizedMsg,
	}

	ep.logger.Infof("Sending Message event to webhook: %s", ep.webhookURL)
	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send Message webhook: %v", err)
	} else {
		ep.logger.Infof("Successfully sent Message event")
	}
}

func (ep *EventProcessor) normalizeMessage(msg *events.Message) *events.Message {
	normalizedMsg := *msg

	if msg.Message.ExtendedTextMessage != nil && msg.Message.ExtendedTextMessage.Text != nil {
		text := *msg.Message.ExtendedTextMessage.Text

		ep.logger.Debugf("Normalizing extendedTextMessage to conversation format for session %s: %s", ep.sessionID, text)

		normalizedMsg.Message.Conversation = &text
		normalizedMsg.Message.ExtendedTextMessage = nil

		if normalizedMsg.RawMessage != nil {
			if normalizedMsg.RawMessage.DeviceSentMessage != nil &&
				normalizedMsg.RawMessage.DeviceSentMessage.Message != nil {
				if normalizedMsg.RawMessage.DeviceSentMessage.Message.ExtendedTextMessage != nil {
					normalizedMsg.RawMessage.DeviceSentMessage.Message.Conversation = &text
					normalizedMsg.RawMessage.DeviceSentMessage.Message.ExtendedTextMessage = nil
				}
			}

			if normalizedMsg.RawMessage.ExtendedTextMessage != nil {
				normalizedMsg.RawMessage.Conversation = &text
				normalizedMsg.RawMessage.ExtendedTextMessage = nil
			}
		}
	}

	return &normalizedMsg
}

func (ep *EventProcessor) handleConnected(evt interface{}) {
	webhookPayload := map[string]interface{}{
		"event":     "Connected",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      evt,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleDisconnected(evt interface{}) {
	webhookPayload := map[string]interface{}{
		"event":     "Disconnected",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      evt,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleQR(evt interface{}) {
	qr := evt.(*events.QR)
	ep.logger.Infof("QR code generated for session %s", ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":     "QR",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      qr,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handlePairSuccess(evt interface{}) {
	ep.logger.Infof("Pair success for session %s", ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":     "PairSuccess",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      evt,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handlePairError(evt interface{}) {
	pairError := evt.(*events.PairError)
	ep.logger.Errorf("Pair error for session %s: %v", ep.sessionID, pairError.Error)

	webhookPayload := map[string]interface{}{
		"event":     "PairError",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      pairError,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleLoggedOut(evt interface{}) {
	webhookPayload := map[string]interface{}{
		"event":     "LoggedOut",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      evt,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleReceipt(evt interface{}) {
	receipt := evt.(*events.Receipt)

	ep.receiptMutex.Lock()
	ep.receiptCount++
	now := time.Now()

	shouldLog := ep.receiptCount%10 == 0 || now.Sub(ep.lastReceiptLog) > 30*time.Second
	if shouldLog {
		ep.logger.Debugf("Processed %d receipts for session %s (last 30s)", ep.receiptCount, ep.sessionID)
		ep.lastReceiptLog = now
		ep.receiptCount = 0
	}
	ep.receiptMutex.Unlock()

	webhookPayload := map[string]interface{}{
		"event":     "Receipt",
		"sessionID": ep.sessionID,
		"timestamp": now.Unix(),
		"data":      receipt,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send receipt webhook: %v", err)
	}
}

func (ep *EventProcessor) handlePresence(evt interface{}) {
	presence := evt.(*events.Presence)

	webhookPayload := map[string]interface{}{
		"event":     "Presence",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      presence,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleChatPresence(evt interface{}) {
	chatPresence := evt.(*events.ChatPresence)

	webhookPayload := map[string]interface{}{
		"event":     "ChatPresence",
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      chatPresence,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

var globalWebhookService *webhooks.Service

func init() {
	globalWebhookService = webhooks.NewService()
}

func (ep *EventProcessor) sendGenericEvent(eventType string, evt interface{}) {
	if ep.webhookURL == "" {
		ep.logger.Warnf("No webhook URL configured for session %s, skipping event %s", ep.sessionID, eventType)
		return
	}

	webhookPayload := map[string]interface{}{
		"event":     eventType,
		"sessionID": ep.sessionID,
		"timestamp": time.Now().Unix(),
		"data":      evt,
	}

	ep.logger.Infof("Sending generic event: %s to webhook: %s", eventType, ep.webhookURL)
	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send generic webhook for event %s: %v", eventType, err)
	} else {
		ep.logger.Infof("Successfully sent generic event: %s", eventType)
	}
}

func isCommonUnmappedEvent(eventType string) bool {
	commonUnmappedEvents := map[string]bool{
		"*events.QR":                   true,
		"*events.PairSuccess":          true,
		"*events.PairError":            true,
		"*events.LoggedOut":            true,
		"*events.StreamReplaced":       true,
		"*events.TemporaryBan":         true,
		"*events.ConnectFailure":       true,
		"*events.ClientOutdated":       true,
		"*events.KeepAliveTimeout":     true,
		"*events.KeepAliveRestored":    true,
		"*events.Blocklist":            true,
		"*events.PushName":             true,
		"*events.BusinessName":         true,
		"*events.JoinedGroup":          true,
		"*events.GroupInfo":            true,
		"*events.Picture":              true,
		"*events.PushNameSetting":      true,
		"*events.AppStateSyncComplete": true,
		"*events.HistorySync":          true,
		"*events.AppState":             true,
		"*events.MarkChatAsRead":       true,
		"*events.Mute":                 true,
		"*events.Pin":                  true,
		"*events.Star":                 true,
		"*events.Archive":              true,
		"*events.DeleteChat":           true,
		"*events.UndoDeleteChat":       true,
		"*events.DeleteForMe":          true,
		"*events.MediaRetry":           true,
		"*events.UndecryptableMessage": true,
	}
	return commonUnmappedEvents[eventType]
}

func sendWebhook(url string, data interface{}) error {
	if url == "" {
		return fmt.Errorf("webhook URL is empty")
	}

	logger := logging.GetLogger().Sub("webhook-sender")
	logger.Infof("Attempting to send webhook to: %s", url)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := globalWebhookService.SendWebhook(ctx, url, "whatsapp_event", "", data)
	if err != nil {
		logger.Errorf("Failed to send webhook to %s: %v", url, err)
		return err
	}

	logger.Infof("Successfully sent webhook to: %s", url)
	return nil
}
