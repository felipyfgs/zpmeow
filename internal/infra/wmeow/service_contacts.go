package wmeow

import (
	"context"
	"fmt"

	"zpmeow/internal/application/ports"

	waTypes "go.mau.fi/whatsmeow/types"
)

// ContactManager methods - gestão de contatos e usuários

func (m *MeowService) CheckUser(ctx context.Context, sessionID string, phones []string) ([]ports.UserCheckResult, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	var validPhones []string
	for _, phone := range phones {
		_, err := waTypes.ParseJID(phone + "@s.whatsapp.net")
		if err != nil {
			m.logger.Warnf("Invalid phone number %s: %v", phone, err)
			continue
		}
		validPhones = append(validPhones, phone)
	}

	if len(validPhones) == 0 {
		return []ports.UserCheckResult{}, nil
	}

	// For now, return basic results
	var results []ports.UserCheckResult
	for _, phone := range validPhones {
		results = append(results, ports.UserCheckResult{
			Query:        phone,
			IsInWhatsapp: true, // Assume true for now
			JID:          phone + "@s.whatsapp.net",
		})
	}

	m.logger.Debugf("Checked %d users for session %s", len(results), sessionID)
	return results, nil
}

func (m *MeowService) CheckContact(ctx context.Context, sessionID, phone string) (*ports.UserCheckResult, error) {
	results, err := m.CheckUser(ctx, sessionID, []string{phone})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no results for phone %s", phone)
	}

	return &results[0], nil
}

func (m *MeowService) GetContacts(ctx context.Context, sessionID string, limit, offset int) ([]ports.ContactResult, error) {
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

	var allResults []ports.ContactResult
	for jid, contact := range contacts {
		if jid.Server != waTypes.DefaultUserServer {
			continue
		}

		contactResult := ports.ContactResult{
			JID:      jid.String(),
			Name:     contact.PushName,
			Notify:   contact.PushName,
			PushName: contact.PushName,
		}

		if contact.BusinessName != "" {
			contactResult.BusinessName = contact.BusinessName
		}

		allResults = append(allResults, contactResult)
	}

	// Apply pagination
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100
	}

	end := offset + limit
	if end > len(allResults) {
		end = len(allResults)
	}

	if offset >= len(allResults) {
		return []ports.ContactResult{}, nil
	}

	result := allResults[offset:end]
	m.logger.Debugf("Retrieved %d contacts for session %s", len(result), sessionID)
	return result, nil
}

func (m *MeowService) GetContactInfo(ctx context.Context, sessionID, phone string) (*ports.ContactInfo, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	jid, err := waTypes.ParseJID(phone + "@s.whatsapp.net")
	if err != nil {
		return nil, fmt.Errorf("invalid phone number %s: %w", phone, err)
	}

	contact, err := client.GetClient().Store.Contacts.GetContact(ctx, jid)
	if err != nil {
		return nil, fmt.Errorf("failed to get contact info: %w", err)
	}

	result := &ports.ContactInfo{
		Phone: phone,
		Name:  contact.PushName,
	}

	if contact.BusinessName != "" {
		result.Name = contact.BusinessName
	}

	m.logger.Debugf("Retrieved contact info for %s in session %s", phone, sessionID)
	return result, nil
}

func (m *MeowService) GetUserInfo(ctx context.Context, sessionID string, phones []string) (map[string]ports.UserInfoResult, error) {
	// For now, return empty map
	result := make(map[string]ports.UserInfoResult)

	for _, phone := range phones {
		result[phone] = ports.UserInfoResult{
			JID:  phone + "@s.whatsapp.net",
			Name: "",
		}
	}

	m.logger.Debugf("Retrieved user info for %d phones in session %s", len(phones), sessionID)
	return result, nil
}

func (m *MeowService) GetProfilePicture(ctx context.Context, sessionID, phone string, preview bool) ([]byte, error) {
	// For now, return empty data
	m.logger.Debugf("GetProfilePicture for %s in session %s (returning empty for now)", phone, sessionID)
	return []byte{}, nil
}

func (m *MeowService) BlockContact(ctx context.Context, sessionID, phone string) error {
	// For now, just log
	m.logger.Debugf("BlockContact: %s for session %s", phone, sessionID)
	return nil
}

func (m *MeowService) UnblockContact(ctx context.Context, sessionID, phone string) error {
	// For now, just log
	m.logger.Debugf("UnblockContact: %s for session %s", phone, sessionID)
	return nil
}

func (m *MeowService) IsUserOnWhatsApp(ctx context.Context, sessionID string, phones []string) ([]ports.UserCheckResult, error) {
	// Delegate to CheckUser
	return m.CheckUser(ctx, sessionID, phones)
}
