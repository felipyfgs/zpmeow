package dto

import (
	"fmt"
	"time"
)

type PrivacySettingsResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message,omitempty"`
	Data    *PrivacySettingsData  `json:"data,omitempty"`
	Error   *PrivacyErrorResponse `json:"error,omitempty"`
}

type PrivacySettingsData struct {
	GroupAdd     string `json:"groupAdd"`     // Who can add to groups: "all", "contacts", "contact_blacklist", "none"
	LastSeen     string `json:"lastSeen"`     // Who can see last seen: "all", "contacts", "contact_blacklist", "none"
	Status       string `json:"status"`       // Who can see status: "all", "contacts", "contact_blacklist", "none"
	Profile      string `json:"profile"`      // Who can see profile photo: "all", "contacts", "contact_blacklist", "none"
	ReadReceipts string `json:"readReceipts"` // Read receipts: "all", "none"
	CallAdd      string `json:"callAdd"`      // Who can call: "all", "known"
	Online       string `json:"online"`       // Who can see online status: "all", "match_last_seen"
}

type SetPrivacySettingRequest struct {
	Setting string `json:"setting" binding:"required"` // "groupadd", "last", "status", "profile", "readreceipts", "calladd", "online"
	Value   string `json:"value" binding:"required"`   // Depends on setting type
}

type PrivacyErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

type SetGroupAddPrivacyRequest struct {
	Value string `json:"value" binding:"required,oneof=all contacts contact_blacklist none"` // Who can add to groups
}

type SetLastSeenPrivacyRequest struct {
	Value string `json:"value" binding:"required,oneof=all contacts contact_blacklist none"` // Who can see last seen
}

type SetProfilePrivacyRequest struct {
	Value string `json:"value" binding:"required,oneof=all contacts contact_blacklist none"` // Who can see profile
}

type SetReadReceiptsPrivacyRequest struct {
	Enabled bool `json:"enabled"` // Whether read receipts are enabled
}

type SetCallAddPrivacyRequest struct {
	Value string `json:"value" binding:"required,oneof=all known"` // Who can call
}

type SetOnlinePrivacyRequest struct {
	Value string `json:"value" binding:"required,oneof=all match_last_seen"` // Who can see online status
}

type SetAllPrivacySettingsRequest struct {
	GroupAdd     *string `json:"groupAdd,omitempty"`     // Who can add to groups: "all", "contacts", "contact_blacklist", "none"
	LastSeen     *string `json:"lastSeen,omitempty"`     // Who can see last seen: "all", "contacts", "contact_blacklist", "none"
	Status       *string `json:"status,omitempty"`       // Who can see status: "all", "contacts", "contact_blacklist", "none"
	Profile      *string `json:"profile,omitempty"`      // Who can see profile photo: "all", "contacts", "contact_blacklist", "none"
	ReadReceipts *bool   `json:"readReceipts,omitempty"` // Read receipts enabled/disabled
	CallAdd      *string `json:"callAdd,omitempty"`      // Who can call: "all", "known"
	Online       *string `json:"online,omitempty"`       // Who can see online status: "all", "match_last_seen"
}

type FindPrivacySettingsRequest struct {
	Settings []string `json:"settings,omitempty"` // Specific settings to retrieve: ["groupAdd", "lastSeen", "status", "profile", "readReceipts", "callAdd", "online"]
}

var ValidPrivacySettings = map[string][]string{
	"groupadd":     {"all", "contacts", "contact_blacklist", "none"},
	"last":         {"all", "contacts", "contact_blacklist", "none"},
	"status":       {"all", "contacts", "contact_blacklist", "none"},
	"profile":      {"all", "contacts", "contact_blacklist", "none"},
	"readreceipts": {"all", "none"},
	"calladd":      {"all", "known"},
	"online":       {"all", "match_last_seen"},
}

var PrivacySettingDescriptions = map[string]string{
	"groupadd":     "Who can add you to groups",
	"last":         "Who can see your last seen",
	"status":       "Who can see your status",
	"profile":      "Who can see your profile photo",
	"readreceipts": "Read receipts setting",
	"calladd":      "Who can call you",
	"online":       "Who can see when you're online",
}

var PrivacyValueDescriptions = map[string]string{
	"all":               "Everyone",
	"contacts":          "My contacts",
	"contact_blacklist": "My contacts except...",
	"none":              "Nobody",
	"known":             "Known contacts",
	"match_last_seen":   "Same as last seen",
}

func IsValidPrivacySetting(setting, value string) bool {
	validValues, exists := ValidPrivacySettings[setting]
	if !exists {
		return false
	}

	for _, validValue := range validValues {
		if validValue == value {
			return true
		}
	}
	return false
}

func GetPrivacySettingDescription(setting string) string {
	if desc, exists := PrivacySettingDescriptions[setting]; exists {
		return desc
	}
	return "Unknown privacy setting"
}

func GetPrivacyValueDescription(value string) string {
	if desc, exists := PrivacyValueDescriptions[value]; exists {
		return desc
	}
	return "Unknown privacy value"
}

type PrivacySettingsUpdateEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Setting   string    `json:"setting"`
	OldValue  string    `json:"oldValue"`
	NewValue  string    `json:"newValue"`
	UpdatedBy string    `json:"updatedBy"` // "user" or "system"
}

type BulkPrivacyUpdateRequest struct {
	Settings map[string]string `json:"settings" binding:"required"` // Map of setting -> value
	Force    bool              `json:"force"`                       // Force update even if some settings are invalid
}

type BulkPrivacyUpdateResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message,omitempty"`
	Updated []string              `json:"updated"` // Successfully updated settings
	Failed  []string              `json:"failed"`  // Failed to update settings
	Errors  map[string]string     `json:"errors"`  // Error messages for failed settings
	Data    *PrivacySettingsData  `json:"data,omitempty"`
	Error   *PrivacyErrorResponse `json:"error,omitempty"`
}

type PrivacySettingsSummary struct {
	TotalSettings     int       `json:"totalSettings"`
	RestrictiveCount  int       `json:"restrictiveCount"` // Settings set to "none" or "contacts"
	OpenCount         int       `json:"openCount"`        // Settings set to "all"
	LastUpdated       time.Time `json:"lastUpdated"`
	MostRestrictive   string    `json:"mostRestrictive"`   // Most restrictive setting
	LeastRestrictive  string    `json:"leastRestrictive"`  // Least restrictive setting
	RecommendedAction string    `json:"recommendedAction"` // Suggested action for better privacy
}

type PrivacyRecommendation struct {
	Setting     string `json:"setting"`
	Current     string `json:"current"`
	Recommended string `json:"recommended"`
	Reason      string `json:"reason"`
	Priority    string `json:"priority"` // "high", "medium", "low"
}

type PrivacySettingsAnalysis struct {
	Summary         PrivacySettingsSummary  `json:"summary"`
	Recommendations []PrivacyRecommendation `json:"recommendations"`
	SecurityScore   int                     `json:"securityScore"` // 0-100 privacy score
	LastAnalyzed    time.Time               `json:"lastAnalyzed"`
}

type WhatsmeowPrivacySettingType string

const (
	WhatsmeowPrivacySettingTypeGroupAdd     WhatsmeowPrivacySettingType = "groupadd"
	WhatsmeowPrivacySettingTypeLastSeen     WhatsmeowPrivacySettingType = "last"
	WhatsmeowPrivacySettingTypeStatus       WhatsmeowPrivacySettingType = "status"
	WhatsmeowPrivacySettingTypeProfile      WhatsmeowPrivacySettingType = "profile"
	WhatsmeowPrivacySettingTypeReadReceipts WhatsmeowPrivacySettingType = "readreceipts"
	WhatsmeowPrivacySettingTypeOnline       WhatsmeowPrivacySettingType = "online"
	WhatsmeowPrivacySettingTypeCallAdd      WhatsmeowPrivacySettingType = "calladd"
)

type WhatsmeowPrivacySetting string

const (
	WhatsmeowPrivacySettingUndefined        WhatsmeowPrivacySetting = ""
	WhatsmeowPrivacySettingAll              WhatsmeowPrivacySetting = "all"
	WhatsmeowPrivacySettingContacts         WhatsmeowPrivacySetting = "contacts"
	WhatsmeowPrivacySettingContactBlacklist WhatsmeowPrivacySetting = "contact_blacklist"
	WhatsmeowPrivacySettingMatchLastSeen    WhatsmeowPrivacySetting = "match_last_seen"
	WhatsmeowPrivacySettingKnown            WhatsmeowPrivacySetting = "known"
	WhatsmeowPrivacySettingNone             WhatsmeowPrivacySetting = "none"
)

func (req *SetAllPrivacySettingsRequest) Validate() map[string]string {
	errors := make(map[string]string)

	if req.GroupAdd != nil && !IsValidPrivacySetting("groupadd", *req.GroupAdd) {
		errors["groupAdd"] = fmt.Sprintf("Invalid value '%s'. Valid values: %v", *req.GroupAdd, ValidPrivacySettings["groupadd"])
	}

	if req.LastSeen != nil && !IsValidPrivacySetting("last", *req.LastSeen) {
		errors["lastSeen"] = fmt.Sprintf("Invalid value '%s'. Valid values: %v", *req.LastSeen, ValidPrivacySettings["last"])
	}

	if req.Status != nil && !IsValidPrivacySetting("status", *req.Status) {
		errors["status"] = fmt.Sprintf("Invalid value '%s'. Valid values: %v", *req.Status, ValidPrivacySettings["status"])
	}

	if req.Profile != nil && !IsValidPrivacySetting("profile", *req.Profile) {
		errors["profile"] = fmt.Sprintf("Invalid value '%s'. Valid values: %v", *req.Profile, ValidPrivacySettings["profile"])
	}

	if req.CallAdd != nil && !IsValidPrivacySetting("calladd", *req.CallAdd) {
		errors["callAdd"] = fmt.Sprintf("Invalid value '%s'. Valid values: %v", *req.CallAdd, ValidPrivacySettings["calladd"])
	}

	if req.Online != nil && !IsValidPrivacySetting("online", *req.Online) {
		errors["online"] = fmt.Sprintf("Invalid value '%s'. Valid values: %v", *req.Online, ValidPrivacySettings["online"])
	}

	return errors
}

func (req *SetAllPrivacySettingsRequest) HasAnySettings() bool {
	return req.GroupAdd != nil || req.LastSeen != nil || req.Status != nil ||
		req.Profile != nil || req.ReadReceipts != nil || req.CallAdd != nil || req.Online != nil
}

type BlocklistResponse struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Data    []string              `json:"data,omitempty"`
	Error   *PrivacyErrorResponse `json:"error,omitempty"`
}

type GetBlocklistRequest struct {
}

type UpdateBlocklistRequest struct {
	JID    string `json:"jid" binding:"required"`    // JID to block/unblock
	Action string `json:"action" binding:"required"` // "block" or "unblock"
}

type BlocklistChangeEvent struct {
	SessionID string                   `json:"sessionId"`
	Event     string                   `json:"event"`     // "blocklist_changed"
	Timestamp int64                    `json:"timestamp"` // Unix timestamp
	Action    string                   `json:"action"`    // "modify" or ""
	DHash     string                   `json:"dhash"`     // Current hash
	PrevDHash string                   `json:"prevDhash"` // Previous hash
	Changes   []BlocklistChangeDetails `json:"changes"`   // List of changes
}

type BlocklistChangeDetails struct {
	JID    string `json:"jid"`    // JID that was blocked/unblocked
	Action string `json:"action"` // "block" or "unblock"
}

type PrivacySettingsChangeEvent struct {
	SessionID     string                 `json:"sessionId"`
	Event         string                 `json:"event"`         // "privacy_settings_changed"
	Timestamp     int64                  `json:"timestamp"`     // Unix timestamp
	Changes       []string               `json:"changes"`       // Human-readable list of changes
	NewSettings   map[string]interface{} `json:"newSettings"`   // Current settings
	ChangedFields map[string]bool        `json:"changedFields"` // Which fields changed
}
