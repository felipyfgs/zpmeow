package chatwoot

import "zpmeow/internal/application/ports"

// ContactAdapter converte entre tipos internos e tipos das interfaces
type ContactAdapter struct{}

// NewContactAdapter cria um novo adaptador de contatos
func NewContactAdapter() *ContactAdapter {
	return &ContactAdapter{}
}

// ToPortsContact converte Contact interno para ports.ContactResponse
func (ca *ContactAdapter) ToPortsContact(contact *Contact) *ports.ContactResponse {
	if contact == nil {
		return nil
	}

	return &ports.ContactResponse{
		ID:          contact.ID,
		Name:        contact.Name,
		PhoneNumber: contact.PhoneNumber,
		Email:       contact.Email,
		Identifier:  contact.Identifier,
		CreatedAt:   convertTimeToString(contact.CreatedAt),
		UpdatedAt:   convertTimeToString(contact.UpdatedAt),
	}
}

// FromPortsContact converte ports.ContactResponse para Contact interno
func (ca *ContactAdapter) FromPortsContact(contact *ports.ContactResponse) *Contact {
	if contact == nil {
		return nil
	}

	return &Contact{
		ID:          contact.ID,
		Name:        contact.Name,
		PhoneNumber: contact.PhoneNumber,
		Email:       contact.Email,
		Identifier:  contact.Identifier,
		CreatedAt:   convertStringToTime(contact.CreatedAt),
		UpdatedAt:   convertStringToTime(contact.UpdatedAt),
	}
}

// ToPortsContactList converte slice de Contact para slice de ports.ContactResponse
func (ca *ContactAdapter) ToPortsContactList(contacts []Contact) []*ports.ContactResponse {
	result := make([]*ports.ContactResponse, len(contacts))
	for i, contact := range contacts {
		result[i] = ca.ToPortsContact(&contact)
	}
	return result
}

// ConversationAdapter converte entre tipos internos e tipos das interfaces
type ConversationAdapter struct{}

// NewConversationAdapter cria um novo adaptador de conversas
func NewConversationAdapter() *ConversationAdapter {
	return &ConversationAdapter{}
}

// ToPortsConversation converte Conversation interno para ports.ConversationResponse
func (ca *ConversationAdapter) ToPortsConversation(conversation *Conversation) *ports.ConversationResponse {
	if conversation == nil {
		return nil
	}

	return &ports.ConversationResponse{
		ID:        conversation.ID,
		InboxID:   conversation.InboxID,
		Status:    conversation.Status,
		ContactID: conversation.ContactID,
		CreatedAt: convertTimeToString(conversation.CreatedAt),
		UpdatedAt: convertTimeToString(conversation.UpdatedAt),
	}
}

// FromPortsConversation converte ports.ConversationResponse para Conversation interno
func (ca *ConversationAdapter) FromPortsConversation(conversation *ports.ConversationResponse) *Conversation {
	if conversation == nil {
		return nil
	}

	return &Conversation{
		ID:        conversation.ID,
		InboxID:   conversation.InboxID,
		Status:    conversation.Status,
		ContactID: conversation.ContactID,
		CreatedAt: convertStringToTime(conversation.CreatedAt),
		UpdatedAt: convertStringToTime(conversation.UpdatedAt),
	}
}

// ToPortsConversationList converte slice de Conversation para slice de ports.ConversationResponse
func (ca *ConversationAdapter) ToPortsConversationList(conversations []Conversation) []*ports.ConversationResponse {
	result := make([]*ports.ConversationResponse, len(conversations))
	for i, conversation := range conversations {
		result[i] = ca.ToPortsConversation(&conversation)
	}
	return result
}

// MessageAdapter converte entre tipos internos e tipos das interfaces
type MessageAdapter struct{}

// NewMessageAdapter cria um novo adaptador de mensagens
func NewMessageAdapter() *MessageAdapter {
	return &MessageAdapter{}
}

// ToPortsMessage converte Message interno para ports.MessageResponse
func (ma *MessageAdapter) ToPortsMessage(message *Message) *ports.MessageResponse {
	if message == nil {
		return nil
	}

	return &ports.MessageResponse{
		ID:             message.ID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		ConversationID: message.ConversationID,
		CreatedAt:      convertTimeToString(message.CreatedAt),
		UpdatedAt:      convertTimeToString(message.UpdatedAt),
	}
}

// FromPortsMessage converte ports.MessageResponse para Message interno
func (ma *MessageAdapter) FromPortsMessage(message *ports.MessageResponse) *Message {
	if message == nil {
		return nil
	}

	return &Message{
		ID:             message.ID,
		Content:        message.Content,
		MessageType:    message.MessageType,
		ConversationID: message.ConversationID,
		CreatedAt:      convertStringToTime(message.CreatedAt),
		UpdatedAt:      convertStringToTime(message.UpdatedAt),
	}
}

// InboxAdapter converte entre tipos internos e tipos das interfaces
type InboxAdapter struct{}

// NewInboxAdapter cria um novo adaptador de inboxes
func NewInboxAdapter() *InboxAdapter {
	return &InboxAdapter{}
}

// ToPortsInbox converte Inbox interno para ports.InboxResponse
func (ia *InboxAdapter) ToPortsInbox(inbox *Inbox) *ports.InboxResponse {
	if inbox == nil {
		return nil
	}

	return &ports.InboxResponse{
		ID:          inbox.ID,
		Name:        inbox.Name,
		ChannelType: inbox.ChannelType,
		WebhookURL:  inbox.WebhookURL,
		CreatedAt:   "",
		UpdatedAt:   "",
	}
}

// FromPortsInbox converte ports.InboxResponse para Inbox interno
func (ia *InboxAdapter) FromPortsInbox(inbox *ports.InboxResponse) *Inbox {
	if inbox == nil {
		return nil
	}

	return &Inbox{
		ID:          inbox.ID,
		Name:        inbox.Name,
		ChannelType: inbox.ChannelType,
		WebhookURL:  inbox.WebhookURL,
	}
}

// ToPortsInboxList converte slice de Inbox para slice de ports.InboxResponse
func (ia *InboxAdapter) ToPortsInboxList(inboxes []Inbox) []*ports.InboxResponse {
	result := make([]*ports.InboxResponse, len(inboxes))
	for i, inbox := range inboxes {
		result[i] = ia.ToPortsInbox(&inbox)
	}
	return result
}

// RequestAdapter converte entre tipos de request internos e das interfaces
type RequestAdapter struct{}

// NewRequestAdapter cria um novo adaptador de requests
func NewRequestAdapter() *RequestAdapter {
	return &RequestAdapter{}
}

// FromPortsContactCreateRequest converte ports.ContactCreateRequest para ContactCreateRequest interno
func (ra *RequestAdapter) FromPortsContactCreateRequest(req *ports.ContactCreateRequest) *ContactCreateRequest {
	if req == nil {
		return nil
	}

	return &ContactCreateRequest{
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		Email:       req.Email,
		Identifier:  req.Identifier,
		InboxID:     req.InboxID,
		AvatarURL:   req.AvatarURL,
	}
}

// FromPortsConversationCreateRequest converte ports.ConversationCreateRequest para ConversationCreateRequest interno
func (ra *RequestAdapter) FromPortsConversationCreateRequest(req *ports.ConversationCreateRequest) *ConversationCreateRequest {
	if req == nil {
		return nil
	}

	return &ConversationCreateRequest{
		ContactID: req.ContactID,
		InboxID:   req.InboxID,
		Status:    req.Status,
	}
}

// FromPortsMessageCreateRequest converte ports.MessageCreateRequest para MessageCreateRequest interno
func (ra *RequestAdapter) FromPortsMessageCreateRequest(req *ports.MessageCreateRequest) *MessageCreateRequest {
	if req == nil {
		return nil
	}

	return &MessageCreateRequest{
		Content:     req.Content,
		MessageType: req.MessageType,
		SourceID:    req.SourceID,
	}
}

// FromPortsInboxCreateRequest converte ports.InboxCreateRequest para InboxCreateRequest interno
func (ra *RequestAdapter) FromPortsInboxCreateRequest(req *ports.InboxCreateRequest) *InboxCreateRequest {
	if req == nil {
		return nil
	}

	return &InboxCreateRequest{
		Name:    req.Name,
		Channel: req.Channel,
	}
}

// WhatsAppAdapter converte entre tipos de WhatsApp internos e das interfaces
type WhatsAppAdapter struct{}

// NewWhatsAppAdapter cria um novo adaptador de WhatsApp
func NewWhatsAppAdapter() *WhatsAppAdapter {
	return &WhatsAppAdapter{}
}

// FromPortsWhatsAppMessage converte ports.WhatsAppMessage para WhatsAppMessage interno
func (wa *WhatsAppAdapter) FromPortsWhatsAppMessage(msg *ports.WhatsAppMessage) *WhatsAppMessage {
	if msg == nil {
		return nil
	}

	return &WhatsAppMessage{
		ID:        msg.ID,
		From:      msg.From,
		To:        msg.To,
		Body:      msg.Body,
		Type:      msg.Type,
		Timestamp: msg.Timestamp,
		FromMe:    msg.FromMe,
		PushName:  msg.PushName,
		ChatName:  msg.ChatName,
		Caption:   msg.Caption,
		FileName:  msg.FileName,
		MediaURL:  msg.MediaURL,
		MimeType:  msg.MimeType,
	}
}

// ToPortsWhatsAppMessage converte WhatsAppMessage interno para ports.WhatsAppMessage
func (wa *WhatsAppAdapter) ToPortsWhatsAppMessage(msg *WhatsAppMessage) *ports.WhatsAppMessage {
	if msg == nil {
		return nil
	}

	return &ports.WhatsAppMessage{
		ID:        msg.ID,
		From:      msg.From,
		To:        msg.To,
		Body:      msg.Body,
		Type:      msg.Type,
		Timestamp: msg.Timestamp,
		FromMe:    msg.FromMe,
		PushName:  msg.PushName,
		ChatName:  msg.ChatName,
		Caption:   msg.Caption,
		FileName:  msg.FileName,
		MediaURL:  msg.MediaURL,
		MimeType:  msg.MimeType,
		Data:      make(map[string]interface{}),
	}
}

// Utility functions for time conversion
func convertTimeToString(t interface{}) string {
	if t == nil {
		return ""
	}
	// Implementação simplificada - pode ser melhorada conforme necessário
	return ""
}

func convertStringToTime(s string) interface{} {
	if s == "" {
		return nil
	}
	// Implementação simplificada - pode ser melhorada conforme necessário
	return s
}
