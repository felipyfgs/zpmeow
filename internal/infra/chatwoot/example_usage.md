# Exemplo de Uso da Integração Chatwoot

Este documento mostra como usar a integração Chatwoot completa na zpmeow.

## 1. Configuração Inicial

### Configurar Chatwoot via API

```bash
# Configurar integração Chatwoot para uma sessão
curl -X POST http://localhost:8080/session/minha-sessao/chatwoot/config \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "accountId": "1",
    "token": "seu-token-chatwoot",
    "url": "https://app.chatwoot.com",
    "nameInbox": "WhatsApp zpmeow",
    "signMsg": false,
    "reopenConversation": true,
    "conversationPending": false,
    "mergeBrazilContacts": true,
    "autoCreate": true,
    "organization": "Minha Empresa",
    "ignoreJids": ["status@broadcast"]
  }'
```

### Resposta da Configuração

```json
{
  "success": true,
  "message": "Chatwoot configured successfully",
  "data": {
    "enabled": true,
    "accountId": "1",
    "url": "https://app.chatwoot.com",
    "nameInbox": "WhatsApp zpmeow",
    "signMsg": false,
    "reopenConversation": true,
    "conversationPending": false,
    "mergeBrazilContacts": true,
    "autoCreate": true,
    "organization": "Minha Empresa",
    "ignoreJids": ["status@broadcast"],
    "webhookUrl": "http://localhost:8080/session/minha-sessao/chatwoot/webhook"
  }
}
```

## 2. Verificar Status

```bash
# Verificar status da integração
curl -X GET http://localhost:8080/session/minha-sessao/chatwoot/status
```

### Resposta do Status

```json
{
  "success": true,
  "message": "Chatwoot status retrieved",
  "data": {
    "enabled": true,
    "connected": true,
    "inboxId": 123,
    "inboxName": "WhatsApp zpmeow",
    "lastSync": "2024-01-15T10:30:00Z",
    "messagesCount": 150,
    "contactsCount": 45
  }
}
```

## 3. Testar Conexão

```bash
# Testar conexão com Chatwoot
curl -X POST http://localhost:8080/session/minha-sessao/chatwoot/test \
  -H "Content-Type: application/json" \
  -d '{
    "accountId": "1",
    "token": "seu-token-chatwoot",
    "url": "https://app.chatwoot.com"
  }'
```

### Resposta do Teste

```json
{
  "success": true,
  "message": "Connection test completed",
  "data": {
    "success": true,
    "message": "Successfully connected to Chatwoot",
    "accountInfo": {
      "id": 1,
      "name": "Test Account"
    },
    "inboxesCount": 3
  }
}
```

## 4. Configurar Webhook no Chatwoot

No painel do Chatwoot:

1. Vá em **Settings** > **Inboxes**
2. Edite sua inbox API
3. Configure a **Webhook URL**: `http://localhost:8080/session/minha-sessao/chatwoot/webhook`
4. Salve as configurações

## 5. Fluxo de Mensagens

### Mensagem WhatsApp → Chatwoot

Quando uma mensagem é recebida no WhatsApp, ela é automaticamente enviada para o Chatwoot:

```
WhatsApp → zpmeow → Chatwoot
```

### Mensagem Chatwoot → WhatsApp

Quando um agente responde no Chatwoot, a mensagem é enviada via webhook:

```
Chatwoot → Webhook → zpmeow → WhatsApp
```

## 6. Exemplos de Webhooks

### Webhook de Mensagem Recebida

```json
{
  "event": "message_created",
  "message_type": "outgoing",
  "id": 123,
  "content": "Olá! Como posso ajudá-lo?",
  "created_at": "2024-01-15T10:30:00Z",
  "private": false,
  "source_id": "WAID:msg_123",
  "content_type": "text",
  "sender": {
    "id": 1,
    "name": "Agente",
    "email": "agente@empresa.com"
  },
  "contact": {
    "id": 456,
    "name": "João Silva",
    "phone_number": "+5511999999999",
    "identifier": "5511999999999@s.whatsapp.net"
  },
  "conversation": {
    "id": 789,
    "account_id": 1,
    "inbox_id": 123,
    "status": "open"
  }
}
```

## 7. Gerenciar Configuração

### Obter Configuração Atual

```bash
curl -X GET http://localhost:8080/session/minha-sessao/chatwoot/config
```

### Atualizar Configuração

```bash
curl -X PUT http://localhost:8080/session/minha-sessao/chatwoot/config \
  -H "Content-Type: application/json" \
  -d '{
    "enabled": true,
    "accountId": "1",
    "token": "novo-token",
    "url": "https://app.chatwoot.com",
    "nameInbox": "WhatsApp zpmeow Updated",
    "signMsg": true,
    "signDelimiter": "\n\n---\nEnviado via zpmeow"
  }'
```

### Remover Configuração

```bash
curl -X DELETE http://localhost:8080/session/minha-sessao/chatwoot/config
```

## 8. Estrutura do Banco de Dados

A configuração é salva na tabela `chatwoot`:

```sql
SELECT * FROM chatwoot WHERE session_id = 'minha-sessao';
```

### Campos Principais

- `session_id`: ID da sessão WhatsApp
- `enabled`: Se a integração está habilitada
- `account_id`, `token`, `url`: Credenciais do Chatwoot
- `inbox_id`, `inbox_name`: Informações da inbox criada
- `messages_count`, `contacts_count`: Métricas
- `sync_status`: Status da sincronização
- `last_sync`: Última sincronização

## 9. Monitoramento

### Logs da Aplicação

```bash
# Filtrar logs do Chatwoot
tail -f app.log | grep chatwoot
```

### Métricas do Banco

```sql
-- Estatísticas gerais
SELECT 
    COUNT(*) as total_configs,
    COUNT(CASE WHEN enabled = true THEN 1 END) as enabled_configs,
    SUM(messages_count) as total_messages,
    SUM(contacts_count) as total_contacts
FROM chatwoot;

-- Configurações por status
SELECT sync_status, COUNT(*) 
FROM chatwoot 
GROUP BY sync_status;
```

## 10. Troubleshooting

### Problemas Comuns

1. **Webhook não funciona**
   - Verifique se a URL está acessível
   - Confirme as configurações no Chatwoot
   - Teste com ngrok para desenvolvimento

2. **Mensagens não aparecem no Chatwoot**
   - Verifique se a integração está habilitada
   - Confirme as credenciais
   - Verifique os logs da aplicação

3. **Contatos duplicados**
   - Habilite `mergeBrazilContacts`
   - Configure `ignoreJids` adequadamente

### Debug

```bash
# Testar webhook manualmente
curl -X POST http://localhost:8080/session/minha-sessao/chatwoot/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event": "message_created",
    "message_type": "outgoing",
    "content": "Teste manual",
    "contact": {
      "phone_number": "+5511999999999"
    },
    "conversation": {
      "id": 1
    }
  }'
```

## 11. Exemplo de Integração Completa

```go
package main

import (
    "log/slog"
    
    "github.com/gofiber/fiber/v2"
    "zpmeow/internal/application"
    "zpmeow/internal/infra/chatwoot"
    "zpmeow/internal/infra/database/repository"
    "zpmeow/internal/infra/http/handlers"
)

func main() {
    // Configurar dependências
    logger := slog.Default()
    
    // Repositórios
    chatwootRepo := repository.NewChatwootRepository(db)
    
    // Serviços
    sessionService := application.NewSessionApp(sessionRepo)
    chatwootIntegration := chatwoot.NewIntegration(logger)
    
    // Handlers
    chatwootHandler := handlers.NewChatwootHandler(
        sessionService, 
        chatwootIntegration, 
        chatwootRepo,
    )
    
    // Router
    app := fiber.New()
    
    // Rotas Chatwoot
    sessionGroup := app.Group("/session/:sessionId")
    chatwootGroup := sessionGroup.Group("/chatwoot")
    
    chatwootGroup.Post("/config", chatwootHandler.SetChatwootConfig)
    chatwootGroup.Get("/config", chatwootHandler.GetChatwootConfig)
    chatwootGroup.Put("/config", chatwootHandler.UpdateChatwootConfig)
    chatwootGroup.Delete("/config", chatwootHandler.DeleteChatwootConfig)
    chatwootGroup.Get("/status", chatwootHandler.GetChatwootStatus)
    chatwootGroup.Post("/webhook", chatwootHandler.ReceiveChatwootWebhook)
    chatwootGroup.Post("/test", chatwootHandler.TestChatwootConnection)
    
    app.Listen(":8080")
}
```

## 12. Próximos Passos

- Implementar sincronização de histórico
- Adicionar suporte a templates
- Criar dashboard de métricas
- Implementar retry automático para falhas
- Adicionar cache Redis para performance
