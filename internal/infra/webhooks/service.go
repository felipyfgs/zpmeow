package webhooks

import (
	"context"
	"fmt"
	"time"

	"zpmeow/internal/infra/logging"
)

type Service struct {
	httpClient    HTTPClient
	retryStrategy *RetryStrategy
	logger        logging.Logger
}

func NewService() *Service {
	retryConfig := &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}

	return &Service{
		httpClient:    NewWebhookHTTPClient(30 * time.Second),
		retryStrategy: NewRetryStrategy(retryConfig),
		logger:        logging.GetLogger().Sub("webhook-service"),
	}
}

func NewServiceWithConfig(timeout time.Duration, retryConfig *RetryConfig) *Service {
	return &Service{
		httpClient:    NewWebhookHTTPClient(timeout),
		retryStrategy: NewRetryStrategy(retryConfig),
		logger:        logging.GetLogger().Sub("webhook-service"),
	}
}

func (w *Service) SendWebhook(ctx context.Context, webhookURL, event, sessionID string, data interface{}) error {
	if webhookURL == "" {
		return fmt.Errorf("webhooks: URL is empty")
	}

	w.logger.Infof("Sending webhook to %s for event %s (session: %s)", webhookURL, event, sessionID)

	err := w.httpClient.Post(ctx, webhookURL, data, nil)
	if err != nil {
		w.logger.Errorf("Failed to send webhook to %s: %v", webhookURL, err)
		return fmt.Errorf("webhooks: failed to send to %s: %w", webhookURL, err)
	}

	w.logger.Infof("Successfully sent webhook to %s", webhookURL)
	return nil
}

func (w *Service) SendWebhookWithRetry(ctx context.Context, webhookURL, event, sessionID string, data interface{}) error {
	if webhookURL == "" {
		return fmt.Errorf("webhooks: URL is empty")
	}

	operationName := fmt.Sprintf("webhook to %s for event %s", webhookURL, event)

	err := w.retryStrategy.ExecuteWithRetry(ctx, func() error {
		w.logger.Debugf("Attempting to send %s (session: %s)", operationName, sessionID)
		return w.httpClient.Post(ctx, webhookURL, data, nil)
	}, operationName)

	if err != nil {
		return fmt.Errorf("webhooks: retry failed for %s: %w", operationName, err)
	}

	return nil
}

func (w *Service) SendWebhookAsync(webhookURL, event, sessionID string, data interface{}) error {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		err := w.SendWebhookWithRetry(ctx, webhookURL, event, sessionID, data)
		if err != nil {
			w.logger.Errorf("Async webhook failed: %v", err)
		}
	}()
	return nil
}

func (w *Service) SendWebhookWithHeaders(ctx context.Context, webhookURL, event, sessionID string, data interface{}, headers map[string]string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}

	w.logger.Infof("Sending webhook with headers to %s for event %s (session: %s)", webhookURL, event, sessionID)

	err := w.httpClient.Post(ctx, webhookURL, data, headers)
	if err != nil {
		w.logger.Errorf("Failed to send webhook to %s: %v", webhookURL, err)
		return err
	}

	w.logger.Infof("Successfully sent webhook to %s", webhookURL)
	return nil
}

func (w *Service) SendWebhookWithHeadersAndRetry(ctx context.Context, webhookURL, event, sessionID string, data interface{}, headers map[string]string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL is empty")
	}

	operationName := fmt.Sprintf("webhook with headers to %s for event %s", webhookURL, event)

	return w.retryStrategy.ExecuteWithRetry(ctx, func() error {
		w.logger.Debugf("Attempting to send %s (session: %s)", operationName, sessionID)
		return w.httpClient.Post(ctx, webhookURL, data, headers)
	}, operationName)
}
