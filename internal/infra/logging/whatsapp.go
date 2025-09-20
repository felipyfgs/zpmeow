package logging

import (
	"strings"

	waLog "go.mau.fi/whatsmeow/util/log"
)

type WhatsAppLogger = waLog.Logger

type WALoggerAdapter struct {
	logger Logger
}

func (w *WALoggerAdapter) Errorf(msg string, args ...interface{}) {
	cleanMsg := w.cleanWhatsAppMessage(msg)
	w.logger.Errorf(cleanMsg, args...)
}

func (w *WALoggerAdapter) Warnf(msg string, args ...interface{}) {
	cleanMsg := w.cleanWhatsAppMessage(msg)
	w.logger.Warnf(cleanMsg, args...)
}

func (w *WALoggerAdapter) Infof(msg string, args ...interface{}) {
	if strings.HasPrefix(strings.TrimSpace(msg), "<") {
		return
	}
	cleanMsg := w.cleanWhatsAppMessage(msg)
	w.logger.Infof(cleanMsg, args...)
}

func (w *WALoggerAdapter) Debugf(msg string, args ...interface{}) {
}

func (w *WALoggerAdapter) Sub(module string) waLog.Logger {
	return &WALoggerAdapter{
		logger: w.logger.Sub(module),
	}
}

func (w *WALoggerAdapter) cleanWhatsAppMessage(msg string) string {
	msg = strings.ReplaceAll(msg, "Successfully ", "")
	msg = strings.ReplaceAll(msg, "Failed to ", "")

	replacements := map[string]string{
		"Sending message to":                     "Sending to",
		"Received message from":                  "Received from",
		"Connected to meow":                      "Connected",
		"Disconnected from meow":                 "Disconnected",
		"Logged in to meow":                      "Logged in",
		"Logged out from meow":                   "Logged out",
		"Received QR code":                       "QR code received",
		"QR code scanned":                        "QR scanned",
		"Session restored":                       "Session restored",
		"Session created":                        "Session created",
		"Connection established":                 "Connected",
		"Connection lost":                        "Disconnected",
		"Reconnecting to meow":                   "Reconnecting",
		"Reconnected to meow":                    "Reconnected",
		"Message sent successfully":              "Message sent",
		"Message delivery confirmed":             "Message delivered",
		"Message read by recipient":              "Message read",
		"Group message received":                 "Group message",
		"Private message received":               "Private message",
		"Media message received":                 "Media received",
		"Document message received":              "Document received",
		"Voice message received":                 "Voice received",
		"Video message received":                 "Video received",
		"Image message received":                 "Image received",
		"Sticker message received":               "Sticker received",
		"Location message received":              "Location received",
		"Contact message received":               "Contact received",
		"Status update received":                 "Status update",
		"Presence update received":               "Presence update",
		"Typing indicator received":              "Typing indicator",
		"Message encryption successful":          "Message encrypted",
		"Message decryption successful":          "Message decrypted",
		"Key exchange completed":                 "Key exchange done",
		"Protocol version negotiated":            "Protocol negotiated",
		"Device registration completed":          "Device registered",
		"Device verification successful":         "Device verified",
		"Backup restoration completed":           "Backup restored",
		"Backup creation successful":             "Backup created",
		"Contact synchronization completed":      "Contacts synced",
		"Group synchronization completed":        "Groups synced",
		"Chat history synchronization completed": "History synced",
		"Media download completed":               "Media downloaded",
		"Media upload completed":                 "Media uploaded",
		"Profile picture updated":                "Profile updated",
		"Status message updated":                 "Status updated",
		"Privacy settings updated":               "Privacy updated",
		"Notification settings updated":          "Notifications updated",
		"Security settings updated":              "Security updated",
		"Account settings updated":               "Account updated",
		"Application settings updated":           "Settings updated",
		"Database connection established":        "DB connected",
		"Database query executed":                "DB query",
		"Database transaction committed":         "DB committed",
		"Database transaction rolled back":       "DB rolled back",
		"Cache entry created":                    "Cache created",
		"Cache entry updated":                    "Cache updated",
		"Cache entry deleted":                    "Cache deleted",
		"Cache entry retrieved":                  "Cache hit",
		"Cache miss occurred":                    "Cache miss",
		"Memory allocation successful":           "Memory allocated",
		"Memory deallocation successful":         "Memory freed",
		"Thread creation successful":             "Thread created",
		"Thread termination successful":          "Thread terminated",
		"Process creation successful":            "Process created",
		"Process termination successful":         "Process terminated",
		"File operation completed":               "File operation",
		"Network operation completed":            "Network operation",
		"HTTP request processed":                 "HTTP processed",
		"HTTP response sent":                     "HTTP response",
		"WebSocket connection established":       "WebSocket connected",
		"WebSocket connection closed":            "WebSocket closed",
		"API call successful":                    "API call",
		"API response received":                  "API response",
		"Configuration loaded":                   "Config loaded",
		"Configuration updated":                  "Config updated",
		"Service started":                        "Service started",
		"Service stopped":                        "Service stopped",
		"Service restarted":                      "Service restarted",
		"Health check passed":                    "Health OK",
		"Health check failed":                    "Health failed",
		"Monitoring alert triggered":             "Alert triggered",
		"Monitoring alert resolved":              "Alert resolved",
		"Performance metric recorded":            "Metric recorded",
		"Performance threshold exceeded":         "Threshold exceeded",
		"Security scan completed":                "Security scan",
		"Security vulnerability detected":        "Vulnerability found",
		"Security patch applied":                 "Patch applied",
		"Audit log entry created":                "Audit logged",
		"Compliance check passed":                "Compliance OK",
		"Compliance check failed":                "Compliance failed",
	}

	for old, new := range replacements {
		msg = strings.ReplaceAll(msg, old, new)
	}

	redundantWords := []string{
		"successfully",
		"properly",
		"correctly",
		"appropriately",
		"efficiently",
		"effectively",
		"completely",
		"entirely",
		"fully",
		"totally",
		"absolutely",
		"perfectly",
		"exactly",
		"precisely",
		"accurately",
		"thoroughly",
		"comprehensively",
		"extensively",
		"intensively",
		"systematically",
		"methodically",
		"strategically",
		"tactically",
		"operationally",
		"functionally",
		"technically",
		"mechanically",
		"automatically",
		"manually",
		"dynamically",
		"statically",
		"actively",
		"passively",
		"interactively",
		"reactively",
		"proactively",
		"retrospectively",
		"prospectively",
		"concurrently",
		"simultaneously",
		"sequentially",
		"consecutively",
		"continuously",
		"constantly",
		"persistently",
		"consistently",
		"reliably",
		"dependably",
		"predictably",
		"unexpectedly",
		"surprisingly",
		"obviously",
		"clearly",
		"evidently",
		"apparently",
		"seemingly",
		"presumably",
		"supposedly",
		"allegedly",
		"reportedly",
		"potentially",
		"possibly",
		"probably",
		"likely",
		"unlikely",
		"definitely",
		"certainly",
		"surely",
		"undoubtedly",
		"unquestionably",
		"indubitably",
		"inevitably",
		"necessarily",
		"essentially",
		"basically",
		"fundamentally",
		"primarily",
		"mainly",
		"mostly",
		"generally",
		"typically",
		"usually",
		"normally",
		"ordinarily",
		"regularly",
		"frequently",
		"occasionally",
		"rarely",
		"seldom",
		"never",
		"always",
		"sometimes",
		"often",
		"usually",
		"commonly",
		"uncommonly",
		"exceptionally",
		"remarkably",
		"notably",
		"significantly",
		"considerably",
		"substantially",
		"dramatically",
		"drastically",
		"radically",
		"fundamentally",
		"critically",
		"crucially",
		"vitally",
		"importantly",
		"urgently",
		"immediately",
		"instantly",
		"promptly",
		"quickly",
		"rapidly",
		"swiftly",
		"speedily",
		"hastily",
		"hurriedly",
		"slowly",
		"gradually",
		"progressively",
		"incrementally",
		"stepwise",
		"systematically",
	}

	words := strings.Fields(msg)
	var filteredWords []string

	for _, word := range words {
		word = strings.ToLower(strings.Trim(word, ".,!?;:"))
		isRedundant := false
		for _, redundant := range redundantWords {
			if word == redundant {
				isRedundant = true
				break
			}
		}
		if !isRedundant {
			filteredWords = append(filteredWords, word)
		}
	}

	if len(filteredWords) > 0 {
		msg = strings.Join(filteredWords, " ")
	}

	if len(msg) > 0 {
		msg = strings.ToUpper(string(msg[0])) + msg[1:]
	}

	return SanitizeMessage(msg)
}

func NewWALogger(module string) waLog.Logger {
	return &WALoggerAdapter{
		logger: GetLogger().Sub(module),
	}
}
