package wmeow

import (
	"fmt"
	"regexp"
	"strings"

	"zpmeow/internal/application/ports"
)

// SessionValidator provides session validation utilities
type SessionValidator struct{}

// NewSessionValidator creates a new session validator
func NewSessionValidator() *SessionValidator {
	return &SessionValidator{}
}

// ValidateSessionID validates session ID format and content
func (v *SessionValidator) ValidateSessionID(sessionID string) error {
	if sessionID == "" {
		return newValidationError("sessionID", "session ID cannot be empty")
	}

	if len(sessionID) < 8 {
		return newValidationError("sessionID", "session ID must be at least 8 characters long")
	}

	// Check for valid UUID format (optional but recommended)
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	_ = uuidRegex.MatchString(sessionID) // Allow non-UUID formats for backward compatibility

	return nil
}

// PhoneValidator provides phone number validation utilities
type PhoneValidator struct{}

// NewPhoneValidator creates a new phone validator
func NewPhoneValidator() *PhoneValidator {
	return &PhoneValidator{}
}

// ValidatePhoneNumber validates phone number format
func (v *PhoneValidator) ValidatePhoneNumber(phone string) error {
	if phone == "" {
		return newValidationError("phone", "phone number cannot be empty")
	}

	// Remove common separators
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "+", "")

	// Check if contains only digits
	digitRegex := regexp.MustCompile(`^\d+$`)
	if !digitRegex.MatchString(cleanPhone) {
		return newValidationError("phone", "phone number must contain only digits and common separators")
	}

	// Check minimum length (international format)
	if len(cleanPhone) < 10 {
		return newValidationError("phone", "phone number must be at least 10 digits long")
	}

	// Check maximum length
	if len(cleanPhone) > 15 {
		return newValidationError("phone", "phone number cannot be longer than 15 digits")
	}

	return nil
}

// ContentValidator provides message content validation utilities
type ContentValidator struct{}

// NewContentValidator creates a new content validator
func NewContentValidator() *ContentValidator {
	return &ContentValidator{}
}

// ValidateTextMessage validates text message content
func (v *ContentValidator) ValidateTextMessage(text string) error {
	if text == "" {
		return newValidationError("text", "message text cannot be empty")
	}

	// Check maximum length (WhatsApp limit is ~4096 characters)
	if len(text) > 4096 {
		return newValidationError("text", "message text cannot exceed 4096 characters")
	}

	return nil
}

// ValidateMediaData validates media data
func (v *ContentValidator) ValidateMediaData(data []byte, maxSize int64) error {
	if len(data) == 0 {
		return newValidationError("media", "media data cannot be empty")
	}

	if int64(len(data)) > maxSize {
		return newValidationError("media", fmt.Sprintf("media size cannot exceed %d bytes", maxSize))
	}

	return nil
}

// ValidateFileName validates file name
func (v *ContentValidator) ValidateFileName(fileName string) error {
	if fileName == "" {
		return newValidationError("fileName", "file name cannot be empty")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(fileName, char) {
			return newValidationError("fileName", fmt.Sprintf("file name cannot contain '%s'", char))
		}
	}

	// Check maximum length
	if len(fileName) > 255 {
		return newValidationError("fileName", "file name cannot exceed 255 characters")
	}

	return nil
}

// ButtonValidator provides button validation utilities
type ButtonValidator struct{}

// NewButtonValidator creates a new button validator
func NewButtonValidator() *ButtonValidator {
	return &ButtonValidator{}
}

// ValidateButtons validates button data
func (v *ButtonValidator) ValidateButtons(buttons []ports.ButtonData) error {
	if len(buttons) == 0 {
		return newValidationError("buttons", "at least one button is required")
	}

	if len(buttons) > 3 {
		return newValidationError("buttons", "maximum 3 buttons allowed")
	}

	for i, button := range buttons {
		if err := v.validateSingleButton(button, i); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleButton validates a single button
func (v *ButtonValidator) validateSingleButton(button ports.ButtonData, index int) error {
	if button.ID == "" {
		return newValidationError("buttons", fmt.Sprintf("button %d: ID cannot be empty", index))
	}

	if button.Text == "" {
		return newValidationError("buttons", fmt.Sprintf("button %d: text cannot be empty", index))
	}

	if len(button.Text) > 20 {
		return newValidationError("buttons", fmt.Sprintf("button %d: text cannot exceed 20 characters", index))
	}

	return nil
}

// ListValidator provides list validation utilities
type ListValidator struct{}

// NewListValidator creates a new list validator
func NewListValidator() *ListValidator {
	return &ListValidator{}
}

// ValidateListSections validates list sections
func (v *ListValidator) ValidateListSections(sections []ports.ListSection) error {
	if len(sections) == 0 {
		return newValidationError("sections", "at least one section is required")
	}

	if len(sections) > 10 {
		return newValidationError("sections", "maximum 10 sections allowed")
	}

	for i, section := range sections {
		if err := v.validateSingleSection(section, i); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleSection validates a single list section
func (v *ListValidator) validateSingleSection(section ports.ListSection, index int) error {
	if section.Title == "" {
		return newValidationError("sections", fmt.Sprintf("section %d: title cannot be empty", index))
	}

	if len(section.Rows) == 0 {
		return newValidationError("sections", fmt.Sprintf("section %d: at least one row is required", index))
	}

	if len(section.Rows) > 10 {
		return newValidationError("sections", fmt.Sprintf("section %d: maximum 10 rows allowed", index))
	}

	for j, row := range section.Rows {
		if err := v.validateSingleRow(row, index, j); err != nil {
			return err
		}
	}

	return nil
}

// validateSingleRow validates a single list row
func (v *ListValidator) validateSingleRow(row ports.ListItem, sectionIndex, rowIndex int) error {
	if row.ID == "" {
		return newValidationError("sections", fmt.Sprintf("section %d, row %d: ID cannot be empty", sectionIndex, rowIndex))
	}

	if row.Title == "" {
		return newValidationError("sections", fmt.Sprintf("section %d, row %d: title cannot be empty", sectionIndex, rowIndex))
	}

	if len(row.Title) > 24 {
		return newValidationError("sections", fmt.Sprintf("section %d, row %d: title cannot exceed 24 characters", sectionIndex, rowIndex))
	}

	if len(row.Description) > 72 {
		return newValidationError("sections", fmt.Sprintf("section %d, row %d: description cannot exceed 72 characters", sectionIndex, rowIndex))
	}

	return nil
}
