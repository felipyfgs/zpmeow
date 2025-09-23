package chatwoot

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClientConfig holds HTTP client configuration
type HTTPClientConfig struct {
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

// DefaultHTTPClientConfig returns default HTTP client configuration
func DefaultHTTPClientConfig() *HTTPClientConfig {
	return &HTTPClientConfig{
		Timeout:         time.Duration(DefaultHTTPTimeout) * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}
}

// HTTPClientBuilder provides utilities for building HTTP clients
type HTTPClientBuilder struct {
	config *HTTPClientConfig
}

// NewHTTPClientBuilder creates a new HTTP client builder
func NewHTTPClientBuilder() *HTTPClientBuilder {
	return &HTTPClientBuilder{
		config: DefaultHTTPClientConfig(),
	}
}

// WithTimeout sets the HTTP timeout
func (b *HTTPClientBuilder) WithTimeout(timeout time.Duration) *HTTPClientBuilder {
	b.config.Timeout = timeout
	return b
}

// WithMaxIdleConns sets the maximum idle connections
func (b *HTTPClientBuilder) WithMaxIdleConns(maxIdleConns int) *HTTPClientBuilder {
	b.config.MaxIdleConns = maxIdleConns
	return b
}

// WithIdleConnTimeout sets the idle connection timeout
func (b *HTTPClientBuilder) WithIdleConnTimeout(timeout time.Duration) *HTTPClientBuilder {
	b.config.IdleConnTimeout = timeout
	return b
}

// Build creates the HTTP client with configured settings
func (b *HTTPClientBuilder) Build() *http.Client {
	transport := &http.Transport{
		MaxIdleConns:    b.config.MaxIdleConns,
		IdleConnTimeout: b.config.IdleConnTimeout,
	}

	return &http.Client{
		Timeout:   b.config.Timeout,
		Transport: transport,
	}
}

// RequestExecutor provides utilities for HTTP requests
type RequestExecutor struct {
	client *http.Client
}

// NewRequestExecutor creates a new request executor
func NewRequestExecutor(client *http.Client) *RequestExecutor {
	return &RequestExecutor{
		client: client,
	}
}

// Execute performs HTTP request with context
func (r *RequestExecutor) Execute(ctx context.Context, req *http.Request) (*http.Response, error) {
	req = req.WithContext(ctx)
	return r.client.Do(req)
}

// ExecuteWithTimeout performs HTTP request with timeout
func (r *RequestExecutor) ExecuteWithTimeout(ctx context.Context, req *http.Request, timeout time.Duration) (*http.Response, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req = req.WithContext(timeoutCtx)
	return r.client.Do(req)
}

// FileDownloaderImpl implements FileDownloader interface
type FileDownloaderImpl struct {
	client *http.Client
}

// NewFileDownloader creates a new file downloader
func NewFileDownloader(client *http.Client) *FileDownloaderImpl {
	return &FileDownloaderImpl{
		client: client,
	}
}

// Download downloads a file from the given URL
func (f *FileDownloaderImpl) Download(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	return resp.Body, nil
}

// URLBuilder provides utilities for building URLs
type URLBuilder struct {
	baseURL   string
	accountID string
}

// NewURLBuilder creates a new URL builder
func NewURLBuilder(baseURL, accountID string) *URLBuilder {
	return &URLBuilder{
		baseURL:   baseURL,
		accountID: accountID,
	}
}

// InboxesURL builds URL for inboxes endpoint
func (b *URLBuilder) InboxesURL() string {
	return b.baseURL + fmt.Sprintf(EndpointInboxes, b.accountID)
}

// ContactsURL builds URL for contacts endpoint
func (b *URLBuilder) ContactsURL() string {
	return b.baseURL + fmt.Sprintf(EndpointContacts, b.accountID)
}

// ConversationsURL builds URL for conversations endpoint
func (b *URLBuilder) ConversationsURL() string {
	return b.baseURL + fmt.Sprintf(EndpointConversations, b.accountID)
}

// MessagesURL builds URL for messages endpoint
func (b *URLBuilder) MessagesURL(conversationID int) string {
	return b.baseURL + fmt.Sprintf(EndpointMessages, b.accountID, conversationID)
}

// ResponseValidator provides utilities for validating HTTP responses
type ResponseValidator struct{}

// NewResponseValidator creates a new response validator
func NewResponseValidator() *ResponseValidator {
	return &ResponseValidator{}
}

// IsSuccess checks if HTTP status code indicates success
func (v *ResponseValidator) IsSuccess(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}

// IsClientError checks if HTTP status code indicates client error
func (v *ResponseValidator) IsClientError(statusCode int) bool {
	return statusCode >= 400 && statusCode < 500
}

// IsServerError checks if HTTP status code indicates server error
func (v *ResponseValidator) IsServerError(statusCode int) bool {
	return statusCode >= 500 && statusCode < 600
}

// ValidateResponse validates HTTP response and returns error if not successful
func (v *ResponseValidator) ValidateResponse(resp *http.Response) error {
	if v.IsSuccess(resp.StatusCode) {
		return nil
	}

	if v.IsClientError(resp.StatusCode) {
		return fmt.Errorf("client error: status %d", resp.StatusCode)
	}

	if v.IsServerError(resp.StatusCode) {
		return fmt.Errorf("server error: status %d", resp.StatusCode)
	}

	return fmt.Errorf("unexpected status: %d", resp.StatusCode)
}
