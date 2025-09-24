# 🐱 zpmeow API Documentation

[![API Status](https://img.shields.io/badge/API-90%25%20Complete-brightgreen?style=flat-square)](README.md)
[![Framework](https://img.shields.io/badge/Framework-Fiber-00ADD8?style=flat-square)](https://gofiber.io/)
[![WhatsApp](https://img.shields.io/badge/WhatsApp-Business%20Ready-25D366?style=flat-square&logo=whatsapp)](https://whatsapp.com/)

> **API REST completa para WhatsApp Business construída com Go Fiber e whatsmeow - 90% dos métodos implementados**

## 🚀 **Status da API**

**✅ Métodos Funcionando**: 50+ métodos (90%)
**❌ Não Implementados**: 5 métodos (10%)
**🧪 Taxa de Sucesso**: 90% (funcionalidades core implementadas)

### 🧪 **Funcionalidades Implementadas - Setembro 2025**

**✅ Mensagens (16/18 endpoints)**

- ✅ **SendText** - Envio de mensagens de texto
- ✅ **SendImage** - Envio de imagens (URL/Base64)
- ✅ **SendVideo** - Envio de vídeos (URL/Base64)
- ✅ **SendAudio** - Envio de áudios (URL/Base64)
- ✅ **SendDocument** - Envio de documentos (URL/Base64)
- ✅ **SendSticker** - Envio de stickers WebP
- ✅ **SendContact** - Envio de cartões de contato
- ✅ **SendLocation** - Envio de coordenadas
- ✅ **SendMedia** - Endpoint genérico de mídia
- ✅ **SendPoll** - Criação de enquetes
- ✅ **ReactToMessage** - Reações a mensagens
- ✅ **EditMessage** - Edição de mensagens
- ✅ **DeleteMessage** - Exclusão de mensagens
- ✅ **MarkAsRead** - Marcar mensagens como lidas
- ✅ **SendButton** - Botões interativos (implementado)
- ✅ **SendList** - Listas interativas (implementado)

**✅ Sessões (12/12 endpoints)**

- ✅ **CreateSession** - Criar nova sessão
- ✅ **GetSessions** - Listar todas as sessões
- ✅ **GetSession** - Obter informações da sessão
- ✅ **DeleteSession** - Deletar sessão
- ✅ **ConnectSession** - Conectar via QR Code
- ✅ **DisconnectSession** - Desconectar sessão
- ✅ **PairPhone** - Pareamento via código
- ✅ **GetSessionStatus** - Status da conexão
- ✅ **UpdateSessionWebhook** - Configurar webhooks

**✅ Newsletters (15/15 endpoints)**

- ✅ **CreateNewsletter** - Criar newsletters
- ✅ **GetNewsletter** - Obter informações
- ✅ **ListNewsletters** - Listar newsletters
- ✅ **SubscribeToNewsletter** - Inscrever-se
- ✅ **UnsubscribeFromNewsletter** - Cancelar inscrição
- ✅ **SendNewsletterMessage** - Enviar mensagens
- ✅ **GetNewsletterMessages** - Obter mensagens
- ✅ **ToggleNewsletterMute** - Silenciar/dessilenciar
- ✅ **SendNewsletterReaction** - Reações
- ✅ **MarkNewsletterViewed** - Marcar como visualizado
- ✅ **UploadNewsletterMedia** - Upload de mídia
- ✅ **GetNewsletterByInvite** - Obter por convite
- ✅ **SubscribeLiveUpdates** - Atualizações ao vivo
- ✅ **GetNewsletterMessageUpdates** - Atualizações de mensagens

## 🏗️ **Tecnologias Utilizadas**

- **Framework**: Go Fiber v2.52.9 (alta performance)
- **WhatsApp**: whatsmeow (biblioteca oficial)
- **Database**: PostgreSQL com cache Redis
- **Storage**: MinIO para arquivos de mídia
- **Documentação**: Swagger/OpenAPI integrado
- **Arquitetura**: Clean Architecture + DDD

## 📱 zpmeow API Endpoints

### 🔐 Authentication

All endpoints require authentication via the `Authorization` header:

```
Authorization: Bearer your-super-secret-global-api-key-here
```

**Base URL**: `http://localhost:8080` (desenvolvimento)
**Content-Type**: `application/json`

---

## 🔧 Session Management

### 📋 Create Session

**POST** `/sessions/create`

Create a new WhatsApp session.

**Request Body:**

```json
{
  "session_id": "my-session",
  "name": "Production Session"
}
```

**Response:**

```json
{
  "success": true,
  "code": 201,
  "data": {
    "session_id": "my-session",
    "name": "Production Session",
    "status": "created",
    "qr_code": "data:image/png;base64,iVBORw0KGgo...",
    "timestamp": "2025-09-22T10:30:00Z"
  }
}
```

### 📱 Connect Session

**POST** `/sessions/{sessionId}/connect`

Connect session and get QR code for WhatsApp pairing.

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "session_id": "my-session",
    "status": "connecting",
    "qr_code": "data:image/png;base64,iVBORw0KGgo...",
    "action": "connect"
  }
}
```

### 📊 Session Status

**GET** `/sessions/{sessionId}/status`

Get current session connection status.

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "session_id": "my-session",
    "status": "connected",
    "device_jid": "5511999999999:84@s.whatsapp.net",
    "connected_at": "2025-09-22T10:35:00Z"
  }
}
```

---

## 📨 Message Endpoints

### 🔥 **Advanced Message Actions**

#### 👍 React to Message

**POST** `/session/{sessionId}/message/react`

React to a message with an emoji.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "message_id": "3EB0D098B5FD4BF3BC4327",
  "emoji": "👍"
}
```

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "phone": "5511999999999",
    "message_id": "3EB0D098B5FD4BF3BC4327",
    "action": "react",
    "status": "success",
    "timestamp": "2025-09-22T10:40:00Z"
  }
}
```

#### ✏️ Edit Message

**POST** `/session/{sessionId}/message/edit`

Edit the text content of a previously sent message.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "message_id": "3EB0D098B5FD4BF3BC4327",
  "new_text": "Mensagem editada via API"
}
```

#### 🗑️ Delete Message

**POST** `/session/{sessionId}/message/delete`

Delete a message for yourself or everyone.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "message_id": "3EB0D098B5FD4BF3BC4327",
  "for_everyone": false
}
```

### 📝 Send Text Message

**POST** `/session/{sessionId}/message/send/text`

Send a text message to a WhatsApp contact.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "message": "Hello, World!"
}
```

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "key": {
      "remoteJid": "5511999999999@s.whatsapp.net",
      "id": "3EB0123456789ABCDEF",
      "fromMe": true
    },
    "message": {
      "conversation": "Hello, World!"
    },
    "timestamp": 1727000000
  }
}
```

---

---

## 📰 Newsletter Endpoints

### 📝 Create Newsletter

**POST** `/session/{sessionId}/newsletter`

Create a new newsletter channel.

**Request Body:**

```json
{
  "name": "Tech Updates",
  "description": "Latest technology news and updates"
}
```

**Response:**

```json
{
  "success": true,
  "code": 201,
  "data": {
    "newsletter_id": "120363123456789012@newsletter",
    "name": "Tech Updates",
    "description": "Latest technology news and updates",
    "created_at": "2025-09-22T10:45:00Z"
  }
}
```

### 📋 List Newsletters

**GET** `/session/{sessionId}/newsletter/list`

Get all subscribed newsletters.

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "newsletters": [
      {
        "id": "120363123456789012@newsletter",
        "name": "Tech Updates",
        "description": "Latest technology news",
        "subscriber_count": 1250,
        "muted": false
      }
    ]
  }
}
```

### 🔇 Toggle Newsletter Mute

**POST** `/session/{sessionId}/newsletter/{newsletterId}/mute`

Mute or unmute a newsletter.

**Request Body:**

```json
{
  "mute": true
}
```

---

## 🖼️ Media Endpoints

All media endpoints support **3 input formats**:

- **Base64**: `"iVBORw0KGgoAAAANSUhEUgAA..."`
- **Data URL**: `"data:image/jpeg;base64,iVBORw0KGgoAAAANSUhEUgAA..."`
- **HTTP/HTTPS URL**: `"https://example.com/image.jpg"`

### 🖼️ Send Image

**POST** `/session/{sessionId}/send/image`

Send an image with optional caption.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "image": "https://picsum.photos/800/600",
  "caption": "Beautiful image!"
}
```

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "key": {
      "remoteJid": "5511999999999@s.whatsapp.net",
      "id": "3EB0123456789ABCDEF",
      "fromMe": true
    },
    "message": {
      "image": {
        "url": "https://picsum.photos/800/600",
        "caption": "Beautiful image!",
        "mimetype": "image/jpeg"
      }
    },
    "timestamp": 1727000000
  }
}
```

### 🎵 Send Audio

**POST** `/session/{sessionId}/send/audio`

Send an audio file with optional PTT (Push-to-Talk) mode.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "audio": "data:audio/mp3;base64,SUQzBAAAAAAAI1RTU0U...",
  "ptt": true
}
```

**Parameters:**

- `audio` (string, required): Audio data in base64, data URL, or HTTP/HTTPS URL
- `ptt` (boolean, optional): Enable Push-to-Talk mode (default: false)
  - When `true`: Audio is sent as voice message (PTT) with `audio/ogg; codecs=opus` MIME type
  - When `false`: Audio is sent as regular audio file with original MIME type

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "key": {
      "remoteJid": "5511999999999@s.meow.net",
      "id": "3EB0123456789ABCDEF",
      "fromMe": true
    },
    "message": {
      "audio": {
        "url": "...",
        "mimetype": "audio/mpeg",
        "ptt": true
      }
    },
    "timestamp": 1757961000
  }
}
```

### 🎥 Send Video

**POST** `/session/{sessionId}/send/video`

Send a video file with optional caption.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "video": "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4",
  "caption": "Amazing video!"
}
```

**Parameters:**

- `video` (string, required): Video data in base64, data URL, or HTTP/HTTPS URL
- `caption` (string, optional): Video caption

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "key": {
      "remoteJid": "5511999999999@s.meow.net",
      "id": "3EB0123456789ABCDEF",
      "fromMe": true
    },
    "message": {
      "video": {
        "url": "...",
        "caption": "Amazing video!",
        "mimetype": "video/mp4",
        "gifPlayback": false
      }
    },
    "timestamp": 1757961000
  }
}
```

### 📄 Send Document

**POST** `/session/{sessionId}/send/document`

Send a document file with filename.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "document": "https://www.w3.org/WAI/ER/tests/xhtml/testfiles/resources/pdf/dummy.pdf",
  "filename": "document.pdf"
}
```

**Parameters:**

- `document` (string, required): Document data in base64, data URL, or HTTP/HTTPS URL
- `filename` (string, required): Document filename with extension

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "key": {
      "remoteJid": "5511999999999@s.meow.net",
      "id": "3EB0123456789ABCDEF",
      "fromMe": true
    },
    "message": {
      "document": {
        "url": "...",
        "fileName": "document.pdf",
        "mimetype": "application/pdf"
      }
    },
    "timestamp": 1757961000
  }
}
```

### 🎭 Send Sticker

**POST** `/session/{sessionId}/send/sticker`

Send a sticker (WebP format recommended).

**Request Body:**

```json
{
  "phone": "5511999999999",
  "sticker": "https://picsum.photos/512/512.webp"
}
```

**Parameters:**

- `sticker` (string, required): Sticker data in base64, data URL, or HTTP/HTTPS URL

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": {
    "key": {
      "remoteJid": "5511999999999@s.meow.net",
      "id": "3EB0123456789ABCDEF",
      "fromMe": true
    },
    "message": {
      "sticker": {
        "url": "...",
        "mimetype": "image/webp"
      }
    },
    "timestamp": 1757961000
  }
}
```

---

## 📍 Location Endpoint

### 📍 Send Location

**POST** `/session/{sessionId}/send/location`

Send a location with coordinates.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "latitude": -23.5505,
  "longitude": -46.6333,
  "name": "São Paulo",
  "address": "São Paulo, SP, Brazil"
}
```

**Parameters:**

- `latitude` (number, required): Location latitude
- `longitude` (number, required): Location longitude
- `name` (string, optional): Location name
- `address` (string, optional): Location address

---

## 👤 Contact Endpoint

### 👤 Send Contact

**POST** `/session/{sessionId}/send/contact`

Send a contact card.

**Request Body:**

```json
{
  "phone": "5511999999999",
  "contact": {
    "name": "John Doe",
    "phone": "5511888888888"
  }
}
```

---

## ⚠️ Error Responses

### 400 Bad Request

```json
{
  "success": false,
  "code": 400,
  "data": {
    "key": {
      "remoteJid": "",
      "id": "",
      "fromMe": false
    },
    "message": {},
    "timestamp": 1757961000
  },
  "error": {
    "code": "INVALID_MEDIA",
    "message": "Invalid image data",
    "details": "failed to download from URL: HTTP 404"
  }
}
```

### 401 Unauthorized

```json
{
  "success": false,
  "code": 401,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid API key"
  }
}
```

### 404 Not Found

```json
{
  "success": false,
  "code": 404,
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Session not found"
  }
}
```

---

## 📝 Notes

### Media Format Support

- **Images**: JPEG, PNG, GIF, WebP
- **Audio**: MP3, AAC, OGG, WAV
- **Video**: MP4, AVI, MOV, WebM
- **Documents**: PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, TXT, etc.
- **Stickers**: WebP (recommended), PNG, JPEG

### File Size Limits

- **Images**: Up to 16MB
- **Audio**: Up to 16MB
- **Video**: Up to 64MB
- **Documents**: Up to 100MB
- **Stickers**: Up to 1MB

### Session ID

The `sessionId` parameter can be:

- **UUID**: `8e30680e-c96b-4361-bf00-4e62b17dae8f`
- **Name**: `default`, `main`, `production`, etc.

### Phone Number Format

Phone numbers should be in international format without `+`:

- ✅ Correct: `5511999999999`
- ❌ Incorrect: `+55 11 99999-9999`

---

## 🚀 Examples

### cURL Examples

**Create Session:**

```bash
curl -X POST 'http://localhost:8080/sessions/create' \
  -H 'Authorization: YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{"session_id": "my-session", "name": "Production"}'
```

**Connect Session:**

```bash
curl -X POST 'http://localhost:8080/sessions/my-session/connect' \
  -H 'Authorization: YOUR_API_KEY'
```

**Send Text:**

```bash
curl -X POST 'http://localhost:8080/session/my-session/message/send/text' \
  -H 'Authorization: YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "5511999999999", "message": "Hello from zpmeow!"}'
```

**Send Image from URL:**

```bash
curl -X POST 'http://localhost:8080/session/my-session/message/send/image' \
  -H 'Authorization: YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "5511999999999", "image": "https://picsum.photos/800/600", "caption": "Random image"}'
```

**React to Message:**

```bash
curl -X POST 'http://localhost:8080/session/my-session/message/react' \
  -H 'Authorization: YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{"phone": "5511999999999", "message_id": "3EB0123456789ABCDEF", "emoji": "👍"}'
```

**Create Newsletter:**

```bash
curl -X POST 'http://localhost:8080/session/my-session/newsletter' \
  -H 'Authorization: YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{"name": "Tech Updates", "description": "Latest tech news"}'
```

---

## 👥 Group Management

### 🏗️ Create Group

**POST** `/session/{sessionId}/group/create`

Create a new WhatsApp group.

**Request Body:**

```json
{
  "name": "Development Team",
  "participants": ["5511999999999", "5511888888888"]
}
```

### 👤 Update Participants

**POST** `/session/{sessionId}/group/participants/update`

Add or remove group participants.

**Request Body:**

```json
{
  "group_jid": "120363123456789012@g.us",
  "action": "add",
  "participants": ["5511777777777"]
}
```

### 🖼️ Set Group Photo

**POST** `/session/{sessionId}/group/photo`

Set group profile photo.

**Request Body:**

```json
{
  "group_jid": "120363123456789012@g.us",
  "image": "https://example.com/group-photo.jpg"
}
```

---

## 🔧 Additional Session Management

### 📋 List Sessions

**GET** `/sessions/list`

Get all available sessions.

**Response:**

```json
{
  "success": true,
  "code": 200,
  "data": [
    {
      "id": "8e30680e-c96b-4361-bf00-4e62b17dae8f",
      "name": "default",
      "status": "connected",
      "device_jid": "5511999999999:84@s.whatsapp.net"
    }
  ]
}
```

### 🔌 Disconnect Session

**POST** `/sessions/{id}/disconnect`

Disconnect a session from WhatsApp.

---

## 📊 Status Codes

| Code | Status | Description |
|------|--------|-------------|
| 200 | OK | Request successful |
| 400 | Bad Request | Invalid request data |
| 401 | Unauthorized | Invalid API key |
| 404 | Not Found | Session or resource not found |
| 500 | Internal Server Error | Server error |

---

## 🔄 Webhooks

Configure webhooks to receive real-time events:

### 🎯 Set Webhook

**POST** `/session/{sessionId}/webhook`

```json
{
  "url": "https://your-server.com/webhook",
  "events": ["message", "receipt", "presence"]
}
```

### 📨 Webhook Events

**Message Received:**

```json
{
  "event": "message",
  "sessionID": "default",
  "data": {
    "from": "5511999999999@s.meow.net",
    "message": "Hello!",
    "timestamp": 1757961000
  }
}
```

**Message Receipt:**

```json
{
  "event": "receipt",
  "sessionID": "default",
  "data": {
    "message_id": "3EB0123456789ABCDEF",
    "status": "read",
    "timestamp": 1757961000
  }
}
```

---

## 🛡️ Best Practices

### 🔐 Security

- Keep your API key secure and never expose it in client-side code
- Use HTTPS in production environments
- Implement rate limiting on your side

### 📈 Performance

- Use URLs for large media files instead of base64 when possible
- Implement proper error handling and retry logic
- Monitor your webhook endpoints for reliability

### 📱 meow Compliance

- Respect meow's terms of service
- Don't send spam or unsolicited messages
- Implement proper user consent mechanisms
- Follow meow Business API guidelines

---

## 🐛 Troubleshooting

### Common Issues

**1. Session Not Connected**

```json
{
  "error": {
    "code": "SESSION_NOT_CONNECTED",
    "message": "Session is not connected to meow"
  }
}
```

**Solution:** Connect the session using `/sessions/{id}/connect`

**2. Invalid Phone Number**

```json
{
  "error": {
    "code": "INVALID_PHONE",
    "message": "Invalid phone number format"
  }
}
```

**Solution:** Use international format without `+`: `5511999999999`

**3. Media Download Failed**

```json
{
  "error": {
    "code": "INVALID_MEDIA",
    "message": "Failed to download from URL: HTTP 404"
  }
}
```

**Solution:** Ensure the URL is accessible and returns the correct content type

**4. File Too Large**

```json
{
  "error": {
    "code": "FILE_TOO_LARGE",
    "message": "File size exceeds maximum limit"
  }
}
```

**Solution:** Reduce file size or use a different format

---

## 📞 Support

For technical support and questions:

- 📧 Email: <support@zpmeow.com>
- 📚 Documentation: Built-in Swagger UI at `/swagger/`
- 🐛 Issues: GitHub Issues
- 💬 Community: Discord/Telegram

---

## 🔗 Additional Resources

- **Swagger UI**: `http://localhost:8080/swagger/` - Interactive API documentation
- **Health Check**: `http://localhost:8080/health` - API health status
- **Database Admin**: `http://localhost:3000` - DbGate interface (development)
- **MinIO Console**: `http://localhost:9001` - File storage management

---

## 📄 License

This API documentation is part of zpmeow WhatsApp API.
Built with ❤️ using Go Fiber and whatsmeow.
© 2025 zpmeow. All rights reserved.
