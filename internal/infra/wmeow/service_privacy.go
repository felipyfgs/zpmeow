package wmeow

import (
	"context"
	"fmt"

	"zpmeow/internal/application/ports"

	"go.mau.fi/whatsmeow"
	waTypes "go.mau.fi/whatsmeow/types"
)

// PrivacyManager methods - gestão de configurações de privacidade

func (m *MeowService) SetAllPrivacySettings(ctx context.Context, sessionID string, settings ports.PrivacySettings) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	// Set Last Seen privacy
	if settings.LastSeen != "" {
		lastSeenSetting := convertPrivacySettingToWhatsmeow(settings.LastSeen)
		err := client.GetClient().SetPrivacySetting(whatsmeow.PrivacySettingLastSeen, lastSeenSetting)
		if err != nil {
			m.logger.Warnf("Failed to set last seen privacy for session %s: %v", sessionID, err)
		} else {
			m.logger.Debugf("Set last seen privacy to %s for session %s", settings.LastSeen, sessionID)
		}
	}

	// Set Online privacy
	if settings.Online != "" {
		onlineSetting := convertPrivacySettingToWhatsmeow(settings.Online)
		err := client.GetClient().SetPrivacySetting(whatsmeow.PrivacySettingOnline, onlineSetting)
		if err != nil {
			m.logger.Warnf("Failed to set online privacy for session %s: %v", sessionID, err)
		} else {
			m.logger.Debugf("Set online privacy to %s for session %s", settings.Online, sessionID)
		}
	}

	// Set Profile Photo privacy
	if settings.ProfilePhoto != "" {
		profilePhotoSetting := convertPrivacySettingToWhatsmeow(settings.ProfilePhoto)
		err := client.GetClient().SetPrivacySetting(whatsmeow.PrivacySettingProfilePhoto, profilePhotoSetting)
		if err != nil {
			m.logger.Warnf("Failed to set profile photo privacy for session %s: %v", sessionID, err)
		} else {
			m.logger.Debugf("Set profile photo privacy to %s for session %s", settings.ProfilePhoto, sessionID)
		}
	}

	// Set Status privacy
	if settings.Status != "" {
		statusSetting := convertPrivacySettingToWhatsmeow(settings.Status)
		err := client.GetClient().SetPrivacySetting(whatsmeow.PrivacySettingStatus, statusSetting)
		if err != nil {
			m.logger.Warnf("Failed to set status privacy for session %s: %v", sessionID, err)
		} else {
			m.logger.Debugf("Set status privacy to %s for session %s", settings.Status, sessionID)
		}
	}

	// Set Read Receipts
	if settings.ReadReceipts != "" {
		readReceiptsSetting := convertPrivacySettingToWhatsmeow(settings.ReadReceipts)
		err := client.GetClient().SetPrivacySetting(whatsmeow.PrivacySettingReadReceipts, readReceiptsSetting)
		if err != nil {
			m.logger.Warnf("Failed to set read receipts privacy for session %s: %v", sessionID, err)
		} else {
			m.logger.Debugf("Set read receipts privacy to %s for session %s", settings.ReadReceipts, sessionID)
		}
	}

	// Set Groups privacy
	if settings.Groups != "" {
		groupsSetting := convertPrivacySettingToWhatsmeow(settings.Groups)
		err := client.GetClient().SetPrivacySetting(whatsmeow.PrivacySettingGroupAdd, groupsSetting)
		if err != nil {
			m.logger.Warnf("Failed to set groups privacy for session %s: %v", sessionID, err)
		} else {
			m.logger.Debugf("Set groups privacy to %s for session %s", settings.Groups, sessionID)
		}
	}

	// Set Calls privacy
	if settings.Calls != "" {
		callsSetting := convertPrivacySettingToWhatsmeow(settings.Calls)
		err := client.GetClient().SetPrivacySetting(whatsmeow.PrivacySettingCall, callsSetting)
		if err != nil {
			m.logger.Warnf("Failed to set calls privacy for session %s: %v", sessionID, err)
		} else {
			m.logger.Debugf("Set calls privacy to %s for session %s", settings.Calls, sessionID)
		}
	}

	m.logger.Infof("Privacy settings updated for session %s", sessionID)
	return nil
}

func (m *MeowService) GetPrivacySettings(ctx context.Context, sessionID string) (*ports.PrivacySettings, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	settings, err := client.GetClient().TryFetchPrivacySettings(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	result := &ports.PrivacySettings{
		LastSeen:     convertWhatsmeowPrivacySettingToString(settings.LastSeen),
		Online:       convertWhatsmeowPrivacySettingToString(settings.Online),
		ProfilePhoto: convertWhatsmeowPrivacySettingToString(settings.ProfilePhoto),
		Status:       convertWhatsmeowPrivacySettingToString(settings.Status),
		ReadReceipts: convertWhatsmeowPrivacySettingToString(settings.ReadReceipts),
		Groups:       convertWhatsmeowPrivacySettingToString(settings.GroupAdd),
		Calls:        convertWhatsmeowPrivacySettingToString(settings.Call),
	}

	m.logger.Debugf("Retrieved privacy settings for session %s", sessionID)
	return result, nil
}

func (m *MeowService) UpdateBlocklist(ctx context.Context, sessionID, action string, phones []string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	if action != "add" && action != "remove" {
		return fmt.Errorf("invalid action %s, must be 'add' or 'remove'", action)
	}

	var jids []waTypes.JID
	for _, phone := range phones {
		jid, err := parsePhoneToJID(phone)
		if err != nil {
			m.logger.Warnf("Invalid phone number %s: %v", phone, err)
			continue
		}
		jids = append(jids, jid)
	}

	if len(jids) == 0 {
		return fmt.Errorf("no valid phone numbers provided")
	}

	for _, jid := range jids {
		err := client.GetClient().UpdateBlocklist(jid, action)
		if err != nil {
			m.logger.Warnf("Failed to %s %s to blocklist for session %s: %v", action, jid.String(), sessionID, err)
		} else {
			m.logger.Debugf("Successfully %sed %s to blocklist for session %s", action, jid.String(), sessionID)
		}
	}

	m.logger.Infof("Updated blocklist with %d contacts for session %s", len(jids), sessionID)
	return nil
}

func (m *MeowService) FindPrivacySettings(ctx context.Context, sessionID, category string) (*ports.PrivacySettingInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	settings, err := client.GetClient().TryFetchPrivacySettings(false)
	if err != nil {
		return nil, fmt.Errorf("failed to get privacy settings: %w", err)
	}

	result := &ports.PrivacySettingInfo{
		Category: category,
	}

	switch category {
	case "last_seen":
		result.Value = convertWhatsmeowPrivacySettingToString(settings.LastSeen)
	case "online":
		result.Value = convertWhatsmeowPrivacySettingToString(settings.Online)
	case "profile_photo":
		result.Value = convertWhatsmeowPrivacySettingToString(settings.ProfilePhoto)
	case "status":
		result.Value = convertWhatsmeowPrivacySettingToString(settings.Status)
	case "read_receipts":
		result.Value = convertWhatsmeowPrivacySettingToString(settings.ReadReceipts)
	case "groups":
		result.Value = convertWhatsmeowPrivacySettingToString(settings.GroupAdd)
	case "calls":
		result.Value = convertWhatsmeowPrivacySettingToString(settings.Call)
	default:
		return nil, fmt.Errorf("unknown privacy category: %s", category)
	}

	m.logger.Debugf("Retrieved privacy setting %s = %s for session %s", category, result.Value, sessionID)
	return result, nil
}

// Helper functions for privacy settings conversion

func convertPrivacySettingToWhatsmeow(setting string) whatsmeow.PrivacySetting {
	switch setting {
	case "everyone":
		return whatsmeow.PrivacySettingEveryone
	case "contacts":
		return whatsmeow.PrivacySettingContacts
	case "contact_blacklist":
		return whatsmeow.PrivacySettingContactBlacklist
	case "none":
		return whatsmeow.PrivacySettingNone
	default:
		return whatsmeow.PrivacySettingUndefined
	}
}

func convertWhatsmeowPrivacySettingToString(setting whatsmeow.PrivacySetting) string {
	switch setting {
	case whatsmeow.PrivacySettingEveryone:
		return "everyone"
	case whatsmeow.PrivacySettingContacts:
		return "contacts"
	case whatsmeow.PrivacySettingContactBlacklist:
		return "contact_blacklist"
	case whatsmeow.PrivacySettingNone:
		return "none"
	default:
		return "undefined"
	}
}
