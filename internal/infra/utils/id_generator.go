package utils

import (
	"crypto/rand"
	"fmt"
	"time"

	"zpmeow/internal/application/ports"
)

// UUIDGenerator implementa a interface ports.IDGenerator
type UUIDGenerator struct {
	counter int64
}

// NewUUIDGenerator cria uma nova instância do gerador de IDs
func NewUUIDGenerator() ports.IDGenerator {
	return &UUIDGenerator{
		counter: time.Now().UnixNano(),
	}
}

// Generate gera um ID único baseado em timestamp e contador
func (g *UUIDGenerator) Generate() string {
	g.counter++
	return fmt.Sprintf("id_%d_%d", time.Now().UnixNano(), g.counter)
}

// GenerateAPIKey gera uma chave API segura de 32 caracteres
func (g *UUIDGenerator) GenerateAPIKey() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const keyLength = 32

	b := make([]byte, keyLength)
	for i := range b {
		randomByte := make([]byte, 1)
		if _, err := rand.Read(randomByte); err != nil {
			// Fallback para timestamp se crypto/rand falhar
			b[i] = charset[int(time.Now().UnixNano())%len(charset)]
		} else {
			b[i] = charset[randomByte[0]%byte(len(charset))]
		}
	}
	return string(b)
}
