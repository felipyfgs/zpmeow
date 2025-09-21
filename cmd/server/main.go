















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

	appContactService := application.NewContactApp(sessionRepo, wmeowService)
	appChatService := application.NewChatApp(sessionRepo, wmeowService)
	appGroupService := application.NewGroupApp(sessionRepo, wmeowService)

	sessionHandler := handlers.NewSessionHandler(appSessionService, wmeowService)
	healthHandler := handlers.NewHealthHandler(db)
	messageHandler := handlers.NewMessageHandler(appSessionService, wmeowService)
	chatHandler := handlers.NewChatHandler(appChatService, wmeowService)
	groupHandler := handlers.NewGroupHandler(appGroupService, wmeowService)
	communityHandler := handlers.NewCommunityHandler(appSessionService, wmeowService)
	webhookHandler := handlers.NewWebhookHandler(appSessionService, webhookAppService, wmeowService)
	contactHandler := handlers.NewContactHandler(appContactService, wmeowService)
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
