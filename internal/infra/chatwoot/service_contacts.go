package chatwoot

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"zpmeow/internal/application/ports"
)

// ContactService gerencia operações relacionadas a contatos
type ContactService struct {
	client       *Client
	logger       *slog.Logger
	cacheManager ports.ChatwootCacheManager
	errorHandler ports.ChatwootErrorHandler
	validator    ports.ChatwootValidator
	phoneUtils   *PhoneNumberUtils
	adapter      *ContactAdapter
}

// NewContactService cria um novo serviço de contatos
func NewContactService(client *Client, logger *slog.Logger, cacheManager ports.ChatwootCacheManager) *ContactService {
	return &ContactService{
		client:       client,
		logger:       logger,
		cacheManager: cacheManager,
		errorHandler: NewErrorHandler(),
		validator:    NewValidator(),
		phoneUtils:   NewPhoneNumberUtils(),
		adapter:      NewContactAdapter(),
	}
}

// FindOrCreateContact encontra ou cria um contato
func (cs *ContactService) FindOrCreateContact(ctx context.Context, phoneNumber, name, avatarURL string, isGroup bool, inboxID int) (*ports.ContactResponse, error) {
	// Valida dados de entrada
	if err := cs.validator.ValidateContactData(name, phoneNumber, isGroup); err != nil {
		return nil, cs.errorHandler.HandleContactError(err, phoneNumber)
	}

	// Verifica cache primeiro
	if contact, found := cs.cacheManager.GetContact(phoneNumber); found {
		cs.logger.Info("Contact found in cache", "phone", phoneNumber, "id", contact.ID)
		return contact, nil
	}

	// Busca contato existente
	internalContact, err := cs.searchExistingContact(ctx, phoneNumber, isGroup)
	if err != nil {
		cs.logger.Error("Failed to search contacts", "error", err, "phone", phoneNumber)
	}

	if internalContact != nil {
		// Converte para tipo da interface e salva no cache
		contact := cs.adapter.ToPortsContact(internalContact)
		cs.cacheManager.SetContact(phoneNumber, contact, DefaultCacheTTL)
		cs.logger.Info("Found existing contact", "phone", phoneNumber, "id", contact.ID, "name", contact.Name)
		return contact, nil
	}

	// Cria novo contato
	internalContact, err = cs.createNewContact(ctx, phoneNumber, name, avatarURL, isGroup, inboxID)
	if err != nil {
		return nil, cs.errorHandler.HandleContactError(err, phoneNumber)
	}

	// Converte para tipo da interface e salva no cache
	contact := cs.adapter.ToPortsContact(internalContact)
	cs.cacheManager.SetContact(phoneNumber, contact, DefaultCacheTTL)
	cs.logger.Info("Successfully created contact", "phone", phoneNumber, "id", contact.ID, "name", contact.Name)
	return contact, nil
}

// searchExistingContact busca um contato existente
func (cs *ContactService) searchExistingContact(ctx context.Context, phoneNumber string, isGroup bool) (*Contact, error) {
	var searchQuery string
	if isGroup {
		searchQuery = phoneNumber
	} else {
		searchQuery = fmt.Sprintf("+%s", phoneNumber)
	}

	cs.logger.Info("Searching for existing contact", "query", searchQuery, "is_group", isGroup)

	var contacts []Contact
	var err error

	if isGroup {
		contacts, err = cs.client.SearchContacts(ctx, searchQuery)
	} else {
		contacts, err = cs.client.FilterContacts(ctx, searchQuery)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search contacts: %w", err)
	}

	if len(contacts) > 0 {
		bestMatch := cs.findBestMatchContact(contacts, searchQuery)
		cs.logger.Info("Found existing contact", "phone", phoneNumber, "contact_id", bestMatch.ID)
		return &bestMatch, nil
	}

	cs.logger.Info("No existing contact found", "phone", phoneNumber)
	return nil, nil
}

// createNewContact cria um novo contato
func (cs *ContactService) createNewContact(ctx context.Context, phoneNumber, name, avatarURL string, isGroup bool, inboxID int) (*Contact, error) {
	req := ContactCreateRequest{
		InboxID: inboxID,
		Name:    name,
	}

	if !isGroup {
		req.PhoneNumber = fmt.Sprintf("+%s", phoneNumber)
		req.Identifier = fmt.Sprintf("%s@s.whatsapp.net", phoneNumber)
	} else {
		req.Identifier = phoneNumber
	}

	if avatarURL != "" {
		req.AvatarURL = avatarURL
	}

	cs.logger.Info("Creating new contact", "request", req)
	contact, err := cs.client.CreateContact(ctx, req)
	if err != nil {
		// Tenta buscar novamente se falhou por contato duplicado
		if cs.isContactDuplicateError(err) {
			cs.logger.Warn("Contact already exists, searching again", "phone", phoneNumber, "error", err)
			return cs.retrySearchAfterDuplicateError(ctx, phoneNumber, isGroup)
		}
		return nil, fmt.Errorf("failed to create contact: %w", err)
	}

	// Valida o contato criado
	if err := cs.validateContactResponse(contact); err != nil {
		return nil, fmt.Errorf("invalid contact response: %w", err)
	}

	cs.logger.Info("Successfully created contact", "name", name, "phone", phoneNumber, "id", contact.ID)
	return contact, nil
}

// findBestMatchContact encontra o melhor contato correspondente
func (cs *ContactService) findBestMatchContact(contacts []Contact, searchQuery string) Contact {
	if len(contacts) == 0 {
		return Contact{}
	}

	// Procura por correspondência exata primeiro
	for _, contact := range contacts {
		if contact.PhoneNumber == searchQuery || contact.Identifier == searchQuery {
			return contact
		}
	}

	// Se não encontrou correspondência exata, retorna o primeiro
	return contacts[0]
}

// isContactDuplicateError verifica se o erro é de contato duplicado
func (cs *ContactService) isContactDuplicateError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "already been taken") ||
		strings.Contains(errStr, "identifier has already been taken") ||
		strings.Contains(errStr, "duplicate")
}

// retrySearchAfterDuplicateError tenta buscar novamente após erro de duplicação
func (cs *ContactService) retrySearchAfterDuplicateError(ctx context.Context, phoneNumber string, isGroup bool) (*Contact, error) {
	// Tenta buscar com diferentes métodos
	searchQuery := phoneNumber
	if !isGroup {
		searchQuery = fmt.Sprintf("+%s", phoneNumber)
	}

	// Primeiro tenta com FilterContacts
	contacts, err := cs.client.FilterContacts(ctx, searchQuery)
	if err == nil && len(contacts) > 0 {
		contact := cs.findBestMatchContact(contacts, searchQuery)
		if contact.ID != 0 {
			cs.logger.Info("Found existing contact after creation failure",
				"name", contact.Name, "phone", contact.PhoneNumber, "id", contact.ID)
			return &contact, nil
		}
	}

	// Se falhou, tenta com SearchContacts
	contacts, err = cs.client.SearchContacts(ctx, searchQuery)
	if err == nil && len(contacts) > 0 {
		contact := cs.findBestMatchContact(contacts, searchQuery)
		if contact.ID != 0 {
			cs.logger.Info("Found existing contact with search after creation failure",
				"name", contact.Name, "phone", contact.PhoneNumber, "id", contact.ID)
			return &contact, nil
		}
	}

	return nil, fmt.Errorf("failed to find contact after duplicate error")
}

// validateContactResponse valida uma response de contato
func (cs *ContactService) validateContactResponse(contact *Contact) error {
	if contact == nil {
		return fmt.Errorf("contact response is nil")
	}

	if contact.ID == 0 {
		return fmt.Errorf("contact ID is invalid")
	}

	if contact.Name == "" {
		return fmt.Errorf("contact name is empty")
	}

	return nil
}

// GetContactByPhone busca um contato pelo número de telefone
func (cs *ContactService) GetContactByPhone(ctx context.Context, phoneNumber string) (*ports.ContactResponse, error) {
	// Verifica cache primeiro
	if contact, found := cs.cacheManager.GetContact(phoneNumber); found {
		return contact, nil
	}

	// Busca na API
	internalContact, err := cs.searchExistingContact(ctx, phoneNumber, false)
	if err != nil {
		return nil, cs.errorHandler.HandleContactError(err, phoneNumber)
	}

	if internalContact != nil {
		// Converte e salva no cache
		contact := cs.adapter.ToPortsContact(internalContact)
		cs.cacheManager.SetContact(phoneNumber, contact, DefaultCacheTTL)
		return contact, nil
	}

	return nil, nil
}

// ClearContactCache limpa o cache de um contato específico
func (cs *ContactService) ClearContactCache(phoneNumber string) {
	cs.cacheManager.DeleteContact(phoneNumber)
	cs.logger.Info("Cleared contact cache", "phone", phoneNumber)
}

// ExtractContactInfo extrai informações do contato de uma mensagem
func (cs *ContactService) ExtractContactInfo(msg *WhatsAppMessage) (phoneNumber string, isGroup bool, contactName string) {
	phoneNumber, _ = cs.phoneUtils.ExtractPhoneNumber(msg.From)
	isGroup = cs.phoneUtils.IsGroupJID(msg.From)

	if isGroup {
		if msg.ChatName != "" {
			contactName = msg.ChatName
		} else {
			contactName = fmt.Sprintf("Grupo %s", phoneNumber)
		}
	} else {
		if msg.PushName != "" {
			contactName = msg.PushName
		} else {
			contactName = phoneNumber
		}
	}

	return phoneNumber, isGroup, contactName
}
