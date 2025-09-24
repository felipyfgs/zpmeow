package chatwoot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// RateLimiter controla a taxa de envio de mensagens
type RateLimiter struct {
	maxRequests int           // Máximo de requests por janela
	window      time.Duration // Janela de tempo
	requests    []time.Time   // Timestamps dos requests
	mutex       sync.Mutex
	logger      *slog.Logger
}

// NewRateLimiter cria um novo rate limiter
func NewRateLimiter(maxRequests int, window time.Duration, logger *slog.Logger) *RateLimiter {
	return &RateLimiter{
		maxRequests: maxRequests,
		window:      window,
		requests:    make([]time.Time, 0),
		logger:      logger,
	}
}

// Allow verifica se uma request pode ser processada
func (rl *RateLimiter) Allow() bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()

	// Remove requests antigas (fora da janela)
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0)
	for _, req := range rl.requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}
	rl.requests = validRequests

	// Verifica se pode adicionar nova request
	if len(rl.requests) >= rl.maxRequests {
		rl.logger.Warn("Rate limit exceeded",
			"current_requests", len(rl.requests),
			"max_requests", rl.maxRequests,
			"window", rl.window)
		return false
	}

	// Adiciona nova request
	rl.requests = append(rl.requests, now)
	return true
}

// Wait aguarda até que uma request possa ser processada
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		// Calcula tempo de espera
		rl.mutex.Lock()
		if len(rl.requests) > 0 {
			oldestRequest := rl.requests[0]
			waitTime := rl.window - time.Since(oldestRequest)
			rl.mutex.Unlock()

			if waitTime > 0 {
				rl.logger.Info("Rate limit hit, waiting",
					"wait_time", waitTime,
					"current_requests", len(rl.requests))

				select {
				case <-time.After(waitTime):
					continue
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		} else {
			rl.mutex.Unlock()
		}

		// Pequena pausa antes de tentar novamente
		select {
		case <-time.After(100 * time.Millisecond):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// GetStats retorna estatísticas do rate limiter
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	return map[string]interface{}{
		"max_requests":     rl.maxRequests,
		"window":           rl.window.String(),
		"current_requests": len(rl.requests),
		"available_slots":  rl.maxRequests - len(rl.requests),
	}
}

// CircuitBreaker implementa padrão circuit breaker para falhas
type CircuitBreaker struct {
	maxFailures  int
	resetTimeout time.Duration
	failures     int
	lastFailTime time.Time
	state        CircuitState
	mutex        sync.Mutex
	logger       *slog.Logger
}

// CircuitState representa o estado do circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker cria um novo circuit breaker
func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration, logger *slog.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
		logger:       logger,
	}
}

// Call executa uma função com circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Verifica estado atual
	switch cb.state {
	case StateOpen:
		// Verifica se pode tentar reset
		if time.Since(cb.lastFailTime) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.logger.Info("Circuit breaker transitioning to half-open")
		} else {
			return fmt.Errorf("circuit breaker is open")
		}
	case StateHalfOpen:
		// No estado half-open, permite uma tentativa
	case StateClosed:
		// Estado normal, permite execução
	}

	// Executa a função
	err := fn()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// onFailure registra uma falha
func (cb *CircuitBreaker) onFailure() {
	cb.failures++
	cb.lastFailTime = time.Now()

	if cb.failures >= cb.maxFailures {
		cb.state = StateOpen
		cb.logger.Warn("Circuit breaker opened due to failures",
			"failures", cb.failures,
			"max_failures", cb.maxFailures)
	}
}

// onSuccess registra um sucesso
func (cb *CircuitBreaker) onSuccess() {
	cb.failures = 0
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.logger.Info("Circuit breaker closed after successful call")
	}
}

// GetState retorna o estado atual
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	return cb.state
}

// GetStats retorna estatísticas do circuit breaker
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	stateNames := map[CircuitState]string{
		StateClosed:   "closed",
		StateOpen:     "open",
		StateHalfOpen: "half-open",
	}

	return map[string]interface{}{
		"state":         stateNames[cb.state],
		"failures":      cb.failures,
		"max_failures":  cb.maxFailures,
		"reset_timeout": cb.resetTimeout.String(),
		"last_fail":     cb.lastFailTime.Format(time.RFC3339),
	}
}

// Reset força o reset do circuit breaker
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failures = 0
	cb.state = StateClosed
	cb.logger.Info("Circuit breaker manually reset")
}

// MediaRateLimiter combina rate limiting e circuit breaker para mídia
type MediaRateLimiter struct {
	rateLimiter    *RateLimiter
	circuitBreaker *CircuitBreaker
	logger         *slog.Logger
}

// NewMediaRateLimiter cria um rate limiter específico para mídia
func NewMediaRateLimiter(logger *slog.Logger) *MediaRateLimiter {
	return &MediaRateLimiter{
		rateLimiter:    NewRateLimiter(10, 1*time.Minute, logger),   // 10 mídias por minuto
		circuitBreaker: NewCircuitBreaker(5, 2*time.Minute, logger), // 5 falhas, reset em 2min
		logger:         logger,
	}
}

// ProcessWithLimiting processa uma função com rate limiting e circuit breaker
func (mrl *MediaRateLimiter) ProcessWithLimiting(ctx context.Context, fn func() error) error {
	// Primeiro verifica rate limiting
	if err := mrl.rateLimiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// Depois executa com circuit breaker
	return mrl.circuitBreaker.Call(fn)
}

// GetStats retorna estatísticas combinadas
func (mrl *MediaRateLimiter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"rate_limiter":    mrl.rateLimiter.GetStats(),
		"circuit_breaker": mrl.circuitBreaker.GetStats(),
	}
}
