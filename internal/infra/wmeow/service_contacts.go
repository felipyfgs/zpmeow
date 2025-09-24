package wmeow

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/ports"

	waTypes "go.mau.fi/whatsmeow/types"
)

// ContactManager methods - gestão de contatos e usuários

func (m *MeowService) CheckUser(ctx context.Context, sessionID string, phones []string) ([]ports.UserCheckResult, error) {
	client, err := m.validateAndGetConnectedClient(sessionID)
	if err != nil {
		return nil, err
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

	var results []ports.UserCheckResult
	for _, result := range resp {
		userResult := ports.UserCheckResult{
			Phone:       result.Query,
			IsOnWhatsApp: result.IsIn,
		}

		if result.JID != nil {
			userResult.JID = result.JID.String()
		}

		results = append(results, userResult)
	}

	m.logger.Debugf("Checked %d phone numbers for session %s", len(results), sessionID)
	return results, nil
}

func (m *MeowService) GetContacts(ctx context.Context, sessionID string, offset, limit int) ([]ports.ContactInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	contacts, err := client.GetClient().Store.Contacts.GetAllContacts()
	if err != nil {
		return nil, fmt.Errorf("failed to get contacts: %w", err)
	}

	var allResults []ports.ContactInfo
	for jid, contact := range contacts {
		if jid.Server != waTypes.DefaultUserServer {
			continue
		}

		phone := jid.User
		if !strings.HasPrefix(phone, "+") {
			phone = "+" + phone
		}

		contactInfo := ports.ContactInfo{
			JID:   jid.String(),
			Phone: phone,
			Name:  contact.PushName,
		}

		if contact.BusinessName != "" {
			contactInfo.BusinessName = contact.BusinessName
		}

		allResults = append(allResults, contactInfo)
	}

	// Apply pagination
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
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

func (m *MeowService) GetContactInfo(ctx context.Context, sessionID, phone string) (*ports.ContactInfo, error) {
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

	contact, err := client.GetClient().Store.Contacts.GetContact(jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact info: %w", err)
	}

	result := &ports.ContactInfo{
		JID:   jid.String(),
		Phone: phone,
		Name:  contact.PushName,
	}

	if contact.BusinessName != "" {
		result.BusinessName = contact.BusinessName
	}

	m.logger.Debugf("Retrieved contact info for %s in session %s", phone, sessionID)
	return result, nil
}

func (m *MeowService) GetUserInfo(ctx context.Context, sessionID, phone string) (*UserInfo, error) {
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

	userInfo, err := client.GetClient().GetUserInfo([]waTypes.JID{jid})
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	if len(userInfo) == 0 {
		return nil, fmt.Errorf("user info not found for %s", phone)
	}

	info := userInfo[0]
	result := &UserInfo{
		JID:    jid.String(),
		Phone:  phone,
		Status: info.Status,
	}

	if info.PictureID != "" {
		result.PictureID = info.PictureID
	}

	m.logger.Debugf("Retrieved user info for %s in session %s", phone, sessionID)
	return result, nil
}

func (m *MeowService) GetProfilePicture(ctx context.Context, sessionID, phone string, preview bool) ([]byte, error) {
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

	pic, err := client.GetClient().GetProfilePictureInfo(jid, &preview)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile picture info: %w", err)
	}

	if pic == nil {
		return nil, fmt.Errorf("profile picture not found for %s", phone)
	}

	// Download the actual picture
	resp, err := client.GetClient().DangerousInternals().HTTP.Get(pic.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to download profile picture: %w", err)
	}
	defer resp.Body.Close()

	data := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile picture data: %w", err)
	}

	m.logger.Debugf("Retrieved profile picture for %s in session %s", phone, sessionID)
	return data, nil
}

func (m *MeowService) BlockUser(ctx context.Context, sessionID, phone string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := parsePhoneToJID(phone)
	if err != nil {
		return fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	err = client.GetClient().UpdateBlocklist(jid, "add")
	if err != nil {
		return fmt.Errorf("failed to block user: %w", err)
	}

	m.logger.Debugf("Blocked user %s in session %s", phone, sessionID)
	return nil
}

func (m *MeowService) UnblockUser(ctx context.Context, sessionID, phone string) error {
	client := m.getClient(sessionID)
	if client == nil {
		return fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := parsePhoneToJID(phone)
	if err != nil {
		return fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	err = client.GetClient().UpdateBlocklist(jid, "remove")
	if err != nil {
		return fmt.Errorf("failed to unblock user: %w", err)
	}

	m.logger.Debugf("Unblocked user %s in session %s", phone, sessionID)
	return nil
}

// Additional methods required by ContactManager interface

func (m *MeowService) CheckContact(ctx context.Context, sessionID, phone string) (*ports.UserCheckResult, error) {
	// Use CheckUser instead of GetContactInfo to match interface
	results, err := m.CheckUser(ctx, sessionID, []string{phone})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &ports.UserCheckResult{
			Phone:       phone,
			IsOnWhatsApp: false,
		}, nil
	}

	return &results[0], nil
}

// Helper methods for contact management - validateAndGetConnectedClient moved to service.go
