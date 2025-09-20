package webhook

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/common"
	"zpmeow/internal/application/ports"
	"zpmeow/internal/domain/session"
)

type ConfigureWebhookCommand struct {
	SessionID string
	URL       string
	Events    []string
	Enabled   bool
}

func (c ConfigureWebhookCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if c.Enabled {
		if strings.TrimSpace(c.URL) == "" {
			return common.NewValidationError("url", c.URL, "webhook URL is required when enabled")
		}

		if !strings.HasPrefix(c.URL, "http://") && !strings.HasPrefix(c.URL, "https://") {
			return common.NewValidationError("url", c.URL, "webhook URL must start with http:// or https://")
		}

		if len(c.Events) == 0 {
			return common.NewValidationError("events", "", "at least one event type is required when webhook is enabled")
		}
	}

	return nil
}

type ConfigureWebhookResult struct {
	SessionID string
	URL       string
	Events    []string
	Enabled   bool
	Success   bool
}

type ConfigureWebhookUseCase struct {
	sessionRepo session.Repository
	logger      ports.Logger
}

func NewConfigureWebhookUseCase(
	sessionRepo session.Repository,
	logger ports.Logger,
) *ConfigureWebhookUseCase {
	return &ConfigureWebhookUseCase{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (uc *ConfigureWebhookUseCase) Handle(ctx context.Context, cmd ConfigureWebhookCommand) (*ConfigureWebhookResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid configure webhook command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sessionEntity, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if cmd.Enabled && cmd.URL != "" {
		if err := sessionEntity.SetWebhookEndpoint(cmd.URL); err != nil {
			uc.logger.Error(ctx, "Failed to set webhook endpoint",
				"sessionID", cmd.SessionID,
				"url", cmd.URL,
				"error", err)
			return nil, fmt.Errorf("failed to set webhook endpoint: %w", err)
		}
	} else {
		if err := sessionEntity.SetWebhookEndpoint(""); err != nil {
			uc.logger.Error(ctx, "Failed to clear webhook endpoint",
				"sessionID", cmd.SessionID,
				"error", err)
			return nil, fmt.Errorf("failed to clear webhook endpoint: %w", err)
		}
	}

	if err := uc.sessionRepo.Update(ctx, sessionEntity); err != nil {
		uc.logger.Error(ctx, "Failed to update session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	uc.logger.Info(ctx, "Webhook configured successfully",
		"sessionID", cmd.SessionID,
		"url", cmd.URL,
		"enabled", cmd.Enabled,
		"eventCount", len(cmd.Events))

	return &ConfigureWebhookResult{
		SessionID: cmd.SessionID,
		URL:       cmd.URL,
		Events:    cmd.Events,
		Enabled:   cmd.Enabled,
		Success:   true,
	}, nil
}

type TestWebhookCommand struct {
	SessionID string
	URL       string
}

func (c TestWebhookCommand) Validate() error {
	if strings.TrimSpace(c.SessionID) == "" {
		return common.NewValidationError("sessionID", c.SessionID, "session ID is required")
	}

	if strings.TrimSpace(c.URL) == "" {
		return common.NewValidationError("url", c.URL, "webhook URL is required")
	}

	if !strings.HasPrefix(c.URL, "http://") && !strings.HasPrefix(c.URL, "https://") {
		return common.NewValidationError("url", c.URL, "webhook URL must start with http:// or https://")
	}

	return nil
}

type TestWebhookResult struct {
	SessionID    string
	URL          string
	Success      bool
	ResponseCode int
	Message      string
}

type TestWebhookUseCase struct {
	sessionRepo         session.Repository
	notificationService ports.NotificationService
	logger              ports.Logger
}

func NewTestWebhookUseCase(
	sessionRepo session.Repository,
	notificationService ports.NotificationService,
	logger ports.Logger,
) *TestWebhookUseCase {
	return &TestWebhookUseCase{
		sessionRepo:         sessionRepo,
		notificationService: notificationService,
		logger:              logger,
	}
}

func (uc *TestWebhookUseCase) Handle(ctx context.Context, cmd TestWebhookCommand) (*TestWebhookResult, error) {
	if err := cmd.Validate(); err != nil {
		uc.logger.Warn(ctx, "Invalid test webhook command", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	_, err := uc.sessionRepo.GetByID(ctx, cmd.SessionID)
	if err != nil {
		uc.logger.Error(ctx, "Failed to get session", "sessionID", cmd.SessionID, "error", err)
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	testPayload := map[string]interface{}{
		"event":     "webhook.test",
		"sessionID": cmd.SessionID,
		"timestamp": "2024-01-01T00:00:00Z",
		"data": map[string]interface{}{
			"message": "This is a test webhook from WhatsApp API",
		},
	}

	if err := uc.notificationService.SendWebhook(ctx, cmd.URL, testPayload); err != nil {
		uc.logger.Error(ctx, "Failed to send test webhook",
			"sessionID", cmd.SessionID,
			"url", cmd.URL,
			"error", err)
		return &TestWebhookResult{
			SessionID:    cmd.SessionID,
			URL:          cmd.URL,
			Success:      false,
			ResponseCode: 0,
			Message:      fmt.Sprintf("Failed to send webhook: %v", err),
		}, nil
	}

	uc.logger.Info(ctx, "Test webhook sent successfully",
		"sessionID", cmd.SessionID,
		"url", cmd.URL)

	return &TestWebhookResult{
		SessionID:    cmd.SessionID,
		URL:          cmd.URL,
		Success:      true,
		ResponseCode: 200,
		Message:      "Test webhook sent successfully",
	}, nil
}
