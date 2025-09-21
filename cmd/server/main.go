// Package main provides the entry point for the WhatsApp API server
//
// @title WhatsApp API Gateway
// @version 1.0
// @description A comprehensive REST API for WhatsApp Business integration built with Go and whatsmeow
// @termsOfService http://swagger.io/terms/
//
// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io
//
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
//
// @host localhost:8080
// @BasePath /
//
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Enter your API key directly
//
// @tag.name Health
// @tag.description Health check and monitoring endpoints
//
// @tag.name Sessions
// @tag.description WhatsApp session management endpoints
//
// @tag.name Messages
// @tag.description Message sending and management endpoints
//
// @tag.name Privacy
// @tag.description Privacy settings and configuration endpoints
//
// @tag.name Chat
// @tag.description Chat management and history endpoints
//
// @tag.name Contacts
// @tag.description Contact management and presence endpoints
//
// @tag.name Groups
// @tag.description WhatsApp group management endpoints
//
// @tag.name Communities
// @tag.description WhatsApp community management endpoints
//
// @tag.name Newsletters
// @tag.description Newsletter management and broadcasting endpoints
//
// @tag.name Webhooks
// @tag.description Webhook configuration and event management endpoints
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "zpmeow/docs"
	"zpmeow/internal/application"
	"zpmeow/internal/config"
	"zpmeow/internal/domain/session"
	"zpmeow/internal/infra/cache"
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

	// Initialize cache service
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

	// Initialize repositories
	baseSessionRepo := repository.NewPostgresRepo(db)
	sessionRepo := cache.NewCachedSessionRepository(baseSessionRepo, cacheService)

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

	// Initialize handlers in the specified order:
	// Health, Sessions, Messages, Privacy, Chat, Contacts, Groups, Communities, Newsletters, Webhooks
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

	gin.SetMode(cfg.GetServer().GetMode())

	ginRouter := gin.New()

	// Add correlation ID middleware first
	ginRouter.Use(middleware.CorrelationIDMiddleware())

	// Add structured logging middleware
	ginRouter.Use(middleware.Logger())

	ginRouter.Use(middleware.CORS(cfg.GetCORS()))

	// Handler dependencies in the specified order:
	// Health, Sessions, Messages, Privacy, Chat, Contacts, Groups, Communities, Newsletters, Webhooks
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
