// @title WhatsApp API Gateway
// @version 1.0
// @description A comprehensive REST API for WhatsApp Business integration built with Go and whatsmeow
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Enter your API key directly
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "zpmeow/docs"
	"zpmeow/internal/application"
	"zpmeow/internal/config"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/cache"
	"zpmeow/internal/infra/chatwoot"
	"zpmeow/internal/infra/database"
	"zpmeow/internal/infra/database/repository"
	"zpmeow/internal/infra/http/handlers"
	"zpmeow/internal/infra/http/middleware"
	"zpmeow/internal/infra/http/routes"
	"zpmeow/internal/infra/logging"
	"zpmeow/internal/infra/webhooks"
	"zpmeow/internal/infra/wmeow"

	"github.com/gofiber/fiber/v2"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

// @tag.name Health
// @tag.description Health check and monitoring endpoints
// @tag.name Sessions
// @tag.description WhatsApp session management endpoints
// @tag.name Messages
// @tag.description Message sending and management endpoints
// @tag.name Privacy
// @tag.description Privacy settings and configuration endpoints
// @tag.name Chat
// @tag.description Chat management and history endpoints
// @tag.name Contacts
// @tag.description Contact management and presence endpoints
// @tag.name Groups
// @tag.description WhatsApp group management endpoints
// @tag.name Communities
// @tag.description WhatsApp community management endpoints
// @tag.name Newsletters
// @tag.description Newsletter management and broadcasting endpoints
// @tag.name Webhooks
// @tag.description Webhook configuration and event management endpoints
func main() {
	cfg, err := config.LoadConfig()
	if err != nil {

		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	log := logging.Initialize(cfg.GetLogging())
	logging.SetLogger(log)
	log.Info("Starting meow server")

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Errorf("Error closing database connection: %v", err)
		}
	}()

	if err := database.RunMigrations(cfg); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	dbLog := logging.GetWALogger("Database")
	container, err := sqlstore.New(context.Background(), "postgres", cfg.GetDatabaseURL(), dbLog)
	if err != nil {
		log.Fatalf("Failed to create whatsmeow container: %v", err)
	}

	cacheService, err := cache.NewRedisService(cfg.GetCache())
	if err != nil {
		log.Fatalf("Failed to initialize cache service: %v", err)
	}
	defer func() {
		if closer, ok := cacheService.(*cache.RedisService); ok {
			if err := closer.Close(); err != nil {
				log.Errorf("Error closing cache service: %v", err)
			}
		}
	}()

	baseSessionRepo := repository.NewPostgresRepo(db)
	sessionRepo := cache.NewCachedSessionRepository(baseSessionRepo, cacheService)

	_ = webhooks.NewService()

	waLogger := logging.GetWALogger("WhatsApp")

	// Chatwoot integration (criar antes do wmeowService)
	chatwootLogger := slog.Default().With("component", "chatwoot")
	chatwootRepo := repository.NewChatwootRepository(db)
	chatwootIntegration := chatwoot.NewIntegration(chatwootLogger)

	// Carregar configurações existentes do Chatwoot
	if err := loadChatwootConfigurations(context.Background(), chatwootIntegration, chatwootRepo, chatwootLogger); err != nil {
		log.Errorf("Failed to load Chatwoot configurations: %v", err)
	}

	// Criar wmeowService com integração Chatwoot
	wmeowService := wmeow.NewMeowServiceWithChatwoot(container, waLogger, sessionRepo, chatwootIntegration, chatwootRepo)

	domainService := session.NewService()

	appSessionService := application.NewSessionApp(sessionRepo, domainService)
	webhookAppService := application.NewWebhookApp(sessionRepo)

	log.Info("WhatsApp service initialized")
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		if err := wmeowService.ConnectOnStartup(ctx); err != nil {
			log.Errorf("ConnectOnStartup failed: %v", err)
		} else {
			log.Info("ConnectOnStartup completed")
		}
	}()

	log.Info("Session service initialized")

	authMiddleware := middleware.NewAuthMiddleware(cfg, sessionRepo, log)

	appContactService := application.NewContactApp(sessionRepo, wmeowService)
	appChatService := application.NewChatApp(sessionRepo, wmeowService)
	appGroupService := application.NewGroupApp(sessionRepo, wmeowService)

	healthHandler := handlers.NewHealthHandlerWithCache(db, cacheService)
	sessionHandler := handlers.NewSessionHandler(appSessionService, wmeowService)
	messageHandler := handlers.NewMessageHandler(appSessionService, wmeowService)
	privacyHandler := handlers.NewPrivacyHandler(appSessionService, wmeowService)
	chatHandler := handlers.NewChatHandler(appChatService, wmeowService)
	contactHandler := handlers.NewContactHandler(appContactService, wmeowService)
	groupHandler := handlers.NewGroupHandler(appGroupService, wmeowService)
	communityHandler := handlers.NewCommunityHandler(appSessionService, wmeowService)
	newsletterHandler := handlers.NewNewsletterHandler(appSessionService, wmeowService)
	webhookHandler := handlers.NewWebhookHandler(appSessionService, webhookAppService, wmeowService)

	// Chatwoot handler (usando as instâncias já criadas)
	chatwootHandler := handlers.NewChatwootHandler(appSessionService, chatwootIntegration, chatwootRepo, wmeowService)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(middleware.CorrelationIDMiddleware())

	app.Use(middleware.Logger())

	app.Use(middleware.CORS(cfg.GetCORS()))

	handlerDeps := &routes.HandlerDependencies{
		HealthHandler:     healthHandler,
		SessionHandler:    sessionHandler,
		MessageHandler:    messageHandler,
		PrivacyHandler:    privacyHandler,
		ChatHandler:       chatHandler,
		ContactHandler:    contactHandler,
		GroupHandler:      groupHandler,
		CommunityHandler:  communityHandler,
		NewsletterHandler: newsletterHandler,
		WebhookHandler:    webhookHandler,
		ChatwootHandler:   chatwootHandler,
	}

	routes.SetupRoutes(app, handlerDeps, authMiddleware)

	addr := fmt.Sprintf(":%s", cfg.GetServer().GetPort())

	go func() {
		log.Infof("Server listening on %s", addr)
		if err := app.Listen(addr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}

// loadChatwootConfigurations carrega todas as configurações habilitadas do Chatwoot na inicialização
func loadChatwootConfigurations(ctx context.Context, integration *chatwoot.Integration, repo *repository.ChatwootRepository, logger *slog.Logger) error {
	// Busca todas as configurações habilitadas
	enabled := true
	configs, err := repo.List(ctx, &enabled)
	if err != nil {
		return fmt.Errorf("failed to list enabled Chatwoot configurations: %w", err)
	}

	logger.Info("Loading Chatwoot configurations", "count", len(configs))

	for _, dbConfig := range configs {
		// Converte modelo do banco para configuração
		config := &chatwoot.ChatwootConfig{
			Enabled:                    dbConfig.Enabled,
			SignMsg:                    dbConfig.SignMsg,
			SignDelimiter:              dbConfig.SignDelimiter,
			Number:                     dbConfig.Number,
			ReopenConversation:         dbConfig.ReopenConversation,
			ConversationPending:        dbConfig.ConversationPending,
			MergeBrazilContacts:        dbConfig.MergeBrazilContacts,
			ImportContacts:             dbConfig.ImportContacts,
			ImportMessages:             dbConfig.ImportMessages,
			DaysLimitImportMessages:    dbConfig.DaysLimitImportMessages,
			AutoCreate:                 dbConfig.AutoCreate,
			Organization:               dbConfig.Organization,
			Logo:                       dbConfig.Logo,
			IgnoreJids:                 []string(dbConfig.IgnoreJids),
		}

		// Campos opcionais (ponteiros)
		if dbConfig.AccountID != nil {
			config.AccountID = *dbConfig.AccountID
		}
		if dbConfig.Token != nil {
			config.Token = *dbConfig.Token
		}
		if dbConfig.URL != nil {
			config.URL = *dbConfig.URL
		}
		if dbConfig.NameInbox != nil {
			config.NameInbox = *dbConfig.NameInbox
		}

		// Registra a configuração na integração
		if err := integration.RegisterInstance(dbConfig.SessionID, config); err != nil {
			logger.Error("Failed to register Chatwoot instance",
				"sessionID", dbConfig.SessionID,
				"error", err)
			continue
		}

		logger.Info("Chatwoot configuration loaded",
			"sessionID", dbConfig.SessionID,
			"accountID", dbConfig.AccountID,
			"url", dbConfig.URL)
	}

	logger.Info("Chatwoot configurations loaded successfully", "loaded", len(configs))
	return nil
}
