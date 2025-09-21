package wmeow

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/logging"
	"zpmeow/internal/infra/webhooks"

	"go.mau.fi/whatsmeow/types/events"
)

type EventProcessor struct {
	sessionID        string
	webhookURL       string
	sessionRepo      session.Repository
	logger           logging.Logger
	subscribedEvents []string
}

// Mapping from whatsmeow event types to our system event types
var eventTypeMapping = map[string]string{
	"*events.Message":                     "Message",
	"*events.UndecryptableMessage":        "UndecryptableMessage",
	"*events.Receipt":                     "Receipt",
	"*events.MediaRetry":                  "MediaRetry",
	"*events.Connected":                   "Connected",
	"*events.Disconnected":                "Disconnected",
	"*events.ConnectFailure":              "ConnectFailure",
	"*events.KeepAliveRestored":           "KeepAliveRestored",
	"*events.KeepAliveTimeout":            "KeepAliveTimeout",
	"*events.LoggedOut":                   "LoggedOut",
	"*events.ClientOutdated":              "ClientOutdated",
	"*events.TemporaryBan":                "TemporaryBan",
	"*events.StreamError":                 "StreamError",
	"*events.StreamReplaced":              "StreamReplaced",
	"*events.PairSuccess":                 "PairSuccess",
	"*events.PairError":                   "PairError",
	"*events.QR":                          "QR",
	"*events.QRScannedWithoutMultidevice": "QRScannedWithoutMultidevice",
	"*events.PrivacySettings":             "PrivacySettings",
	"*events.PushNameSetting":             "PushNameSetting",
	"*events.UserAbout":                   "UserAbout",
	"*events.AppState":                    "AppState",
	"*events.AppStateSyncComplete":        "AppStateSyncComplete",
	"*events.HistorySync":                 "HistorySync",
	"*events.OfflineSyncCompleted":        "OfflineSyncCompleted",
	"*events.OfflineSyncPreview":          "OfflineSyncPreview",
	"*events.CallOffer":                   "CallOffer",
	"*events.CallAccept":                  "CallAccept",
	"*events.CallTerminate":               "CallTerminate",
	"*events.CallOfferNotice":             "CallOfferNotice",
	"*events.CallRelayLatency":            "CallRelayLatency",
	"*events.Presence":                    "Presence",
	"*events.ChatPresence":                "ChatPresence",
	"*events.IdentityChange":              "IdentityChange",
	"*events.NewsletterJoin":              "NewsletterJoin",
	"*events.NewsletterLeave":             "NewsletterLeave",
	"*events.NewsletterMuteChange":        "NewsletterMuteChange",
	"*events.NewsletterLiveUpdate":        "NewsletterLiveUpdate",
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
	ep := &EventProcessor{
		sessionID:        sessionID,
		webhookURL:       webhookURL,
		sessionRepo:      sessionRepo,
		logger:           logging.GetLogger().Sub("events").Sub(sessionID),
		subscribedEvents: []string{}, // Will be loaded from session
	}

	// Load subscribed events from session
	ep.loadSubscribedEvents()

	return ep
}

// loadSubscribedEvents loads the subscribed events from the session
func (ep *EventProcessor) loadSubscribedEvents() {
	sessionEntity, err := ep.sessionRepo.GetByID(context.Background(), ep.sessionID)
	if err != nil {
		ep.logger.Warnf("Failed to load session for events: %v", err)
		// Default to subscribing to all events if session can't be loaded
		ep.subscribedEvents = []string{"All"}
		ep.logger.Infof("Loaded default subscribed events: %v", ep.subscribedEvents)
		return
	}

	// Get webhook events from session
	events := sessionEntity.GetWebhookEvents()
	if len(events) > 0 {
		ep.subscribedEvents = events
	} else {
		// Default to subscribing to all events if no specific events are configured
		ep.subscribedEvents = []string{"All"}
	}

	ep.logger.Infof("Loaded subscribed events: %v", ep.subscribedEvents)
}

// UpdateSubscribedEvents updates the list of subscribed events
func (ep *EventProcessor) UpdateSubscribedEvents(events []string) {
	ep.subscribedEvents = events
	ep.logger.Infof("Updated subscribed events: %v", ep.subscribedEvents)
}

// isSubscribedToEvent checks if the processor is subscribed to a specific event
func (ep *EventProcessor) isSubscribedToEvent(eventType string) bool {
	// If no events are subscribed, don't send anything
	if len(ep.subscribedEvents) == 0 {
		return false
	}

	// Check for "All" subscription
	for _, subscribedEvent := range ep.subscribedEvents {
		if subscribedEvent == "All" {
			return true
		}
		if subscribedEvent == eventType {
			return true
		}
	}

	return false
}

func (ep *EventProcessor) HandleEvent(evt interface{}) {
	eventType := fmt.Sprintf("%T", evt)

	ep.logger.Debugf("📨 Event received: %s", eventType)

	// Map whatsmeow event type to our system event type
	systemEventType, exists := eventTypeMapping[eventType]
	if !exists {
		ep.logger.Debugf("❓ Unmapped event: %s", eventType)
		return
	}

	// Check if we're subscribed to this event
	if !ep.isSubscribedToEvent(systemEventType) {
		ep.logger.Debugf("🚫 Not subscribed to event: %s (system: %s)", eventType, systemEventType)
		return
	}

	ep.logger.Debugf("✅ Processing subscribed event: %s -> %s", eventType, systemEventType)

	if handler, exists := eventHandlers[eventType]; exists {
		handler(ep, evt)
	} else {
		// Even if we don't have a specific handler, we can still send the raw event
		ep.sendGenericEvent(systemEventType, evt)
	}
}

func (ep *EventProcessor) handleMessage(evt interface{}) {
	msg := evt.(*events.Message)
	ep.logger.Infof("Message received from %s in session %s", msg.Info.Sender, ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":      "Message",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       msg, // Raw whatsmeow message event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleConnected(evt interface{}) {
	webhookPayload := map[string]interface{}{
		"event":      "Connected",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       evt, // Raw whatsmeow connected event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleDisconnected(evt interface{}) {
	webhookPayload := map[string]interface{}{
		"event":      "Disconnected",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       evt, // Raw whatsmeow disconnected event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleQR(evt interface{}) {
	qr := evt.(*events.QR)
	ep.logger.Infof("QR code generated for session %s", ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":      "QR",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       qr, // Raw whatsmeow QR event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handlePairSuccess(evt interface{}) {
	ep.logger.Infof("Pair success for session %s", ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":      "PairSuccess",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       evt, // Raw whatsmeow pair success event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handlePairError(evt interface{}) {
	pairError := evt.(*events.PairError)
	ep.logger.Errorf("Pair error for session %s: %v", ep.sessionID, pairError.Error)

	webhookPayload := map[string]interface{}{
		"event":      "PairError",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       pairError, // Raw whatsmeow pair error event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleLoggedOut(evt interface{}) {
	webhookPayload := map[string]interface{}{
		"event":      "LoggedOut",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       evt, // Raw whatsmeow logged out event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleReceipt(evt interface{}) {
	receipt := evt.(*events.Receipt)
	ep.logger.Debugf("Receipt received for session %s", ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":      "Receipt",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       receipt, // Raw whatsmeow receipt event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handlePresence(evt interface{}) {
	presence := evt.(*events.Presence)
	ep.logger.Debugf("Presence update for session %s", ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":      "Presence",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       presence, // Raw whatsmeow presence event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleChatPresence(evt interface{}) {
	chatPresence := evt.(*events.ChatPresence)
	ep.logger.Debugf("Chat presence update for session %s", ep.sessionID)

	webhookPayload := map[string]interface{}{
		"event":      "ChatPresence",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       chatPresence, // Raw whatsmeow chat presence event
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

// Global webhook service instance
var globalWebhookService *webhooks.Service

func init() {
	globalWebhookService = webhooks.NewService()
}

// sendGenericEvent sends a generic event for events that don't have specific handlers
func (ep *EventProcessor) sendGenericEvent(eventType string, evt interface{}) {
	webhookPayload := map[string]interface{}{
		"event":      eventType,
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       evt, // Raw whatsmeow event payload
	}

	ep.logger.Infof("📤 Sending generic event: %s", eventType)
	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("❌ Failed to send generic webhook: %v", err)
	}
}

func sendWebhook(url string, data interface{}) error {
	if url == "" {
		return nil // No webhook configured
	}

	// Use the webhook service to send the webhook asynchronously
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return globalWebhookService.SendWebhook(ctx, url, "whatsapp_event", "", data)
}
