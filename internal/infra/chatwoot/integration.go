package chatwoot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"zpmeow/internal/application/ports"
	"zpmeow/internal/infra/database/repository"
)

// Integration representa a integração completa com Chatwoot
type Integration struct {
	services        map[string]*Service
	configs         map[string]*ChatwootConfig
	logger          *slog.Logger
	mutex           sync.RWMutex
	whatsappService ports.WhatsAppService
	messageRepo     *repository.MessageRepository
	zpCwRepo        *repository.ZpCwMessageRepository
	chatRepo        *repository.ChatRepository
}

// NewIntegration cria uma nova instância da integração Chatwoot
func NewIntegration(logger *slog.Logger, messageRepo *repository.MessageRepository, zpCwRepo *repository.ZpCwMessageRepository, chatRepo *repository.ChatRepository) *Integration {
	return &Integration{
		services:    make(map[string]*Service),
		configs:     make(map[string]*ChatwootConfig),
		logger:      logger,
		messageRepo: messageRepo,
		zpCwRepo:    zpCwRepo,
		chatRepo:    chatRepo,
	}
}

// SetWhatsAppService define o serviço WhatsApp para a integração e atualiza todos os serviços existentes
func (i *Integration) SetWhatsAppService(whatsappService ports.WhatsAppService) {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.whatsappService = whatsappService

	// Atualiza todos os serviços existentes
	for _, service := range i.services {
		service.SetWhatsAppService(whatsappService)
	}
}

// RegisterSession registra uma nova sessão com configuração Chatwoot
func (i *Integration) RegisterSession(sessionId string, config *ChatwootConfig) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if !config.IsActive {
		i.logger.Info("Chatwoot disabled for session", "sessionId", sessionId)
		// Remove serviço se existir
		delete(i.services, sessionId)
		i.configs[sessionId] = config
		return nil
	}

	// Cria serviço para a sessão (whatsappService pode ser nil aqui, será definido depois)
	service, err := NewService(config, i.logger.With("sessionId", sessionId), i.whatsappService, sessionId, i.messageRepo, i.zpCwRepo, i.chatRepo)
	if err != nil {
		return fmt.Errorf("failed to create chatwoot service for session %s: %w", sessionId, err)
	}

	// Armazena configurações
	i.services[sessionId] = service
	i.configs[sessionId] = config

	i.logger.Info("Chatwoot integration registered", "sessionId", sessionId)
	return nil
}

// UnregisterSession remove uma sessão da integração
func (i *Integration) UnregisterSession(sessionId string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	delete(i.services, sessionId)
	delete(i.configs, sessionId)

	i.logger.Info("Chatwoot integration unregistered", "sessionId", sessionId)
}

// GetService retorna o serviço Chatwoot para uma sessão
func (i *Integration) GetService(sessionId string) (*Service, bool) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	service, exists := i.services[sessionId]
	return service, exists
}

// GetConfig retorna a configuração Chatwoot para uma sessão
func (i *Integration) GetConfig(sessionId string) (*ChatwootConfig, bool) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	config, exists := i.configs[sessionId]
	return config, exists
}

// ProcessMessage processa uma mensagem do WhatsApp para uma sessão específica
func (i *Integration) ProcessMessage(ctx context.Context, sessionId string, msg *WhatsAppMessage) error {
	service, exists := i.GetService(sessionId)
	if !exists {
		// Não há integração Chatwoot configurada para esta sessão
		return nil
	}

	return service.ProcessWhatsAppMessage(ctx, msg)
}

// ProcessWebhook processa um webhook do Chatwoot
func (i *Integration) ProcessWebhook(ctx context.Context, sessionId string, payload *WebhookPayload) error {
	service, exists := i.GetService(sessionId)
	if !exists {
		return fmt.Errorf("no chatwoot service found for session: %s", sessionId)
	}

	return service.ProcessWebhook(ctx, payload)
}

// IsEnabled verifica se a integração Chatwoot está habilitada para uma sessão
func (i *Integration) IsEnabled(sessionId string) bool {
	config, exists := i.GetConfig(sessionId)
	return exists && config.IsActive
}

// GetEnabledSessions retorna lista de sessões com Chatwoot habilitado
func (i *Integration) GetEnabledSessions() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	var enabled []string
	for sessionId, config := range i.configs {
		if config.IsActive {
			enabled = append(enabled, sessionId)
		}
	}

	return enabled
}

// GetSessionsCount retorna o número total de sessões registradas
func (i *Integration) GetSessionsCount() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return len(i.configs)
}

// GetEnabledSessionsCount retorna o número de sessões habilitadas
func (i *Integration) GetEnabledSessionsCount() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	count := 0
	for _, config := range i.configs {
		if config.IsActive {
			count++
		}
	}

	return count
}

// ListSessions retorna informações sobre todas as sessões
func (i *Integration) ListSessions() []SessionInfo {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	sessions := make([]SessionInfo, 0, len(i.configs))
	for sessionId, config := range i.configs {
		_, hasService := i.services[sessionId]

		sessions = append(sessions, SessionInfo{
			SessionId: sessionId,
			Enabled:   config.IsActive,
			URL:       config.URL,
			InboxName: config.NameInbox,
			Connected: hasService && config.IsActive,
		})
	}

	return sessions
}

// SessionInfo representa informações sobre uma sessão
type SessionInfo struct {
	SessionId string `json:"sessionId"`
	Enabled   bool   `json:"enabled"`
	URL       string `json:"url"`
	InboxName string `json:"inboxName"`
	Connected bool   `json:"connected"`
}

// Health verifica a saúde da integração
func (i *Integration) Health() HealthStatus {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	total := len(i.configs)
	enabled := 0
	connected := 0

	for sessionId, config := range i.configs {
		if config.IsActive {
			enabled++
			if _, hasService := i.services[sessionId]; hasService {
				connected++
			}
		}
	}

	status := "healthy"
	if enabled > 0 && connected == 0 {
		status = "unhealthy"
	} else if connected < enabled {
		status = "degraded"
	}

	return HealthStatus{
		Status:            status,
		TotalSessions:     total,
		EnabledSessions:   enabled,
		ConnectedSessions: connected,
	}
}

// HealthStatus representa o status de saúde da integração
type HealthStatus struct {
	Status            string `json:"status"`
	TotalSessions     int    `json:"totalSessions"`
	EnabledSessions   int    `json:"enabledSessions"`
	ConnectedSessions int    `json:"connectedSessions"`
}

// Shutdown desliga graciosamente a integração
func (i *Integration) Shutdown(ctx context.Context) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.logger.Info("Shutting down Chatwoot integration")

	// Limpa todos os serviços e configurações
	i.services = make(map[string]*Service)
	i.configs = make(map[string]*ChatwootConfig)

	i.logger.Info("Chatwoot integration shutdown completed")
	return nil
}

// ValidateConfig valida uma configuração Chatwoot
func (i *Integration) ValidateConfig(config *ChatwootConfig) error {
	if !config.IsActive {
		return nil // Configuração desabilitada é válida
	}

	if config.AccountID == "" {
		return fmt.Errorf("account ID is required when Chatwoot is active")
	}

	if config.Token == "" {
		return fmt.Errorf("token is required when Chatwoot is enabled")
	}

	if config.URL == "" {
		return fmt.Errorf("URL is required when Chatwoot is enabled")
	}

	// Valida formato da URL
	if !isValidURL(config.URL) {
		return fmt.Errorf("invalid URL format")
	}

	return nil
}

// TestConnection testa a conexão com uma configuração Chatwoot
func (i *Integration) TestConnection(ctx context.Context, config *ChatwootConfig) error {
	if err := i.ValidateConfig(config); err != nil {
		return err
	}

	if !config.IsActive {
		return fmt.Errorf("chatwoot is disabled")
	}

	// Cria cliente temporário para teste
	client := NewClient(config.URL, config.Token, config.AccountID, nil)

	// Testa conexão listando inboxes
	_, err := client.ListInboxes(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to Chatwoot: %w", err)
	}

	return nil
}

// GetMetrics retorna métricas da integração
func (i *Integration) GetMetrics() Metrics {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	metrics := Metrics{
		TotalSessions:     len(i.configs),
		EnabledSessions:   0,
		ConnectedSessions: 0,
		SessionMetrics:    make(map[string]SessionMetrics),
	}

	for sessionId, config := range i.configs {
		sessionMetric := SessionMetrics{
			Enabled:   config.IsActive,
			Connected: false,
		}

		if config.IsActive {
			metrics.EnabledSessions++

			if _, hasService := i.services[sessionId]; hasService {
				metrics.ConnectedSessions++
				sessionMetric.Connected = true
			}
		}

		metrics.SessionMetrics[sessionId] = sessionMetric
	}

	return metrics
}

// Metrics representa métricas da integração
type Metrics struct {
	TotalSessions     int                       `json:"totalSessions"`
	EnabledSessions   int                       `json:"enabledSessions"`
	ConnectedSessions int                       `json:"connectedSessions"`
	SessionMetrics    map[string]SessionMetrics `json:"sessionMetrics"`
}

// SessionMetrics representa métricas de uma sessão específica
type SessionMetrics struct {
	Enabled   bool `json:"enabled"`
	Connected bool `json:"connected"`
	// Aqui você pode adicionar mais métricas específicas como:
	// MessagesProcessed int `json:"messagesProcessed"`
	// LastActivity      time.Time `json:"lastActivity"`
	// ErrorCount        int `json:"errorCount"`
}

// Métodos de compatibilidade para manter API existente funcionando
// DEPRECATED: Use RegisterSession instead
func (i *Integration) RegisterInstance(instanceName string, config *ChatwootConfig) error {
	return i.RegisterSession(instanceName, config)
}

// DEPRECATED: Use UnregisterSession instead
func (i *Integration) UnregisterInstance(instanceName string) {
	i.UnregisterSession(instanceName)
}

// DEPRECATED: Use GetEnabledSessions instead
func (i *Integration) GetEnabledInstances() []string {
	return i.GetEnabledSessions()
}

// DEPRECATED: Use GetSessionsCount instead
func (i *Integration) GetInstancesCount() int {
	return i.GetSessionsCount()
}

// DEPRECATED: Use GetEnabledSessionsCount instead
func (i *Integration) GetEnabledInstancesCount() int {
	return i.GetEnabledSessionsCount()
}

// DEPRECATED: Use ListSessions instead
func (i *Integration) ListInstances() []SessionInfo {
	return i.ListSessions()
}

// DEPRECATED: Use SessionInfo instead
type InstanceInfo = SessionInfo

// DEPRECATED: Use SessionMetrics instead
type InstanceMetrics = SessionMetrics
