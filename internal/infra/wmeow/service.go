package wmeow

import (
	"context"
	"fmt"
	"sync"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/chatwoot"
	"zpmeow/internal/infra/database/repository"
	"zpmeow/internal/infra/logging"

	"github.com/jmoiron/sqlx"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
)

// Use ports types directly
type WameowService = ports.WameowService

type MeowService struct {
	clients             map[string]*WameowClient
	sessions            session.Repository
	logger              logging.Logger
	container           *sqlstore.Container
	waLogger            waLog.Logger
	mu                  sync.RWMutex
	messageSender       *messageSender
	mimeHelper          *mimeTypeHelper
	chatwootIntegration *chatwoot.Integration
	chatwootRepo        *repository.ChatwootRepository
	messageRepo         *repository.MessageRepository
	chatRepo            *repository.ChatRepository
	webhookRepo         *repository.WebhookRepository
}

// Construtores
func NewMeowService(container *sqlstore.Container, waLogger waLog.Logger, sessionRepo session.Repository, db *sqlx.DB) WameowService {
	// Criar repositórios de mensagem, chat e webhook
	messageRepo := repository.NewMessageRepository(db)
	chatRepo := repository.NewChatRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	return &MeowService{
		clients:       make(map[string]*WameowClient),
		sessions:      sessionRepo,
		logger:        logging.GetLogger().Sub("wameow"),
		container:     container,
		waLogger:      waLogger,
		messageSender: NewMessageSender(),
		mimeHelper:    NewMimeTypeHelper(),
		messageRepo:   messageRepo,
		chatRepo:      chatRepo,
		webhookRepo:   webhookRepo,
	}
}

func NewMeowServiceWithChatwoot(container *sqlstore.Container, waLogger waLog.Logger, sessionRepo session.Repository, chatwootIntegration *chatwoot.Integration, chatwootRepo *repository.ChatwootRepository, db *sqlx.DB) WameowService {
	// Criar repositórios de mensagem, chat e webhook
	messageRepo := repository.NewMessageRepository(db)
	chatRepo := repository.NewChatRepository(db)
	webhookRepo := repository.NewWebhookRepository(db)

	return &MeowService{
		clients:             make(map[string]*WameowClient),
		sessions:            sessionRepo,
		logger:              logging.GetLogger().Sub("wameow"),
		container:           container,
		waLogger:            waLogger,
		messageSender:       NewMessageSender(),
		mimeHelper:          NewMimeTypeHelper(),
		chatwootIntegration: chatwootIntegration,
		chatwootRepo:        chatwootRepo,
		messageRepo:         messageRepo,
		chatRepo:            chatRepo,
		webhookRepo:         webhookRepo,
	}
}

// Métodos de coordenação e helpers internos (não duplicados)

func (m *MeowService) SetChatwootIntegration(integration *chatwoot.Integration) {
	m.chatwootIntegration = integration
}

func (m *MeowService) getClient(sessionID string) *WameowClient {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[sessionID]
}

func (m *MeowService) getOrCreateClient(sessionID string) *WameowClient {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, exists := m.clients[sessionID]; exists {
		return client
	}

	return m.createNewClient(sessionID)
}

func (m *MeowService) createNewClient(sessionID string) *WameowClient {
	sessionConfig := m.loadSessionConfiguration(sessionID)
	if sessionConfig == nil {
		m.logger.Errorf("Failed to load session configuration for %s", sessionID)
		return nil
	}

	// Create event processor
	var eventProcessor *EventProcessor
	if m.chatwootIntegration != nil {
		eventProcessor = NewEventProcessorWithChatwoot(
			sessionID,
			m.sessions,
			m.chatwootIntegration,
			m.chatwootRepo,
			m.messageRepo,
			m.chatRepo,
			m.webhookRepo,
		)
	} else {
		eventProcessor = NewEventProcessor(
			sessionID,
			m.sessions,
			m.messageRepo,
			m.chatRepo,
			m.webhookRepo,
		)
	}

	client, err := NewWameowClient(
		sessionID,
		m.container,
		m.waLogger,
		eventProcessor,
		m.sessions,
	)
	if err != nil {
		m.logger.Errorf("Failed to create WameowClient for session %s: %v", sessionID, err)
		return nil
	}

	m.clients[sessionID] = client
	return client
}

func (m *MeowService) loadSessionConfiguration(sessionID string) *SessionConfiguration {
	sessionEntity, err := m.sessions.GetByID(context.Background(), sessionID)
	if err != nil {
		m.logger.Errorf("Failed to load session %s: %v", sessionID, err)
		return nil
	}

	if sessionEntity == nil {
		m.logger.Errorf("Session %s not found", sessionID)
		return nil
	}

	return &SessionConfiguration{
		SessionID:   sessionID,
		PhoneNumber: "", // TODO: Get phone number from session entity
		Status:      string(sessionEntity.Status()),
		QRCode:      sessionEntity.QRCode().Value(),
		Connected:   sessionEntity.IsConnected(),
		Webhook:     "", // TODO: Get webhook from session entity
	}
}

func (m *MeowService) removeClient(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, sessionID)
}

// Métodos de coordenação para inicialização
func (m *MeowService) ConnectOnStartup(ctx context.Context) error {
	m.logger.Info("Starting connection process for all sessions on startup")

	sessions, err := m.sessions.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	for _, sessionEntity := range sessions {
		sessionID := sessionEntity.ID().Value()
		m.logger.Infof("Attempting to connect session %s on startup", sessionID)

		client := m.getOrCreateClient(sessionID)
		if client == nil {
			m.logger.Errorf("Failed to create client for session %s", sessionID)
			continue
		}

		if err := client.Connect(); err != nil {
			m.logger.Errorf("Failed to connect session %s on startup: %v", sessionID, err)
			continue
		}

		m.logger.Infof("Successfully connected session %s on startup", sessionID)
	}

	m.logger.Info("Completed connection process for all sessions on startup")
	return nil
}

// Helpers para validação (usados pelos arquivos especializados)
func (m *MeowService) validateAndGetClient(sessionID string) (*WameowClient, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}
	return client, nil
}

func (m *MeowService) validateAndGetClientForSending(sessionID string) (*WameowClient, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	return client, nil
}

// Helpers para componentes (usados pelos arquivos especializados)
func (m *MeowService) getValidator() MessageValidator {
	return &messageValidatorWrapper{validator: NewMessageValidator()}
}

func (m *MeowService) getMessageBuilder() MessageBuilder {
	return &messageBuilderWrapper{builder: NewMessageBuilder()}
}

func (m *MeowService) getMessageSender() *messageSender {
	return m.messageSender
}

func (m *MeowService) getMimeHelper() *mimeTypeHelper {
	return m.mimeHelper
}

// Getters para repositórios (usados pelos arquivos especializados)
func (m *MeowService) GetMessageRepo() *repository.MessageRepository {
	return m.messageRepo
}

func (m *MeowService) GetChatRepo() *repository.ChatRepository {
	return m.chatRepo
}

func (m *MeowService) GetWebhookRepo() *repository.WebhookRepository {
	return m.webhookRepo
}

func (m *MeowService) GetChatwootRepo() *repository.ChatwootRepository {
	return m.chatwootRepo
}

func (m *MeowService) GetChatwootIntegration() *chatwoot.Integration {
	return m.chatwootIntegration
}

func (m *MeowService) GetLogger() logging.Logger {
	return m.logger
}

// Additional helper methods
func (m *MeowService) validateAndGetConnectedClient(sessionID string) (*WameowClient, error) {
	client := m.getClient(sessionID)
	if client == nil {
		return nil, fmt.Errorf("client not found for session %s", sessionID)
	}

	if !client.IsConnected() {
		return nil, fmt.Errorf("client not connected for session %s", sessionID)
	}

	return client, nil
}
