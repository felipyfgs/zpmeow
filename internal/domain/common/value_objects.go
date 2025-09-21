package common

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type ID struct {
	value string
}

type IDGenerator interface {
	Generate() string
	GenerateAPIKey() string
}

var defaultGenerator IDGenerator = &counterIDGenerator{counter: 0}


type counterIDGenerator struct {
	counter int
}

func (g *counterIDGenerator) Generate() string {
	g.counter++
	return fmt.Sprintf("domain-id-%d", g.counter)
}

func (g *counterIDGenerator) GenerateAPIKey() string {
	g.counter++
	return fmt.Sprintf("api-key-%d", g.counter)
}

func NewID(value string) (ID, error) {
	if value == "" {
		return ID{value: ""}, nil
	}

	if len(value) > 100 {
		return ID{}, fmt.Errorf("invalid ID format: must not exceed 100 characters")
	}

	return ID{value: value}, nil
}

func GenerateID() ID {
	return ID{value: defaultGenerator.Generate()}
}

func (id ID) Value() string {
	return id.value
}

func (id ID) String() string {
	return id.value
}

func (id ID) IsEmpty() bool {
	return id.value == ""
}

func (id ID) Equals(other ID) bool {
	return id.value == other.value
}

type Timestamp struct {
	value time.Time
}

func NewTimestamp(t time.Time) Timestamp {
	return Timestamp{value: t}
}

func Now() Timestamp {
	return Timestamp{value: time.Now()}
}

func (ts Timestamp) Value() time.Time {
	return ts.value
}

func (ts Timestamp) IsZero() bool {
	return ts.value.IsZero()
}

func (ts Timestamp) Before(other Timestamp) bool {
	return ts.value.Before(other.value)
}

func (ts Timestamp) After(other Timestamp) bool {
	return ts.value.After(other.value)
}

type Name struct {
	value string
}

func NewName(value string, minLength, maxLength int) (Name, error) {
	trimmed := strings.TrimSpace(value)

	if trimmed == "" {
		return Name{}, fmt.Errorf("name cannot be empty")
	}

	if len(trimmed) < minLength {
		return Name{}, fmt.Errorf("name must be at least %d characters long", minLength)
	}

	if len(trimmed) > maxLength {
		return Name{}, fmt.Errorf("name cannot exceed %d characters", maxLength)
	}

	nameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !nameRegex.MatchString(trimmed) {
		return Name{}, fmt.Errorf("name can only contain letters, numbers, hyphens, and underscores")
	}

	return Name{value: trimmed}, nil
}

func (n Name) Value() string {
	return n.value
}

func (n Name) String() string {
	return n.value
}

func (n Name) IsEmpty() bool {
	return n.value == ""
}

func (n Name) Equals(other Name) bool {
	return n.value == other.value
}
