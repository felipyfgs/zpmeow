package chatwoot

import (
	"context"
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
)

// ExampleBasicUsage demonstra uso básico da integração Chatwoot
func ExampleBasicUsage() {
	// Configuração básica
	config := &ChatwootConfig{
		Enabled:             true,
		AccountID:           "1",
		Token:              "seu-token-aqui",
		URL:                "https://app.chatwoot.com",
		NameInbox:          "WhatsApp zpmeow",
		ReopenConversation: true,
		ConversationPending: false,
		MergeBrazilContacts: true,
		AutoCreate:         true,
		Organization:       "Minha Empresa",
		Logo:               "https://exemplo.com/logo.png",
		IgnoreJids: []string{
			"status@broadcast",
		},
	}

	// Criar logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Criar serviço
	service, err := NewService(config, logger, nil, "example-session")
	if err != nil {
		logger.Error("Failed to create Chatwoot service", "error", err)
		return
	}

	// Simular mensagem do WhatsApp
	msg := &WhatsAppMessage{
		ID:        "msg_123456789",
		From:      "5511999999999@s.whatsapp.net",
		Body:      "Olá! Preciso de ajuda com meu pedido.",
		Type:      "text",
		FromMe:    false,
		PushName:  "João Silva",
		Timestamp: 1640995200, // 2022-01-01 00:00:00
	}

	// Processar mensagem
	ctx := context.Background()
	err = service.ProcessWhatsAppMessage(ctx, msg)
	if err != nil {
		logger.Error("Failed to process WhatsApp message", "error", err)
		return
	}

	logger.Info("Message processed successfully")
}

// ExampleGroupMessage demonstra processamento de mensagem de grupo
func ExampleGroupMessage() {
	config := &ChatwootConfig{
		Enabled:   true,
		AccountID: "1",
		Token:     "seu-token",
		URL:       "https://app.chatwoot.com",
		NameInbox: "WhatsApp Groups",
	}

	logger := slog.Default()
	service, _ := NewService(config, logger, nil, "example-session")

	// Mensagem de grupo
	groupMsg := &WhatsAppMessage{
		ID:          "group_msg_123",
		From:        "120363025343298765@g.us", // JID do grupo
		Participant: "5511999999999@s.whatsapp.net", // Participante que enviou
		Body:        "Pessoal, alguém pode me ajudar?",
		Type:        "text",
		FromMe:      false,
		PushName:    "Maria Santos",
		ChatName:    "Suporte Técnico",
	}

	ctx := context.Background()
	err := service.ProcessWhatsAppMessage(ctx, groupMsg)
	if err != nil {
		logger.Error("Failed to process group message", "error", err)
	}
}

// ExampleMediaMessage demonstra processamento de mensagem com mídia
func ExampleMediaMessage() {
	config := &ChatwootConfig{
		Enabled:   true,
		AccountID: "1",
		Token:     "seu-token",
		URL:       "https://app.chatwoot.com",
	}

	logger := slog.Default()
	service, _ := NewService(config, logger, nil, "example-session")

	// Mensagem com imagem
	mediaMsg := &WhatsAppMessage{
		ID:       "media_msg_123",
		From:     "5511999999999@s.whatsapp.net",
		Type:     "image",
		Caption:  "Aqui está a foto do problema",
		MediaURL: "https://exemplo.com/imagem.jpg",
		FileName: "problema.jpg",
		FromMe:   false,
		PushName: "Cliente",
	}

	ctx := context.Background()
	err := service.ProcessWhatsAppMessage(ctx, mediaMsg)
	if err != nil {
		logger.Error("Failed to process media message", "error", err)
	}
}

// ExampleLocationMessage demonstra processamento de mensagem de localização
func ExampleLocationMessage() {
	config := &ChatwootConfig{
		Enabled:   true,
		AccountID: "1",
		Token:     "seu-token",
		URL:       "https://app.chatwoot.com",
	}

	logger := slog.Default()
	service, _ := NewService(config, logger, nil, "example-session")

	// Mensagem de localização
	locationMsg := &WhatsAppMessage{
		ID:     "location_msg_123",
		From:   "5511999999999@s.whatsapp.net",
		Type:   "location",
		FromMe: false,
		Location: &LocationInfo{
			Latitude:  -23.5505,
			Longitude: -46.6333,
			Name:      "São Paulo",
			Address:   "São Paulo, SP, Brasil",
		},
		PushName: "Cliente",
	}

	ctx := context.Background()
	err := service.ProcessWhatsAppMessage(ctx, locationMsg)
	if err != nil {
		logger.Error("Failed to process location message", "error", err)
	}
}

// ExampleContactMessage demonstra processamento de mensagem de contato
func ExampleContactMessage() {
	config := &ChatwootConfig{
		Enabled:   true,
		AccountID: "1",
		Token:     "seu-token",
		URL:       "https://app.chatwoot.com",
	}

	logger := slog.Default()
	service, _ := NewService(config, logger, nil, "example-session")

	// Mensagem de contato
	contactMsg := &WhatsAppMessage{
		ID:     "contact_msg_123",
		From:   "5511999999999@s.whatsapp.net",
		Type:   "contact",
		FromMe: false,
		Contacts: []ContactInfo{
			{
				Name: "Dr. João Silva",
				Phones: []PhoneInfo{
					{Number: "+5511888888888", Type: "work"},
					{Number: "+5511777777777", Type: "mobile"},
				},
			},
		},
		PushName: "Cliente",
	}

	ctx := context.Background()
	err := service.ProcessWhatsAppMessage(ctx, contactMsg)
	if err != nil {
		logger.Error("Failed to process contact message", "error", err)
	}
}

// ExampleWebhookProcessing demonstra processamento de webhook do Chatwoot
func ExampleWebhookProcessing() {
	config := &ChatwootConfig{
		Enabled:   true,
		AccountID: "1",
		Token:     "seu-token",
		URL:       "https://app.chatwoot.com",
	}

	logger := slog.Default()
	service, _ := NewService(config, logger, nil, "example-session")

	// Simular webhook do Chatwoot
	webhook := &WebhookPayload{
		Event:       "message_created",
		MessageType: "outgoing",
		Content:     "Olá! Como posso ajudá-lo hoje?",
		Contact: &Contact{
			ID:          123,
			PhoneNumber: "+5511999999999",
			Name:        "João Silva",
		},
		Conversation: &Conversation{
			ID:      456,
			InboxID: 789,
			Status:  "open",
		},
	}

	ctx := context.Background()
	err := service.ProcessWebhook(ctx, webhook)
	if err != nil {
		logger.Error("Failed to process webhook", "error", err)
	}
}

// ExampleHTTPIntegration demonstra integração com servidor HTTP
func ExampleHTTPIntegration() {
	// Criar app Fiber
	app := fiber.New()

	// Configurar logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Criar integração Chatwoot
	integration := NewIntegration(logger)

	// Configurar instância de exemplo
	config := &ChatwootConfig{
		Enabled:             true,
		AccountID:           "1",
		Token:              "seu-token",
		URL:                "https://app.chatwoot.com",
		NameInbox:          "WhatsApp API",
		ReopenConversation: true,
		AutoCreate:         true,
	}

	err := integration.RegisterInstance("exemplo", config)
	if err != nil {
		logger.Error("Failed to register instance", "error", err)
		return
	}

	// Middleware para log de requisições
	app.Use(func(c *fiber.Ctx) error {
		logger.Info("Request",
			"method", c.Method(),
			"path", c.Path(),
			"ip", c.IP(),
		)
		return c.Next()
	})

	// Iniciar servidor
	logger.Info("Starting HTTP server on :8080")
	app.Listen(":8080")
}

// ExampleMultipleInstances demonstra gerenciamento de múltiplas instâncias
func ExampleMultipleInstances() {
	logger := slog.Default()
	integration := NewIntegration(logger)

	// Configurar primeira instância (vendas)
	salesConfig := &ChatwootConfig{
		Enabled:   true,
		AccountID: "1",
		Token:     "token-vendas",
		URL:       "https://vendas.chatwoot.com",
		NameInbox: "WhatsApp Vendas",
	}

	err := integration.RegisterInstance("vendas", salesConfig)
	if err != nil {
		logger.Error("Failed to register sales instance", "error", err)
		return
	}

	// Configurar segunda instância (suporte)
	supportConfig := &ChatwootConfig{
		Enabled:   true,
		AccountID: "2",
		Token:     "token-suporte",
		URL:       "https://suporte.chatwoot.com",
		NameInbox: "WhatsApp Suporte",
	}

	err = integration.RegisterInstance("suporte", supportConfig)
	if err != nil {
		logger.Error("Failed to register support instance", "error", err)
		return
	}

	// Processar mensagem para instância específica
	msg := &WhatsAppMessage{
		ID:       "msg_123",
		From:     "5511999999999@s.whatsapp.net",
		Body:     "Preciso de suporte técnico",
		FromMe:   false,
		PushName: "Cliente",
	}

	ctx := context.Background()
	err = integration.ProcessMessage(ctx, "suporte", msg)
	if err != nil {
		logger.Error("Failed to process message", "error", err)
	}

	// Listar instâncias habilitadas
	enabled := integration.GetEnabledInstances()
	logger.Info("Enabled instances", "instances", enabled)
}

// ExampleCustomFormatting demonstra formatação personalizada de mensagens
func ExampleCustomFormatting() {
	config := &ChatwootConfig{
		Enabled:       true,
		AccountID:     "1",
		Token:         "seu-token",
		URL:           "https://app.chatwoot.com",
		SignMsg:       true,
		SignDelimiter: "\n\n---\n📱 Enviado via zpmeow",
	}

	logger := slog.Default()
	service, _ := NewService(config, logger, nil, "example-session")

	// Mensagem com formatação WhatsApp
	formattedMsg := &WhatsAppMessage{
		ID:       "formatted_msg_123",
		From:     "5511999999999@s.whatsapp.net",
		Body:     "*Urgente*: Preciso de _ajuda_ com ~problema~ no sistema",
		Type:     "text",
		FromMe:   false,
		PushName: "Cliente VIP",
	}

	ctx := context.Background()
	err := service.ProcessWhatsAppMessage(ctx, formattedMsg)
	if err != nil {
		logger.Error("Failed to process formatted message", "error", err)
	}
}

// ExampleErrorHandling demonstra tratamento de erros
func ExampleErrorHandling() {
	// Configuração com dados inválidos para demonstrar tratamento de erro
	invalidConfig := &ChatwootConfig{
		Enabled:   true,
		AccountID: "",        // Inválido
		Token:     "",        // Inválido
		URL:       "invalid", // Inválido
	}

	logger := slog.Default()

	// Tentar criar serviço com configuração inválida
	service, err := NewService(invalidConfig, logger, nil, "example-session")
	if err != nil {
		logger.Error("Expected error creating service", "error", err)
		return
	}

	// Tentar processar mensagem com serviço inválido
	msg := &WhatsAppMessage{
		ID:   "test_msg",
		From: "5511999999999@s.whatsapp.net",
		Body: "Test message",
	}

	ctx := context.Background()
	err = service.ProcessWhatsAppMessage(ctx, msg)
	if err != nil {
		logger.Error("Expected error processing message", "error", err)
	}
}

// ExampleConfigValidation demonstra validação de configuração
func ExampleConfigValidation() {
	logger := slog.Default()

	// Configurações para testar
	configs := []*ChatwootConfig{
		// Configuração válida
		{
			Enabled:   true,
			AccountID: "1",
			Token:     "valid-token",
			URL:       "https://app.chatwoot.com",
		},
		// Configuração inválida - sem token
		{
			Enabled:   true,
			AccountID: "1",
			Token:     "",
			URL:       "https://app.chatwoot.com",
		},
		// Configuração inválida - URL inválida
		{
			Enabled:   true,
			AccountID: "1",
			Token:     "valid-token",
			URL:       "invalid-url",
		},
		// Configuração desabilitada (válida)
		{
			Enabled: false,
		},
	}

	for i, config := range configs {
		logger.Info("Testing configuration", "index", i)
		
		service, err := NewService(config, logger, nil, "example-session")
		if err != nil {
			logger.Error("Configuration validation failed", "index", i, "error", err)
			continue
		}

		if service != nil {
			logger.Info("Configuration valid", "index", i)
		}
	}
}
