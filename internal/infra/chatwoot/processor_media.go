package chatwoot

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"zpmeow/internal/application/ports"
)

// MediaProcessor processa múltiplas mídias de forma assíncrona
type MediaProcessor struct {
	whatsappService ports.WhatsAppService
	logger          *slog.Logger
	sessionID       string
	maxConcurrent   int
	timeout         time.Duration
	rateLimiter     *MediaRateLimiter
}

// NewMediaProcessor cria um novo processador de mídia
func NewMediaProcessor(whatsappService ports.WhatsAppService, logger *slog.Logger, sessionID string) *MediaProcessor {
	return &MediaProcessor{
		whatsappService: whatsappService,
		logger:          logger,
		sessionID:       sessionID,
		maxConcurrent:   3, // Máximo 3 mídias simultâneas
		timeout:         60 * time.Second,
		rateLimiter:     NewMediaRateLimiter(logger),
	}
}

// MediaItem representa um item de mídia para processamento
type MediaItem struct {
	ID       string
	URL      string
	FileName string
	MimeType string
	FileSize int64
}

// ProcessMultipleMedia processa múltiplas mídias de forma assíncrona
func (mp *MediaProcessor) ProcessMultipleMedia(ctx context.Context, phoneNumber string, mediaItems []MediaItem) error {
	if len(mediaItems) == 0 {
		return nil
	}

	mp.logger.Info("Processing multiple media items",
		"count", len(mediaItems),
		"phone", phoneNumber,
		"max_concurrent", mp.maxConcurrent)

	// Se há apenas 1 mídia, processa diretamente
	if len(mediaItems) == 1 {
		return mp.processSingleMedia(ctx, phoneNumber, mediaItems[0])
	}

	// Para múltiplas mídias, usa processamento em lotes
	return mp.processBatchMedia(ctx, phoneNumber, mediaItems)
}

// processSingleMedia processa uma única mídia
func (mp *MediaProcessor) processSingleMedia(ctx context.Context, phoneNumber string, item MediaItem) error {
	mp.logger.Info("Processing single media",
		"phone", phoneNumber,
		"file", item.FileName,
		"size", item.FileSize)

	// Cria contexto com timeout
	ctx, cancel := context.WithTimeout(ctx, mp.timeout)
	defer cancel()

	// Baixa os dados da URL do Chatwoot
	mediaData, err := mp.downloadMediaFromURL(ctx, item.URL)
	if err != nil {
		mp.logger.Error("Failed to download single media from URL",
			"error", err,
			"url", item.URL,
			"file", item.FileName)
		return fmt.Errorf("failed to download media: %w", err)
	}

	if len(mediaData) == 0 {
		mp.logger.Error("Downloaded single media data is empty",
			"url", item.URL,
			"file", item.FileName)
		return fmt.Errorf("media_data: cannot be empty")
	}

	mp.logger.Info("Successfully downloaded single media data",
		"size", len(mediaData),
		"file", item.FileName,
		"url", item.URL)

	// Prepara a mensagem de mídia com dados reais
	mediaMsg := ports.MediaMessage{
		Type:     mp.getMediaType(item.MimeType),
		Data:     mediaData, // Dados baixados do Chatwoot
		MimeType: item.MimeType,
		Caption:  "",
		Filename: item.FileName,
	}

	// Envia a mídia
	_, err = mp.whatsappService.SendMediaMessage(ctx, mp.sessionID, phoneNumber, mediaMsg)
	if err != nil {
		mp.logger.Error("Failed to send single media",
			"error", err,
			"phone", phoneNumber,
			"file", item.FileName)
		return fmt.Errorf("failed to send media %s: %w", item.FileName, err)
	}

	mp.logger.Info("Successfully sent single media",
		"phone", phoneNumber,
		"file", item.FileName)

	return nil
}

// processBatchMedia processa múltiplas mídias em lotes
func (mp *MediaProcessor) processBatchMedia(ctx context.Context, phoneNumber string, mediaItems []MediaItem) error {
	mp.logger.Info("Processing batch media",
		"phone", phoneNumber,
		"total_items", len(mediaItems))

	// Canal para controlar concorrência
	semaphore := make(chan struct{}, mp.maxConcurrent)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	// Processa cada mídia em goroutine separada
	for i, item := range mediaItems {
		wg.Add(1)
		go func(index int, mediaItem MediaItem) {
			defer wg.Done()

			// Adquire semáforo
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Adiciona delay entre envios para evitar rate limiting
			if index > 0 {
				delay := time.Duration(index) * 2 * time.Second
				mp.logger.Info("Adding delay before processing media",
					"index", index,
					"delay", delay,
					"file", mediaItem.FileName)
				time.Sleep(delay)
			}

			// Cria contexto com timeout para esta mídia
			mediaCtx, cancel := context.WithTimeout(ctx, mp.timeout)
			defer cancel()

			mp.logger.Info("Processing media item",
				"index", index,
				"phone", phoneNumber,
				"file", mediaItem.FileName,
				"size", mediaItem.FileSize)

			// Usa rate limiter para controlar envio
			err := mp.rateLimiter.ProcessWithLimiting(mediaCtx, func() error {
				// Baixa os dados da URL do Chatwoot
				mediaData, downloadErr := mp.downloadMediaFromURL(mediaCtx, mediaItem.URL)
				if downloadErr != nil {
					mp.logger.Error("Failed to download media from URL",
						"error", downloadErr,
						"url", mediaItem.URL,
						"file", mediaItem.FileName)
					return fmt.Errorf("failed to download media: %w", downloadErr)
				}

				if len(mediaData) == 0 {
					mp.logger.Error("Downloaded media data is empty",
						"url", mediaItem.URL,
						"file", mediaItem.FileName)
					return fmt.Errorf("media_data: cannot be empty")
				}

				mp.logger.Info("Successfully downloaded media data",
					"size", len(mediaData),
					"file", mediaItem.FileName,
					"url", mediaItem.URL)

				// Prepara a mensagem de mídia com dados reais
				mediaMsg := ports.MediaMessage{
					Type:     mp.getMediaType(mediaItem.MimeType),
					Data:     mediaData, // Dados baixados do Chatwoot
					MimeType: mediaItem.MimeType,
					Caption:  "",
					Filename: mediaItem.FileName,
				}
				_, sendErr := mp.whatsappService.SendMediaMessage(mediaCtx, mp.sessionID, phoneNumber, mediaMsg)
				return sendErr
			})

			if err != nil {
				mp.logger.Error("Failed to send media in batch",
					"error", err,
					"index", index,
					"phone", phoneNumber,
					"file", mediaItem.FileName)

				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to send media %s (index %d): %w", mediaItem.FileName, index, err))
				mu.Unlock()
				return
			}

			mp.logger.Info("Successfully sent media in batch",
				"index", index,
				"phone", phoneNumber,
				"file", mediaItem.FileName)

		}(i, item)
	}

	// Aguarda todas as goroutines terminarem
	wg.Wait()

	// Verifica se houve erros
	if len(errors) > 0 {
		mp.logger.Error("Some media items failed to send",
			"total_errors", len(errors),
			"total_items", len(mediaItems),
			"phone", phoneNumber)

		// Retorna apenas o primeiro erro para não sobrecarregar os logs
		return errors[0]
	}

	mp.logger.Info("Successfully processed all media items",
		"total_items", len(mediaItems),
		"phone", phoneNumber)

	return nil
}

// ExtractMediaItems extrai itens de mídia dos anexos do Chatwoot
func (mp *MediaProcessor) ExtractMediaItems(attachments []interface{}) []MediaItem {
	var mediaItems []MediaItem

	for i, attachment := range attachments {
		if attachmentMap, ok := attachment.(map[string]interface{}); ok {
			item := MediaItem{
				ID: fmt.Sprintf("attachment_%d", i),
			}

			// Extrai URL
			if dataURL, ok := attachmentMap["data_url"].(string); ok {
				item.URL = dataURL
			}

			// Extrai nome do arquivo
			if fileName := mp.extractFileName(attachmentMap); fileName != "" {
				item.FileName = fileName
			} else {
				item.FileName = fmt.Sprintf("attachment_%d", i)
			}

			// Extrai tipo MIME
			if fileType, ok := attachmentMap["file_type"].(string); ok {
				item.MimeType = mp.mapFileTypeToMimeType(fileType)
			}

			// Extrai tamanho do arquivo
			if fileSize, ok := attachmentMap["file_size"].(float64); ok {
				item.FileSize = int64(fileSize)
			}

			if item.URL != "" {
				mediaItems = append(mediaItems, item)
			}
		}
	}

	mp.logger.Info("Extracted media items from attachments",
		"total_attachments", len(attachments),
		"valid_media_items", len(mediaItems))

	return mediaItems
}

// extractFileName extrai o nome do arquivo do anexo
func (mp *MediaProcessor) extractFileName(attachmentMap map[string]interface{}) string {
	// Tenta extrair do data_url
	if dataURL, ok := attachmentMap["data_url"].(string); ok {
		detector := NewFileTypeDetector()
		return detector.extractFileName(dataURL)
	}
	return ""
}

// mapFileTypeToMimeType mapeia tipo de arquivo para MIME type
func (mp *MediaProcessor) mapFileTypeToMimeType(fileType string) string {
	switch fileType {
	case "image":
		return "image/jpeg"
	case "file":
		return "application/octet-stream"
	case "audio":
		return "audio/ogg"
	case "video":
		return "video/mp4"
	default:
		return "application/octet-stream"
	}
}

// getMediaType determina o tipo de mídia baseado no MIME type
func (mp *MediaProcessor) getMediaType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	default:
		return "document"
	}
}

// GetProcessingStats retorna estatísticas do processamento
func (mp *MediaProcessor) GetProcessingStats() map[string]interface{} {
	return map[string]interface{}{
		"max_concurrent": mp.maxConcurrent,
		"timeout":        mp.timeout.String(),
		"session_id":     mp.sessionID,
	}
}

// SetMaxConcurrent define o número máximo de processamentos simultâneos
func (mp *MediaProcessor) SetMaxConcurrent(max int) {
	if max > 0 && max <= 10 {
		mp.maxConcurrent = max
		mp.logger.Info("Updated max concurrent media processing", "max_concurrent", max)
	}
}

// SetTimeout define o timeout para processamento de mídia
func (mp *MediaProcessor) SetTimeout(timeout time.Duration) {
	if timeout > 0 && timeout <= 5*time.Minute {
		mp.timeout = timeout
		mp.logger.Info("Updated media processing timeout", "timeout", timeout.String())
	}
}

// downloadMediaFromURL baixa dados de mídia de uma URL do Chatwoot
func (mp *MediaProcessor) downloadMediaFromURL(ctx context.Context, dataURL string) ([]byte, error) {
	if dataURL == "" {
		return nil, fmt.Errorf("data URL is empty")
	}

	mp.logger.Info("Downloading media from Chatwoot URL", "url", dataURL)

	// Cria requisição com contexto
	req, err := http.NewRequestWithContext(ctx, "GET", dataURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Adiciona headers necessários para Chatwoot
	req.Header.Set("User-Agent", "zpmeow/1.0")
	req.Header.Set("Accept", "*/*")

	// Se for URL do Chatwoot local, adiciona headers específicos
	if strings.Contains(dataURL, "localhost:3001") || strings.Contains(dataURL, "127.0.0.1:3001") {
		req.Header.Set("X-Forwarded-For", "127.0.0.1")
		req.Header.Set("X-Real-IP", "127.0.0.1")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Permite até 10 redirecionamentos (padrão do Go)
			if len(via) >= 10 {
				return fmt.Errorf("stopped after 10 redirects")
			}
			// Copia headers importantes para requisições redirecionadas
			req.Header.Set("User-Agent", "zpmeow/1.0")
			req.Header.Set("Accept", "*/*")
			return nil
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download from URL: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			mp.logger.Error("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download from URL: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	mp.logger.Info("Successfully downloaded media from URL",
		"url", dataURL,
		"size", len(data),
		"status", resp.StatusCode)

	return data, nil
}
