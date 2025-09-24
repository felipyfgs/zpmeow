package common

import (
	"fmt"
	"strings"
)

type TimeProvider interface {
	Now() Timestamp
}

type systemTimeProvider struct{}

func (p *systemTimeProvider) Now() Timestamp {
	return Now()
}

var defaultTimeProvider TimeProvider = &systemTimeProvider{}

func SetTimeProvider(provider TimeProvider) {
	defaultTimeProvider = provider
}

func GetCurrentTime() Timestamp {
	return defaultTimeProvider.Now()
}

type URLValidator interface {
	ValidateURL(url string) error
	ValidateScheme(url string, allowedSchemes []string) error
	ExtractScheme(url string) string
	HasHost(url string) bool
}

type basicURLValidator struct{}

func (v *basicURLValidator) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if !strings.Contains(url, "://") {
		return fmt.Errorf("URL must contain scheme (://)")
	}

	return nil
}

func (v *basicURLValidator) ValidateScheme(url string, allowedSchemes []string) error {
	scheme := v.ExtractScheme(url)
	if scheme == "" {
		return fmt.Errorf("URL must have a scheme")
	}

	for _, allowed := range allowedSchemes {
		if strings.EqualFold(scheme, allowed) {
			return nil
		}
	}

	return fmt.Errorf("unsupported scheme: %s", scheme)
}

func (v *basicURLValidator) ExtractScheme(url string) string {
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return ""
	}
	return strings.ToLower(parts[0])
}

func (v *basicURLValidator) HasHost(url string) bool {
	parts := strings.Split(url, "://")
	if len(parts) < 2 {
		return false
	}
	return parts[1] != ""
}

var defaultURLValidator URLValidator = &basicURLValidator{}

func SetURLValidator(validator URLValidator) {
	defaultURLValidator = validator
}

func GetURLValidator() URLValidator {
	return defaultURLValidator
}
