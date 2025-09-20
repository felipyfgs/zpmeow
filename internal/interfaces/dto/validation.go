package dto

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

// Validador comum para evitar duplicação de lógica de validação

// ValidatePhone valida se um número de telefone está no formato correto
func ValidatePhone(phone string) error {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return errors.New("phone number is required")
	}

	// Regex para validar formato de telefone brasileiro com código do país
	phoneRegex := regexp.MustCompile(`^55\d{10,11}$`)
	if !phoneRegex.MatchString(phone) {
		return errors.New("phone number must be in format 55XXXXXXXXXX (Brazilian format with country code)")
	}

	return nil
}

// ValidateRequiredString valida se uma string obrigatória não está vazia
func ValidateRequiredString(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

// ValidateStringLength valida o comprimento de uma string
func ValidateStringLength(value, fieldName string, minLength, maxLength int) error {
	value = strings.TrimSpace(value)
	if len(value) < minLength {
		return errors.New(fieldName + " must be at least " + string(rune(minLength)) + " characters")
	}
	if maxLength > 0 && len(value) > maxLength {
		return errors.New(fieldName + " must not exceed " + string(rune(maxLength)) + " characters")
	}
	return nil
}

// ValidateURL valida se uma URL está no formato correto
func ValidateURL(urlStr, fieldName string) error {
	if urlStr == "" {
		return nil // URL opcional
	}

	_, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return errors.New(fieldName + " must be a valid URL")
	}

	return nil
}

// ValidateLatitude valida se uma latitude está no range correto
func ValidateLatitude(lat float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	return nil
}

// ValidateLongitude valida se uma longitude está no range correto
func ValidateLongitude(lng float64) error {
	if lng < -180 || lng > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}

// ValidateArrayLength valida o comprimento de um array
func ValidateArrayLength(arr interface{}, fieldName string, minLength, maxLength int) error {
	var length int

	switch v := arr.(type) {
	case []string:
		length = len(v)
	case []interface{}:
		length = len(v)
	default:
		return errors.New("unsupported array type for validation")
	}

	if length < minLength {
		return errors.New(fieldName + " must have at least " + string(rune(minLength)) + " items")
	}
	if maxLength > 0 && length > maxLength {
		return errors.New(fieldName + " must not exceed " + string(rune(maxLength)) + " items")
	}

	return nil
}

// ValidatePhoneArray valida um array de números de telefone
func ValidatePhoneArray(phones []string, fieldName string) error {
	if len(phones) == 0 {
		return errors.New(fieldName + " must contain at least one phone number")
	}

	for i, phone := range phones {
		if err := ValidatePhone(phone); err != nil {
			return errors.New(fieldName + " at index " + string(rune(i)) + ": " + err.Error())
		}
	}

	return nil
}

// ValidateContactData valida dados de contato
func ValidateContactData(name, phone string) error {
	if err := ValidateRequiredString(name, "contact name"); err != nil {
		return err
	}

	if err := ValidateStringLength(name, "contact name", 1, 100); err != nil {
		return err
	}

	if err := ValidatePhone(phone); err != nil {
		return err
	}

	return nil
}

// ValidateMediaData valida dados de mídia (base64 ou URL)
func ValidateMediaData(data, fieldName string) error {
	data = strings.TrimSpace(data)
	if data == "" {
		return errors.New(fieldName + " is required")
	}

	// Verifica se é uma URL
	if strings.HasPrefix(data, "http://") || strings.HasPrefix(data, "https://") {
		return ValidateURL(data, fieldName)
	}

	// Verifica se é um data URL (base64)
	if strings.HasPrefix(data, "data:") {
		// Validação básica de data URL
		if !strings.Contains(data, ";base64,") {
			return errors.New(fieldName + " must be a valid data URL with base64 encoding")
		}
		return nil
	}

	return errors.New(fieldName + " must be either a valid URL or base64 data URL")
}

// Validator interface para DTOs que implementam validação
type Validator interface {
	Validate() error
}

// ValidateDTO valida um DTO que implementa a interface Validator
func ValidateDTO(dto interface{}) error {
	if validator, ok := dto.(Validator); ok {
		return validator.Validate()
	}
	return nil
}
