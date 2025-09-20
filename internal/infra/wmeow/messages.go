package wmeow

import (
	"context"
	"fmt"
	"strings"

	"zpmeow/internal/application/ports"

	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	waTypes "go.mau.fi/whatsmeow/types"
)

// ContactData movido para internal/application/ports/interfaces.go
type ContactData = ports.ContactData

func sendMessageToJID(client *whatsmeow.Client, to string, message *waProto.Message) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func createMediaMessage(client *whatsmeow.Client, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	return uploadMedia(client, data, mediaType)
}

func validateMessageInput(client *whatsmeow.Client, to string) error {
	if client == nil {
		return fmt.Errorf("client cannot be nil")
	}
	if to == "" {
		return fmt.Errorf("recipient cannot be empty")
	}
	return nil
}

func SendTextMessage(client *whatsmeow.Client, to, text string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}

	message := &waProto.Message{
		Conversation: &text,
	}

	return sendMessageToJID(client, to, message)
}

func SendImageMessage(client *whatsmeow.Client, to string, data []byte, caption string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("image data cannot be empty")
	}

	uploaded, err := createMediaMessage(client, data, whatsmeow.MediaImage)
	if err != nil {
		return nil, err
	}

	mimeType := "image/jpeg"
	message := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
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

	return sendMessageToJID(client, to, message)
}

func SendAudioMessage(client *whatsmeow.Client, to string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("audio data cannot be empty")
	}

	uploaded, err := createMediaMessage(client, data, whatsmeow.MediaAudio)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		AudioMessage: &waProto.AudioMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	return sendMessageToJID(client, to, message)
}

func SendVideoMessage(client *whatsmeow.Client, to string, data []byte, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	if err := validateMessageInput(client, to); err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("video data cannot be empty")
	}

	uploaded, err := createMediaMessage(client, data, whatsmeow.MediaVideo)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		VideoMessage: &waProto.VideoMessage{
			Caption:       &caption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	return sendMessageToJID(client, to, message)
}

func SendDocumentMessage(client *whatsmeow.Client, to string, data []byte, filename, caption, mimeType string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	uploaded, err := uploadMedia(client, data, whatsmeow.MediaDocument)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		DocumentMessage: &waProto.DocumentMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			FileName:      &filename,
			Caption:       &caption,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func SendStickerMessage(client *whatsmeow.Client, to string, data []byte, mimeType string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	uploaded, err := uploadMedia(client, data, whatsmeow.MediaImage) // Stickers use image media type
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		StickerMessage: &waProto.StickerMessage{
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			Mimetype:      &mimeType,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func SendContactMessage(client *whatsmeow.Client, to, contactName, contactPhone string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL;type=CELL;type=VOICE;waid=%s:+%s\nEND:VCARD", contactName, contactPhone, contactPhone)

	message := &waProto.Message{
		ContactMessage: &waProto.ContactMessage{
			DisplayName: &contactName,
			Vcard:       &vcard,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func SendContactsMessage(client *whatsmeow.Client, to string, contacts []ContactData) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	if len(contacts) == 0 {
		return nil, fmt.Errorf("at least one contact is required")
	}

	if len(contacts) > 10 {
		return nil, fmt.Errorf("maximum 10 contacts allowed")
	}

	// For single contact, use ContactMessage for better compatibility
	if len(contacts) == 1 {
		return SendContactMessage(client, to, contacts[0].Name, contacts[0].Phone)
	}

	// For multiple contacts, use ContactsArrayMessage to send all contacts in a single message
	var contactMessages []*waProto.ContactMessage
	for _, contact := range contacts {
		vcard := fmt.Sprintf("BEGIN:VCARD\nVERSION:3.0\nFN:%s\nTEL;type=CELL;type=VOICE;waid=%s:+%s\nEND:VCARD",
			contact.Name, contact.Phone, contact.Phone)

		contactMessages = append(contactMessages, &waProto.ContactMessage{
			DisplayName: &contact.Name,
			Vcard:       &vcard,
		})
	}

	displayName := fmt.Sprintf("%d contacts", len(contacts))
	message := &waProto.Message{
		ContactsArrayMessage: &waProto.ContactsArrayMessage{
			DisplayName: &displayName,
			Contacts:    contactMessages,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func SendLocationMessage(client *whatsmeow.Client, to string, latitude, longitude float64, name, address string) (*whatsmeow.SendResponse, error) {
	jid, err := parsePhoneToJID(to)
	if err != nil {
		return nil, err
	}

	message := &waProto.Message{
		LocationMessage: &waProto.LocationMessage{
			DegreesLatitude:  &latitude,
			DegreesLongitude: &longitude,
			Name:             &name,
			Address:          &address,
		},
	}

	resp, err := client.SendMessage(context.Background(), jid, message)
	return &resp, err
}

func parsePhoneToJID(phone string) (waTypes.JID, error) {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return waTypes.EmptyJID, fmt.Errorf("phone number cannot be empty")
	}

	if phone[0] == '+' {
		phone = phone[1:]
	}

	var digits strings.Builder
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	formattedPhone := digits.String()

	if formattedPhone == "" {
		return waTypes.EmptyJID, fmt.Errorf("phone number cannot be empty")
	}

	if len(formattedPhone) < 7 || len(formattedPhone) > 15 {
		return waTypes.EmptyJID, fmt.Errorf("phone number must be between 7 and 15 digits")
	}

	if formattedPhone[0] == '0' {
		return waTypes.EmptyJID, fmt.Errorf("phone number should not start with 0")
	}

	return waTypes.NewJID(formattedPhone, waTypes.DefaultUserServer), nil
}

func uploadMedia(client *whatsmeow.Client, data []byte, mediaType whatsmeow.MediaType) (*whatsmeow.UploadResponse, error) {
	resp, err := client.Upload(context.Background(), data, mediaType)
	return &resp, err
}
