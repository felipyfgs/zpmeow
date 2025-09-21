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
	sessionRepo      session.Repository
	logger           logging.Logger
	subscribedEvents []string

	// Receipt batching for performance
	receiptMutex     sync.Mutex
	receiptCount     int
	lastReceiptLog   time.Time
}

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
		subscribedEvents: []string{},
	}

	ep.loadSubscribedEvents()

	return ep
}

func (ep *EventProcessor) loadSubscribedEvents() {
	sessionEntity, err := ep.sessionRepo.GetByID(context.Background(), ep.sessionID)
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

func (ep *EventProcessor) isSubscribedToEvent(eventType string) bool {
	if len(ep.subscribedEvents) == 0 {
		return false
	}

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

	systemEventType, exists := eventTypeMapping[eventType]
	if !exists {
		// Only log unmapped events that might be important
		if !isCommonUnmappedEvent(eventType) {
			ep.logger.Debugf("Unmapped event: %s", eventType)
		}
		return
	}

	if !ep.isSubscribedToEvent(systemEventType) {
		return
	}

	if handler, exists := eventHandlers[eventType]; exists {
		handler(ep, evt)
	} else {
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
		"data":       msg,
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
		"data":       evt,
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
		"data":       evt,
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
		"data":       qr,
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
		"data":       evt,
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
		"data":       pairError,
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
		"data":       evt,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleReceipt(evt interface{}) {
	receipt := evt.(*events.Receipt)

	// Batch receipt logs to reduce verbosity
	ep.receiptMutex.Lock()
	ep.receiptCount++
	now := time.Now()

	// Log every 10 receipts or every 30 seconds, whichever comes first
	shouldLog := ep.receiptCount%10 == 0 || now.Sub(ep.lastReceiptLog) > 30*time.Second
	if shouldLog {
		ep.logger.Debugf("Processed %d receipts for session %s (last 30s)", ep.receiptCount, ep.sessionID)
		ep.lastReceiptLog = now
		ep.receiptCount = 0
	}
	ep.receiptMutex.Unlock()

	webhookPayload := map[string]interface{}{
		"event":      "Receipt",
		"session_id": ep.sessionID,
		"timestamp":  now.Unix(),
		"data":       receipt,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send receipt webhook: %v", err)
	}
}

func (ep *EventProcessor) handlePresence(evt interface{}) {
	presence := evt.(*events.Presence)
	// Removed verbose presence logging - too frequent

	webhookPayload := map[string]interface{}{
		"event":      "Presence",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       presence,
	}

	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send webhook: %v", err)
	}
}

func (ep *EventProcessor) handleChatPresence(evt interface{}) {
	chatPresence := evt.(*events.ChatPresence)
	// Removed verbose chat presence logging - too frequent

	webhookPayload := map[string]interface{}{
		"event":      "ChatPresence",
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       chatPresence,
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
	webhookPayload := map[string]interface{}{
		"event":      eventType,
		"session_id": ep.sessionID,
		"timestamp":  time.Now().Unix(),
		"data":       evt,
	}

	ep.logger.Infof("Sending generic event: %s", eventType)
	if err := sendWebhook(ep.webhookURL, webhookPayload); err != nil {
		ep.logger.Errorf("Failed to send generic webhook: %v", err)
	}
}

// isCommonUnmappedEvent checks if an event type is commonly unmapped and can be ignored
func isCommonUnmappedEvent(eventType string) bool {
	commonUnmappedEvents := map[string]bool{
		"*events.QR":                    true,
		"*events.PairSuccess":           true,
		"*events.PairError":             true,
		"*events.LoggedOut":             true,
		"*events.StreamReplaced":        true,
		"*events.TemporaryBan":          true,
		"*events.ConnectFailure":        true,
		"*events.ClientOutdated":        true,
		"*events.KeepAliveTimeout":      true,
		"*events.KeepAliveRestored":     true,
		"*events.Blocklist":             true,
		"*events.PushName":              true,
		"*events.BusinessName":          true,
		"*events.JoinedGroup":           true,
		"*events.GroupInfo":             true,
		"*events.Picture":               true,
		"*events.PushNameSetting":       true,
		"*events.AppStateSyncComplete":  true,
		"*events.HistorySync":           true,
		"*events.AppState":              true,
		"*events.MarkChatAsRead":        true,
		"*events.Mute":                  true,
		"*events.Pin":                   true,
		"*events.Star":                  true,
		"*events.Archive":               true,
		"*events.DeleteChat":            true,
		"*events.UndoDeleteChat":        true,
		"*events.DeleteForMe":           true,
		"*events.MediaRetry":            true,
		"*events.UndecryptableMessage":  true,
	}
	return commonUnmappedEvents[eventType]
}

func sendWebhook(url string, data interface{}) error {
	if url == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return globalWebhookService.SendWebhook(ctx, url, "whatsapp_event", "", data)
}
