//	@title			meow meow API
//	@version		1.0
//	@description	A meow API server built with Go, inspired by meow
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	meow API Support
//	@contact.url	https://github.com/your-username/meow
//	@contact.email	support@meow.com

//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT

//	@host		localhost:8080
//	@BasePath	/

//	@schemes	http https

// @tag.name		Health
// @tag.description	Health check and monitoring endpoints

// @tag.name		Sessions
// @tag.description	Session management endpoints

// @tag.name		Messages
// @tag.description	Message sending and management endpoints

// @tag.name		Privacy
// @tag.description	Privacy settings and blocklist management endpoints

// @tag.name		Webhooks
// @tag.description	Webhook configuration and event management endpoints

// @tag.name		Chat
// @tag.description	Chat management and interaction endpoints

// @tag.name		Groups
// @tag.description	Group management and administration endpoints

// @tag.name		Community
// @tag.description	Community management and linking endpoints

// @tag.name		Contacts
// @tag.description	Contact management and information endpoints

// @tag.name		Media
// @tag.description	Media upload, download and management endpoints

// @tag.name		Newsletters
// @tag.description	Newsletter creation and management endpoints

// @securityDefinitions.apikey	ApiKeyAuth
// @in							header
// @name						Authorization
// @description				API Key authentication. Simply provide your API key directly: "YOUR_API_KEY". The system automatically detects if it's a Global API Key (can access all sessions and session management) or a Session-specific API Key (can only access the specific session it belongs to).
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "zpmeow/docs" // Import for swagger docs
	"zpmeow/internal/application"
	"zpmeow/internal/config"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/database"
	"zpmeow/internal/infra/database/repository"
	"zpmeow/internal/infra/http/handlers"
	"zpmeow/internal/infra/http/middleware"
	"zpmeow/internal/infra/http/routes"
	"zpmeow/internal/infra/logging"
	"zpmeow/internal/infra/webhooks"
	"zpmeow/internal/infra/wmeow"

	"github.com/gin-gonic/gin"
	"go.mau.fi/whatsmeow/store/sqlstore"
)

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

	sessionRepo := repository.NewPostgresRepo(db)

	_ = webhooks.NewService()

	waLogger := logging.GetWALogger("WhatsApp")
	wmeowService := wmeow.NewMeowService(container, waLogger, sessionRepo)

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

	sessionHandler := handlers.NewSessionHandler(appSessionService, wmeowService)
	healthHandler := handlers.NewHealthHandler(db)
	messageHandler := handlers.NewMessageHandler(appSessionService, wmeowService)
	chatHandler := handlers.NewChatHandler(appSessionService, wmeowService)
	groupHandler := handlers.NewGroupHandler(appSessionService, wmeowService)
	communityHandler := handlers.NewCommunityHandler(appSessionService, wmeowService)
	webhookHandler := handlers.NewWebhookHandler(appSessionService, webhookAppService, wmeowService)
	contactHandler := handlers.NewContactHandler(appSessionService, wmeowService)
	newsletterHandler := handlers.NewNewsletterHandler(appSessionService, wmeowService)
	privacyHandler := handlers.NewPrivacyHandler(appSessionService, wmeowService)

	gin.SetMode(cfg.GetServer().GetMode())

	ginRouter := gin.New()

	ginRouter.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/ping", "/health"}, // Skip health check logs
		Formatter: func(param gin.LogFormatterParams) string {
			if param.StatusCode >= 400 ||
				(param.Path != "/ping" && param.Path != "/health") {
				return fmt.Sprintf("%s - %s %s %d %s\n",
					param.TimeStamp.Format("15:04:05"),
					param.Method,
					param.Path,
					param.StatusCode,
					param.Latency,
				)
			}
			return ""
		},
	}))

	ginRouter.Use(middleware.CORS(cfg.GetCORS()))

	handlerDeps := &routes.HandlerDependencies{
		HealthHandler:     healthHandler,
		SessionHandler:    sessionHandler,
		ContactHandler:    contactHandler,
		ChatHandler:       chatHandler,
		MessageHandler:    messageHandler,
		GroupHandler:      groupHandler,
		CommunityHandler:  communityHandler,
		NewsletterHandler: newsletterHandler,
		WebhookHandler:    webhookHandler,
		PrivacyHandler:    privacyHandler,
	}

	routes.SetupRoutes(ginRouter, handlerDeps, authMiddleware)

	addr := fmt.Sprintf(":%s", cfg.GetServer().GetPort())
	srv := &http.Server{
		Addr:         addr,
		Handler:      ginRouter,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Infof("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server forced to shutdown: %v", err)
	}

	log.Info("Server exited")
}
