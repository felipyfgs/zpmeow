package webhooks

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"zpmeow/internal/infra/logging"
)

type RetryConfig struct {
	MaxRetries        int
	InitialBackoff    time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

func defaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    1 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

type RetryStrategy struct {
	config *RetryConfig
	logger logging.Logger
}

func NewRetryStrategy(config *RetryConfig) *RetryStrategy {
	if config == nil {
		config = defaultRetryConfig()
	}

	return &RetryStrategy{
		config: config,
		logger: logging.GetLogger().Sub("webhook-retry"),
	}
}

func (r *RetryStrategy) ExecuteWithRetry(ctx context.Context, operation func() error, operationName string) error {
	var lastErr error

	for attempt := 0; attempt <= r.config.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := r.calculateBackoff(attempt)
			r.logger.Infof("Retrying %s in %v (attempt %d/%d)", operationName, backoff, attempt, r.config.MaxRetries)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := operation()
		if err == nil {
			if attempt > 0 {
				r.logger.Infof("%s succeeded after %d retries", operationName, attempt)
			}
			return nil
		}

		lastErr = err
		r.logger.Warnf("%s attempt %d failed: %v", operationName, attempt+1, err)

		if !r.isRetryableError(err) {
			r.logger.Infof("%s failed with non-retryable error: %v", operationName, err)
			return err
		}
	}

	r.logger.Errorf("%s failed after %d attempts: %v", operationName, r.config.MaxRetries+1, lastErr)
	return fmt.Errorf("%s failed after %d attempts: %w", operationName, r.config.MaxRetries+1, lastErr)
}

func (r *RetryStrategy) calculateBackoff(attempt int) time.Duration {
	backoff := time.Duration(float64(r.config.InitialBackoff) *
		math.Pow(r.config.BackoffMultiplier, float64(attempt-1)))

	if backoff > r.config.MaxBackoff {
		backoff = r.config.MaxBackoff
	}

	return backoff
}

func (r *RetryStrategy) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.IsRetryable()
	}

	errStr := strings.ToLower(err.Error())
	retryablePatterns := []string{
		"connection refused",
		"timeout",
		"temporary failure",
		"network unreachable",
		"status 5",
		"context deadline exceeded",
		"no such host",
		"connection reset",
		"i/o timeout",
		"broken pipe",
		"connection timed out",
		"dial tcp",
		"read: connection reset by peer",
	}

	for _, pattern := range retryablePatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}

func IsHTTPStatusRetryable(statusCode int) bool {
	switch statusCode {
	case http.StatusRequestTimeout,
		http.StatusTooManyRequests,
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}
