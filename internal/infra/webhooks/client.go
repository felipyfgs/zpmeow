package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HTTPClient interface {
	Post(ctx context.Context, url string, payload interface{}, headers map[string]string) error
}

type WebhookHTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

func NewWebhookHTTPClient(timeout time.Duration) *WebhookHTTPClient {
	return &WebhookHTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

func (c *WebhookHTTPClient) Post(ctx context.Context, url string, payload interface{}, headers map[string]string) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhooks: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("webhooks: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "meow-webhook-client/1.0")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhooks: failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("Warning: failed to close response body: %v\n", err)
		}
	}()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		body = []byte("failed to read response body")
	}

	if resp.StatusCode >= 400 {
		return &HTTPError{
			StatusCode: resp.StatusCode,
			Status:     resp.Status,
			Body:       string(body),
			URL:        url,
		}
	}

	return nil
}

type HTTPError struct {
	StatusCode int
	Status     string
	Body       string
	URL        string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("webhook request to %s failed with status %d: %s", e.URL, e.StatusCode, e.Status)
}

func (e *HTTPError) IsRetryable() bool {
	return IsHTTPStatusRetryable(e.StatusCode)
}
