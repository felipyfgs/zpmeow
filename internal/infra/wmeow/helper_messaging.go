package wmeow

import (
	"context"
	"fmt"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
)

type mediaUploader struct{}

func NewMediaUploader() *mediaUploader {
	return &mediaUploader{}
}

func (u *mediaUploader) UploadMedia(client *whatsmeow.Client, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	resp, err := client.Upload(context.Background(), data, mediaType)
	return &resp, err
}

type messageSender struct {
	validator *messageValidator
	parser    *phoneParser
	uploader  *mediaUploader
}

func NewMessageSender() *messageSender {
	return &messageSender{
		validator: NewMessageValidator(),
		parser:    NewPhoneParser(),
		uploader:  NewMediaUploader(),
	}
}

func (s *messageSender) SendToJID(client *whatsmeow.Client, to string, message interface{}) (*whatsmeow.SendResponse, error) {
	jid, err := s.parser.ParseToJID(to)
	if err != nil {
		return nil, fmt.Errorf("invalid phone number: %w", err)
	}

	waMessage, ok := message.(*waE2E.Message)
	if !ok {
		return nil, newValidationError("message", "invalid message type")
	}

	resp, err := client.SendMessage(context.Background(), jid, waMessage)
	return &resp, err
}

func (s *messageSender) CreateMediaMessage(client *whatsmeow.Client, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	if err := s.validator.ValidateMediaData(data); err != nil {
		return nil, err
	}

	return s.uploader.UploadMedia(client, data, mediaType)
}

type mimeTypeHelper struct{}

func NewMimeTypeHelper() *mimeTypeHelper {
	return &mimeTypeHelper{}
}

func (h *mimeTypeHelper) GetOptimalAudioMimeType(originalMimeType string, isPTT bool) string {
	// Para mensagens PTT (Push-to-Talk), WhatsApp prefere OGG/Opus
	if isPTT {
		// Se já é OGG, mantém
		if originalMimeType == "audio/ogg" || originalMimeType == "audio/ogg; codecs=opus" {
			return "audio/ogg; codecs=opus"
		}
		// Para outros formatos em PTT, converte para OGG/Opus
		return "audio/ogg; codecs=opus"
	}
	// Para áudio normal (não PTT), mantém o formato original
	return originalMimeType
}

func (h *mimeTypeHelper) GetDefaultImageMimeType() string {
	return "image/jpeg"
}

func (h *mimeTypeHelper) GetDefaultVideoMimeType() string {
	return "video/mp4"
}

func (h *mimeTypeHelper) GetDefaultDocumentMimeType() string {
	return "application/octet-stream"
}

func (h *mimeTypeHelper) GetDefaultStickerMimeType() string {
	return "image/webp"
}

type MessageBuilder struct {
	validator  *messageValidator
	mimeHelper *mimeTypeHelper
}

func NewMessageBuilder() *MessageBuilder {
	return &MessageBuilder{
		validator:  NewMessageValidator(),
		mimeHelper: NewMimeTypeHelper(),
	}
}

func (b *MessageBuilder) BuildTextMessage(text string) (*waE2E.Message, error) {
	if err := b.validator.ValidateTextContent(text); err != nil {
		return nil, err
	}

	return &waE2E.Message{
		Conversation: &text,
	}, nil
}

func (b *MessageBuilder) BuildImageMessage(uploaded *whatsmeow.UploadResponse, caption string) *waE2E.Message {
	mimeType := b.mimeHelper.GetDefaultImageMimeType()
	return &waE2E.Message{
		ImageMessage: &waE2E.ImageMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &mimeType,
		},
	}
}

func (b *MessageBuilder) BuildAudioMessage(uploaded *whatsmeow.UploadResponse, mimeType string, ptt bool) *waE2E.Message {
	finalMimeType := b.mimeHelper.GetOptimalAudioMimeType(mimeType, ptt)

	// Log para debug
	if mimeType != finalMimeType {
		fmt.Printf("🔄 MIME TYPE CONVERTED: %s -> %s (PTT: %v)\n", mimeType, finalMimeType, ptt)
	}

	return &waE2E.Message{
		AudioMessage: &waE2E.AudioMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &finalMimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			PTT:           &ptt,
		},
	}
}

func (b *MessageBuilder) BuildVideoMessage(uploaded *whatsmeow.UploadResponse, caption, mimeType string) *waE2E.Message {
	if mimeType == "" {
		mimeType = b.mimeHelper.GetDefaultVideoMimeType()
	}
	return &waE2E.Message{
		VideoMessage: &waE2E.VideoMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &mimeType,
		},
	}
}

func (b *MessageBuilder) BuildDocumentMessage(uploaded *whatsmeow.UploadResponse, filename, caption, mimeType string) *waE2E.Message {
	if mimeType == "" {
		mimeType = b.mimeHelper.GetDefaultDocumentMimeType()
	}
	return &waE2E.Message{
		DocumentMessage: &waE2E.DocumentMessage{
			Title:         &filename,
			FileName:      &filename,
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &mimeType,
		},
	}
}

func (b *MessageBuilder) BuildStickerMessage(uploaded *whatsmeow.UploadResponse, mimeType string) *waE2E.Message {
	if mimeType == "" {
		mimeType = b.mimeHelper.GetDefaultStickerMimeType()
	}
	return &waE2E.Message{
		StickerMessage: &waE2E.StickerMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      &mimeType,
		},
	}
}

func (b *MessageBuilder) BuildLocationMessage(latitude, longitude float64, name, address string) *waE2E.Message {
	return &waE2E.Message{
		LocationMessage: &waE2E.LocationMessage{
			DegreesLatitude:  &latitude,
			DegreesLongitude: &longitude,
			Name:             &name,
			Address:          &address,
		},
	}
}

func (b *MessageBuilder) BuildContactMessage(name, phone string) *waE2E.Message {
	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL:%s\nEND:VCARD", name, phone)
	return &waE2E.Message{
		ContactMessage: &waE2E.ContactMessage{
			DisplayName: &name,
			Vcard:       &vcard,
		},
	}
}
