package chatwoot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"zpmeow/internal/application/ports"
)

// Integration representa a integração completa com Chatwoot
type Integration struct {
	services        map[string]*Service
	configs         map[string]*ChatwootConfig
	logger          *slog.Logger
	mutex           sync.RWMutex
	whatsappService ports.WhatsAppService
}

// NewIntegration cria uma nova instância da integração Chatwoot
func NewIntegration(logger *slog.Logger) *Integration {
	return &Integration{
		services: make(map[string]*Service),
		configs:  make(map[string]*ChatwootConfig),
		logger:   logger,
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

// RegisterInstance registra uma nova instância com configuração Chatwoot
func (i *Integration) RegisterInstance(instanceName string, config *ChatwootConfig) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if !config.Enabled {
		i.logger.Info("Chatwoot disabled for instance", "instance", instanceName)
		// Remove serviço se existir
		delete(i.services, instanceName)
		i.configs[instanceName] = config
		return nil
	}

	// Cria serviço para a instância (whatsappService pode ser nil aqui, será definido depois)
	service, err := NewService(config, i.logger.With("instance", instanceName), i.whatsappService, instanceName)
	if err != nil {
		return fmt.Errorf("failed to create chatwoot service for instance %s: %w", instanceName, err)
	}

	// Armazena configurações
	i.services[instanceName] = service
	i.configs[instanceName] = config

	i.logger.Info("Chatwoot integration registered", "instance", instanceName)
	return nil
}

// UnregisterInstance remove uma instância da integração
func (i *Integration) UnregisterInstance(instanceName string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	delete(i.services, instanceName)
	delete(i.configs, instanceName)

	i.logger.Info("Chatwoot integration unregistered", "instance", instanceName)
}

// GetService retorna o serviço Chatwoot para uma instância
func (i *Integration) GetService(instanceName string) (*Service, bool) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	service, exists := i.services[instanceName]
	return service, exists
}

// GetConfig retorna a configuração Chatwoot para uma instância
func (i *Integration) GetConfig(instanceName string) (*ChatwootConfig, bool) {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	config, exists := i.configs[instanceName]
	return config, exists
}

// ProcessMessage processa uma mensagem do WhatsApp para uma instância específica
func (i *Integration) ProcessMessage(ctx context.Context, instanceName string, msg *WhatsAppMessage) error {
	service, exists := i.GetService(instanceName)
	if !exists {
		// Não há integração Chatwoot configurada para esta instância
		return nil
	}

	return service.ProcessWhatsAppMessage(ctx, msg)
}

// ProcessWebhook processa um webhook do Chatwoot
func (i *Integration) ProcessWebhook(ctx context.Context, instanceName string, payload *WebhookPayload) error {
	service, exists := i.GetService(instanceName)
	if !exists {
		return fmt.Errorf("no chatwoot service found for instance: %s", instanceName)
	}

	return service.ProcessWebhook(ctx, payload)
}

// IsEnabled verifica se a integração Chatwoot está habilitada para uma instância
func (i *Integration) IsEnabled(instanceName string) bool {
	config, exists := i.GetConfig(instanceName)
	return exists && config.Enabled
}

// GetEnabledInstances retorna lista de instâncias com Chatwoot habilitado
func (i *Integration) GetEnabledInstances() []string {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	var enabled []string
	for instanceName, config := range i.configs {
		if config.Enabled {
			enabled = append(enabled, instanceName)
		}
	}

	return enabled
}

// GetInstancesCount retorna o número total de instâncias registradas
func (i *Integration) GetInstancesCount() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	return len(i.configs)
}

// GetEnabledInstancesCount retorna o número de instâncias habilitadas
func (i *Integration) GetEnabledInstancesCount() int {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	count := 0
	for _, config := range i.configs {
		if config.Enabled {
			count++
		}
	}

	return count
}

// ListInstances retorna informações sobre todas as instâncias
func (i *Integration) ListInstances() []InstanceInfo {
	i.mutex.RLock()
	defer i.mutex.RUnlock()

	instances := make([]InstanceInfo, 0, len(i.configs))
	for instanceName, config := range i.configs {
		_, hasService := i.services[instanceName]
		
		instances = append(instances, InstanceInfo{
			Name:      instanceName,
			Enabled:   config.Enabled,
			URL:       config.URL,
			InboxName: config.NameInbox,
			Connected: hasService && config.Enabled,
		})
	}

	return instances
}

// InstanceInfo representa informações sobre uma instância
type InstanceInfo struct {
	Name      string `json:"name"`
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

	for _, config := range i.configs {
		if config.Enabled {
			enabled++
			if _, hasService := i.services[config.NameInbox]; hasService {
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
		Status:              status,
		TotalInstances:      total,
		EnabledInstances:    enabled,
		ConnectedInstances:  connected,
	}
}

// HealthStatus representa o status de saúde da integração
type HealthStatus struct {
	Status             string `json:"status"`
	TotalInstances     int    `json:"totalInstances"`
	EnabledInstances   int    `json:"enabledInstances"`
	ConnectedInstances int    `json:"connectedInstances"`
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
	if !config.Enabled {
		return nil // Configuração desabilitada é válida
	}

	if config.AccountID == "" {
		return fmt.Errorf("account ID is required when Chatwoot is enabled")
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

	if !config.Enabled {
		return fmt.Errorf("chatwoot is disabled")
	}

	// Cria cliente temporário para teste
	client := NewClient(config.URL, config.Token, config.AccountID)

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
		TotalInstances:     len(i.configs),
		EnabledInstances:   0,
		ConnectedInstances: 0,
		InstanceMetrics:    make(map[string]InstanceMetrics),
	}

	for instanceName, config := range i.configs {
		instanceMetric := InstanceMetrics{
			Enabled:   config.Enabled,
			Connected: false,
		}

		if config.Enabled {
			metrics.EnabledInstances++
			
			if _, hasService := i.services[instanceName]; hasService {
				metrics.ConnectedInstances++
				instanceMetric.Connected = true
			}
		}

		metrics.InstanceMetrics[instanceName] = instanceMetric
	}

	return metrics
}

// Metrics representa métricas da integração
type Metrics struct {
	TotalInstances     int                        `json:"totalInstances"`
	EnabledInstances   int                        `json:"enabledInstances"`
	ConnectedInstances int                        `json:"connectedInstances"`
	InstanceMetrics    map[string]InstanceMetrics `json:"instanceMetrics"`
}

// InstanceMetrics representa métricas de uma instância específica
type InstanceMetrics struct {
	Enabled   bool `json:"enabled"`
	Connected bool `json:"connected"`
	// Aqui você pode adicionar mais métricas específicas como:
	// MessagesProcessed int `json:"messagesProcessed"`
	// LastActivity      time.Time `json:"lastActivity"`
	// ErrorCount        int `json:"errorCount"`
}
