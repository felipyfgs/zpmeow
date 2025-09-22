package dto

import (
	"fmt"
	"net/http"
	"strings"
)

type SendTextRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
	Body  string `json:"body" binding:"required" example:"Hello, World!"`
}

func (r SendTextRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Body) == "" {
		return fmt.Errorf("body is required")
	}
	if len(r.Body) > 4096 {
		return fmt.Errorf("body must not exceed 4096 characters")
	}
	return nil
}

type SendMediaRequest struct {
	Phone     string `json:"phone" binding:"required" example:"5511999999999"`
	MediaType string `json:"media_type" binding:"required" example:"image"`
	MediaURL  string `json:"media_url" binding:"required" example:"data:image/jpeg;base64,/9j/4AAQ..."`
	Caption   string `json:"caption,omitempty" example:"Check this out!"`
}

func (r SendMediaRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.MediaType) == "" {
		return fmt.Errorf("media_type is required")
	}
	if strings.TrimSpace(r.MediaURL) == "" {
		return fmt.Errorf("media_url is required")
	}
	validTypes := []string{"image", "audio", "video", "document", "sticker"}
	for _, validType := range validTypes {
		if r.MediaType == validType {
			return nil
		}
	}
	return fmt.Errorf("invalid media_type, must be one of: %s", strings.Join(validTypes, ", "))
}

type SendImageRequest struct {
	Phone   string `json:"phone" binding:"required" example:"5511999999999"`
	Image   string `json:"image" binding:"required" example:"data:image/jpeg;base64,/9j/4AAQ..."`
	Caption string `json:"caption,omitempty" example:"Check this image!"`
}

func (r SendImageRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Image) == "" {
		return fmt.Errorf("image is required")
	}
	return nil
}

type SendAudioRequest struct {
	Phone string `json:"phone" binding:"required" example:"5511999999999"`
	Audio string `json:"audio" binding:"required" example:"data:audio/mpeg;base64,SUQzBAAAAAAAI1RTU0UAAAAPAAADTGF2ZjU4Ljc2LjEwMAAAAAAAAAAAAAAA"`
	PTT   bool   `json:"ptt,omitempty" example:"false"`
}

func (r SendAudioRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Audio) == "" {
		return fmt.Errorf("audio is required")
	}
	return nil
}

type SendVideoRequest struct {
	Phone       string `json:"phone" binding:"required" example:"5511999999999"`
	Video       string `json:"video" binding:"required" example:"data:video/mp4;base64,AAAAIGZ0eXBpc29tAAACAGlzb21pc28y"`
	Caption     string `json:"caption,omitempty" example:"Check this video!"`
	GifPlayback bool   `json:"gif_playback,omitempty" example:"false"`
}

func (r SendVideoRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Video) == "" {
		return fmt.Errorf("video is required")
	}
	return nil
}

type SendDocumentRequest struct {
	Phone    string `json:"phone" binding:"required" example:"5511999999999"`
	Document string `json:"document" binding:"required" example:"data:application/pdf;base64,JVBERi0xLjQKJcOkw7zDtsO8"`
	FileName string `json:"filename,omitempty" example:"document.pdf"`
	MimeType string `json:"mime_type,omitempty" example:"application/pdf"`
}

func (r SendDocumentRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Document) == "" {
		return fmt.Errorf("document is required")
	}
	return nil
}

type SendStickerRequest struct {
	Phone   string `json:"phone" binding:"required" example:"5511999999999"`
	Sticker string `json:"sticker" binding:"required" example:"data:image/webp;base64,UklGRnoGAABXRUJQ"`
}

func (r SendStickerRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Sticker) == "" {
		return fmt.Errorf("sticker is required")
	}
	return nil
}

type SendLocationRequest struct {
	Phone     string  `json:"phone" binding:"required" example:"5511999999999"`
	Latitude  float64 `json:"latitude" binding:"required" example:"-23.5505"`
	Longitude float64 `json:"longitude" binding:"required" example:"-46.6333"`
	Name      string  `json:"name,omitempty" example:"São Paulo"`
	Address   string  `json:"address,omitempty" example:"São Paulo, SP, Brazil"`
}

func (r SendLocationRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if r.Latitude < -90 || r.Latitude > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if r.Longitude < -180 || r.Longitude > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

type MessageContactData struct {
	Name  string `json:"name" binding:"required" example:"John Doe"`
	Phone string `json:"phone" binding:"required" example:"5511888888888"`
}

type SendContactRequest struct {
	Phone        string               `json:"phone" binding:"required" example:"5511999999999"`
	ContactName  string               `json:"contact_name,omitempty" example:"John Doe"`
	ContactPhone string               `json:"contact_phone,omitempty" example:"5511888888888"`
	Contacts     []MessageContactData `json:"contacts,omitempty"`
}

func (r SendContactRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}

	if r.IsSingleContact() {
		if strings.TrimSpace(r.ContactName) == "" {
			return fmt.Errorf("contact_name is required")
		}
		if strings.TrimSpace(r.ContactPhone) == "" {
			return fmt.Errorf("contact_phone is required")
		}
		return nil
	}

	if r.IsMultipleContacts() {
		if len(r.Contacts) == 0 {
			return fmt.Errorf("at least one contact is required")
		}
		if len(r.Contacts) > 10 {
			return fmt.Errorf("maximum 10 contacts allowed")
		}
		for i, contact := range r.Contacts {
			if strings.TrimSpace(contact.Name) == "" {
				return fmt.Errorf("contact %d name is required", i)
			}
			if strings.TrimSpace(contact.Phone) == "" {
				return fmt.Errorf("contact %d phone is required", i)
			}
		}
		return nil
	}

	return fmt.Errorf("must provide either single contact or multiple contacts")
}

func (r SendContactRequest) IsSingleContact() bool {
	return r.ContactName != "" || r.ContactPhone != ""
}

func (r SendContactRequest) IsMultipleContacts() bool {
	return len(r.Contacts) > 0
}

type MarkAsReadRequest struct {
	Phone      string   `json:"phone" binding:"required" example:"5511999999999"`
	MessageIDs []string `json:"message_ids" binding:"required" example:"msg_123,msg_456"`
}

type ReactToMessageRequest struct {
	Phone     string `json:"phone" binding:"required" example:"5511999999999"`
	MessageID string `json:"message_id" binding:"required" example:"msg_123"`
	Emoji     string `json:"emoji" binding:"required" example:"👍"`
}

type DeleteMessageRequest struct {
	Phone     string `json:"phone" binding:"required" example:"5511999999999"`
	MessageID string `json:"message_id" binding:"required" example:"msg_123"`
}

type EditMessageRequest struct {
	Phone     string `json:"phone" binding:"required" example:"5511999999999"`
	MessageID string `json:"message_id" binding:"required" example:"msg_123"`
	NewText   string `json:"new_text" binding:"required" example:"Updated message text"`
}

type ButtonData struct {
	ID   string `json:"id" binding:"required" example:"btn_1"`
	Text string `json:"text" binding:"required" example:"Click me"`
	Type string `json:"type,omitempty" example:"reply"`
}

type SendButtonMessageRequest struct {
	Phone   string       `json:"phone" binding:"required" example:"5511999999999"`
	Title   string       `json:"title" binding:"required" example:"Choose an option"`
	Buttons []ButtonData `json:"buttons" binding:"required"`
}

func (r SendButtonMessageRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if len(r.Buttons) == 0 {
		return fmt.Errorf("at least one button is required")
	}
	if len(r.Buttons) > 3 {
		return fmt.Errorf("maximum 3 buttons allowed")
	}
	for i, btn := range r.Buttons {
		if strings.TrimSpace(btn.ID) == "" {
			return fmt.Errorf("button %d id is required", i)
		}
		if strings.TrimSpace(btn.Text) == "" {
			return fmt.Errorf("button %d text is required", i)
		}
	}
	return nil
}

type ListRow struct {
	ID          string `json:"id" binding:"required" example:"row_1"`
	Title       string `json:"title" binding:"required" example:"Option 1"`
	Description string `json:"description,omitempty" example:"Description for option 1"`
}

type ListSection struct {
	Title string    `json:"title" binding:"required" example:"Section 1"`
	Rows  []ListRow `json:"rows" binding:"required"`
}

type SendListMessageRequest struct {
	Phone       string        `json:"phone" binding:"required" example:"5511999999999"`
	Title       string        `json:"title" binding:"required" example:"Choose from list"`
	Description string        `json:"description,omitempty" example:"Please select an option"`
	ButtonText  string        `json:"button_text" binding:"required" example:"Select"`
	FooterText  string        `json:"footer_text,omitempty" example:"Footer text"`
	Sections    []ListSection `json:"sections" binding:"required"`
}

func (r SendListMessageRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if strings.TrimSpace(r.ButtonText) == "" {
		return fmt.Errorf("button_text is required")
	}
	if len(r.Sections) == 0 {
		return fmt.Errorf("at least one section is required")
	}
	if len(r.Sections) > 10 {
		return fmt.Errorf("maximum 10 sections allowed")
	}
	for i, section := range r.Sections {
		if strings.TrimSpace(section.Title) == "" {
			return fmt.Errorf("section %d title is required", i)
		}
		if len(section.Rows) == 0 {
			return fmt.Errorf("section %d must have at least one row", i)
		}
		if len(section.Rows) > 10 {
			return fmt.Errorf("section %d can have maximum 10 rows", i)
		}
		for j, row := range section.Rows {
			if strings.TrimSpace(row.ID) == "" {
				return fmt.Errorf("section %d row %d id is required", i, j)
			}
			if strings.TrimSpace(row.Title) == "" {
				return fmt.Errorf("section %d row %d title is required", i, j)
			}
		}
	}
	return nil
}

type SendPollMessageRequest struct {
	Phone           string   `json:"phone" binding:"required" example:"5511999999999"`
	Name            string   `json:"name" binding:"required" example:"What's your favorite color?"`
	Options         []string `json:"options" binding:"required" example:"Red,Blue,Green"`
	SelectableCount int      `json:"selectable_count,omitempty" example:"1"`
}

func (r SendPollMessageRequest) Validate() error {
	if strings.TrimSpace(r.Phone) == "" {
		return fmt.Errorf("phone is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(r.Options) < 2 {
		return fmt.Errorf("at least 2 options are required")
	}
	if len(r.Options) > 12 {
		return fmt.Errorf("maximum 12 options allowed")
	}
	for i, option := range r.Options {
		if strings.TrimSpace(option) == "" {
			return fmt.Errorf("option %d cannot be empty", i)
		}
	}
	if r.SelectableCount <= 0 {
		r.SelectableCount = 1
	}
	if r.SelectableCount > len(r.Options) {
		return fmt.Errorf("selectable_count cannot be greater than number of options")
	}
	return nil
}

type MessageDownloadMediaRequest struct {
	MessageID string `json:"message_id" binding:"required" example:"msg_123"`
}

type MessageErrorResponse struct {
	Code    string `json:"code" example:"INVALID_PHONE"`
	Message string `json:"message" example:"Invalid phone number format"`
	Details string `json:"details" example:"Phone number must include country code"`
}

type MessageKey struct {
	ID        string `json:"id"`
	RemoteJID string `json:"remoteJid"`
	FromMe    bool   `json:"fromMe"`
}

type TextMessagePayload struct {
	Text string `json:"text" example:"Hello, World!"`
}

type ImageMessagePayload struct {
	URL     string `json:"url" example:"https://example.com/image.jpg"`
	Caption string `json:"caption,omitempty" example:"Check this image!"`
}

type AudioMessagePayload struct {
	URL string `json:"url" example:"https://example.com/audio.mp3"`
	PTT bool   `json:"ptt" example:"false"`
}

type VideoMessagePayload struct {
	URL         string `json:"url" example:"https://example.com/video.mp4"`
	Caption     string `json:"caption,omitempty" example:"Check this video!"`
	GifPlayback bool   `json:"gif_playback" example:"false"`
}

type DocumentMessagePayload struct {
	URL      string `json:"url" example:"https://example.com/document.pdf"`
	FileName string `json:"filename" example:"document.pdf"`
	MimeType string `json:"mime_type" example:"application/pdf"`
}

type StickerMessagePayload struct {
	URL string `json:"url" example:"https://example.com/sticker.webp"`
}

type LocationMessagePayload struct {
	Latitude  float64 `json:"latitude" example:"-23.5505"`
	Longitude float64 `json:"longitude" example:"-46.6333"`
	Name      string  `json:"name,omitempty" example:"Sao Paulo"`
	Address   string  `json:"address,omitempty" example:"Sao Paulo, SP, Brazil"`
}

type ContactMessagePayload struct {
	Name  string `json:"name" example:"John Doe"`
	VCard string `json:"vcard" example:"BEGIN:VCARD\\nVERSION:3.0\\nFN:John Doe\\nTEL:5511888888888\\nEND:VCARD"`
}

type ContactsMessagePayload struct {
	VCards []string `json:"vcards"`
}

type MessagePayload struct {
	Text     *TextMessagePayload     `json:"text,omitempty"`
	Image    *ImageMessagePayload    `json:"image,omitempty"`
	Audio    *AudioMessagePayload    `json:"audio,omitempty"`
	Video    *VideoMessagePayload    `json:"video,omitempty"`
	Document *DocumentMessagePayload `json:"document,omitempty"`
	Sticker  *StickerMessagePayload  `json:"sticker,omitempty"`
	Location *LocationMessagePayload `json:"location,omitempty"`
	Contact  *ContactMessagePayload  `json:"contact,omitempty"`
	Contacts *ContactsMessagePayload `json:"contacts,omitempty"`
}

type MessageResponseData struct {
	Key       MessageKey     `json:"key"`
	Message   MessagePayload `json:"message"`
	Timestamp int64          `json:"timestamp"`
}

type MessageResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *MessageResponseData  `json:"data,omitempty"`
	Error   *MessageErrorResponse `json:"error,omitempty"`
}

type MessageActionResponse struct {
	Success bool                  `json:"success"`
	Code    int                   `json:"code"`
	Data    *MessageActionData    `json:"data,omitempty"`
	Error   *MessageErrorResponse `json:"error,omitempty"`
}

type MessageActionData struct {
	Phone     string `json:"phone"`
	MessageID string `json:"message_id,omitempty"`
	Action    string `json:"action"`
}

type MessageMediaDownloadResponse struct {
	Success   bool   `json:"success"`
	Code      int    `json:"code"`
	MessageID string `json:"message_id"`
	MediaType string `json:"media_type"`
	MimeType  string `json:"mime_type"`
	Data      []byte `json:"data"`
	Size      int    `json:"size"`
}

func NewMessageErrorResponse(code int, errorCode, message, details string) *MessageResponse {
	return &MessageResponse{
		Success: false,
		Code:    code,
		Error: &MessageErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewMessageActionErrorResponse(code int, errorCode, message, details string) *MessageActionResponse {
	return &MessageActionResponse{
		Success: false,
		Code:    code,
		Error: &MessageErrorResponse{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	}
}

func NewMessageActionSuccessResponse(phone, messageID, action string) *MessageActionResponse {
	return &MessageActionResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &MessageActionData{
			Phone:     phone,
			MessageID: messageID,
			Action:    action,
		},
	}
}

func NewTextResponse(success bool, code int, phone, messageID, text string, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Text: &TextMessagePayload{
					Text: text,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewImageResponse(success bool, code int, phone, messageID, imageURL, caption string, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Image: &ImageMessagePayload{
					URL:     imageURL,
					Caption: caption,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewAudioResponse(success bool, code int, phone, messageID, audioURL string, ptt, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Audio: &AudioMessagePayload{
					URL: audioURL,
					PTT: ptt,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewVideoResponse(success bool, code int, phone, messageID, videoURL, caption string, gifPlayback, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Video: &VideoMessagePayload{
					URL:         videoURL,
					Caption:     caption,
					GifPlayback: gifPlayback,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewDocumentResponse(success bool, code int, phone, messageID, documentURL, filename, mimeType string, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Document: &DocumentMessagePayload{
					URL:      documentURL,
					FileName: filename,
					MimeType: mimeType,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewStickerResponse(success bool, code int, phone, messageID, stickerURL string, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Sticker: &StickerMessagePayload{
					URL: stickerURL,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewLocationResponse(success bool, code int, phone, messageID string, latitude, longitude float64, name, address string, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Location: &LocationMessagePayload{
					Latitude:  latitude,
					Longitude: longitude,
					Name:      name,
					Address:   address,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewContactResponse(success bool, code int, phone, messageID, contactName, vcard string, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Contact: &ContactMessagePayload{
					Name:  contactName,
					VCard: vcard,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewContactsMessageResponse(success bool, code int, phone, messageID string, vcards []string, sent bool) *MessageResponse {
	return &MessageResponse{
		Success: success,
		Code:    code,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Contacts: &ContactsMessagePayload{
					VCards: vcards,
				},
			},
			Timestamp: 0,
		},
	}
}

func NewMessageSuccessResponse(sessionID, phone, action, messageID string, timestamp int64) *MessageResponse {
	return &MessageResponse{
		Success: true,
		Code:    http.StatusOK,
		Data: &MessageResponseData{
			Key: MessageKey{
				ID:        messageID,
				RemoteJID: phone + "@s.whatsapp.net",
				FromMe:    true,
			},
			Message: MessagePayload{
				Text: &TextMessagePayload{
					Text: action + " completed successfully",
				},
			},
			Timestamp: timestamp,
		},
	}
}
