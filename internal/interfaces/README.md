# HTTP Interface Layer

Esta camada implementa a interface HTTP da aplicação, fornecendo endpoints REST para interação com o sistema WhatsApp.

## 📁 Estrutura

```
internal/interfaces/http/
├── dto/                    # Data Transfer Objects
│   ├── common.go          # Estruturas de resposta padronizadas
│   ├── types.go           # Tipos customizados e validações
│   ├── validation.go      # Validações de entrada
│   ├── chat.go           # DTOs para operações de chat
│   ├── contact.go        # DTOs para operações de contato
│   ├── group.go          # DTOs para operações de grupo
│   ├── message.go        # DTOs para operações de mensagem
│   ├── newsletter.go     # DTOs para operações de newsletter
│   ├── privacy.go        # DTOs para configurações de privacidade
│   ├── webhook.go        # DTOs para configurações de webhook
│   └── responses.go      # Tipos de resposta legados
├── handlers/              # Handlers HTTP
│   ├── handler.go        # Handler base com funcionalidades comuns
│   ├── health.go         # Endpoints de saúde
│   ├── session.go        # Gerenciamento de sessões
│   ├── contact.go        # Operações de contato
│   ├── chat.go           # Operações de chat
│   ├── message.go        # Envio de mensagens
│   ├── group.go          # Gerenciamento de grupos
│   ├── community.go      # Operações de comunidade
│   ├── newsletter.go     # Operações de newsletter
│   ├── privacy.go        # Configurações de privacidade
│   ├── webhook.go        # Configurações de webhook
│   └── media.go          # Operações de mídia
├── routes/               # Configuração de rotas
│   └── router.go         # Setup das rotas e agrupamentos
├── server.go             # Servidor HTTP principal
└── README.md             # Esta documentação
```

## 🏗️ Arquitetura

### Padrões Implementados

1. **Handler Base Pattern**: Todos os handlers herdam funcionalidades comuns do `BaseHandler`
2. **Dependency Injection**: Handlers recebem dependências via construtor
3. **Standardized Responses**: Respostas padronizadas usando estruturas comuns
4. **Route Grouping**: Rotas organizadas logicamente em grupos
5. **Middleware Integration**: Autenticação e logging integrados

### Fluxo de Requisição

```
HTTP Request → Middleware → Router → Handler → Application Layer → Domain Layer
     ↓             ↓          ↓         ↓            ↓              ↓
  Validation   Logging    Route     Business     Domain         Response
               Auth      Matching   Logic        Rules
```

## 📋 Componentes Principais

### DTOs (Data Transfer Objects)

- **common.go**: Estruturas de resposta padronizadas (`StandardResponse`, `ErrorInfo`)
- **types.go**: Tipos customizados com validação (`SessionID`, `MessageID`, `URL`)
- **validation.go**: Validadores de entrada consolidados

### Handlers

- **BaseHandler**: Funcionalidades comuns (logging, binding, responses)
- **Handlers específicos**: Implementam endpoints para cada domínio

### Rotas

- **HandlerDependencies**: Estrutura que agrupa todos os handlers
- **SetupRoutes**: Configuração centralizada de todas as rotas

### Servidor

- **Server**: Encapsula configuração e inicialização do servidor HTTP
- **ServerConfig**: Configurações do servidor (timeouts, porta, etc.)

## 🔧 Uso

### Inicialização do Servidor

```go
// Configuração
config := http.DefaultServerConfig()
config.Port = "8080"

// Dependências
sessionApp := application.NewSessionApp(...)
webhookApp := application.NewWebhookApp(...)
wmeowService := wmeow.NewService(...)
authMiddleware := middleware.NewAuthMiddleware(...)

// Criar servidor
server := http.NewServer(
    config,
    sessionApp,
    webhookApp,
    wmeowService,
    authMiddleware,
)

// Iniciar
if err := server.Start(); err != nil {
    log.Fatal(err)
}
```

### Criação de Novos Handlers

```go
type MyHandler struct {
    *handlers.BaseHandler
    myService MyService
}

func NewMyHandler(myService MyService) *MyHandler {
    return &MyHandler{
        BaseHandler: handlers.NewBaseHandler("my-handler"),
        myService:   myService,
    }
}

func (h *MyHandler) MyEndpoint(c *gin.Context) {
    var req MyRequest
    if err := h.BindJSON(c, &req); err != nil {
        return // Error response already sent
    }
    
    result, err := h.myService.DoSomething(c.Request.Context(), req)
    if err != nil {
        h.SendInternalErrorResponse(c, err)
        return
    }
    
    h.SendSuccessResponse(c, dto.StatusOK, result)
}
```

## ✅ Melhorias Implementadas

### Antes da Refatoração
- ❌ Duplicação de código entre handlers
- ❌ Respostas inconsistentes
- ❌ Validações espalhadas
- ❌ Rotas desorganizadas
- ❌ Handlers não declarados

### Depois da Refatoração
- ✅ Handler base com funcionalidades comuns
- ✅ Respostas padronizadas
- ✅ Validações centralizadas
- ✅ Rotas organizadas por domínio
- ✅ Dependency injection estruturada
- ✅ Logging consistente
- ✅ Tipos customizados com validação

## 🔄 Próximos Passos

1. **Implementar PrivacyHandler** quando disponível
2. **Adicionar testes unitários** para handlers
3. **Implementar rate limiting** nos middlewares
4. **Adicionar métricas** de performance
5. **Implementar cache** para respostas frequentes
6. **Adicionar documentação OpenAPI** completa

## 📚 Referências

- [Gin Web Framework](https://gin-gonic.com/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go HTTP Best Practices](https://golang.org/doc/articles/wiki/)
