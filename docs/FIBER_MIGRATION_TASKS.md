# 🚀 Plano de Migração: Gin → Fiber

## 📋 **Visão Geral**

Este documento detalha o plano de execução para migração do framework HTTP de **Gin** para **Fiber** no projeto zpmeow.

**Duração Estimada**: 9-12 dias úteis  
**Complexidade**: Média-Alta  
**Risco**: Baixo-Médio (com mitigações adequadas)

---

## 🎯 **Fases da Migração**

### 📦 **Fase 1: Preparação** (1-2 dias)

#### ✅ **1.1 Instalar Fiber e dependências**
```bash
go get github.com/gofiber/fiber/v2
go get github.com/gofiber/swagger
go get github.com/gofiber/cors
go get github.com/gofiber/recover
```

#### ✅ **1.2 Criar branch de migração**
```bash
git checkout -b feature/migrate-to-fiber
git push -u origin feature/migrate-to-fiber
```

#### ✅ **1.3 Backup da arquitetura atual**
- Documentar implementação Gin atual
- Criar snapshot do código
- Listar todas as dependências Gin

#### ✅ **1.4 Setup de benchmarks**
- Criar testes de performance baseline
- Setup de métricas de comparação
- Preparar ambiente de testes

---

### 🔧 **Fase 2: Core Migration** (2-3 dias)

#### ✅ **2.1 Migrar main.go**
```go
// Antes (Gin)
ginRouter := gin.New()
ginRouter.Use(middleware.Logger())

// Depois (Fiber)
app := fiber.New(fiber.Config{
    ErrorHandler: customErrorHandler,
})
app.Use(middleware.Logger())
```

#### ✅ **2.2 Migrar middlewares base**
- **CORS**: `gin-contrib/cors` → `gofiber/cors`
- **Recovery**: `gin.Recovery()` → `gofiber/recover`
- **Validation**: Adaptar para Fiber context

#### ✅ **2.3 Migrar middleware de autenticação**
```go
// Gin → Fiber
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // auth logic
        c.Next()
    }
}

func AuthMiddleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // auth logic
        return c.Next()
    }
}
```

#### ✅ **2.4 Migrar middleware de logging**
- Adaptar HTTPLogEntry para Fiber
- Converter `gin.Context` → `*fiber.Ctx`
- Manter structured logging

#### ✅ **2.5 Migrar correlation ID middleware**
- Adaptar context handling
- Manter compatibilidade com logging

#### ✅ **2.6 Migrar router setup**
- Converter grupos de rotas
- Adaptar middleware assignment
- Manter estrutura de URLs

---

### 🎯 **Fase 3: Handlers** (3-4 dias)

#### ✅ **3.1 Migrar BaseHandler**
```go
// Principais mudanças
type BaseHandler struct {
    logger logging.Logger
}

// Gin → Fiber
func (h *BaseHandler) BindJSON(c *gin.Context, obj interface{}) error {
    return c.ShouldBindJSON(obj)
}

func (h *BaseHandler) BindJSON(c *fiber.Ctx, obj interface{}) error {
    return c.BodyParser(obj)
}
```

#### ✅ **3.2 Migrar HealthHandler**
- Converter endpoints de health check
- Adaptar métricas
- Manter compatibilidade de resposta

#### ✅ **3.3 Migrar SessionHandler**
- Converter todos os endpoints de sessão
- Adaptar binding de parâmetros
- Manter DTOs compatíveis

#### ✅ **3.4 Migrar MessageHandler**
- Converter endpoints de mensagens
- Adaptar upload de mídia
- Manter validações

#### ✅ **3.5 Migrar GroupHandler**
- Converter operações de grupo
- Adaptar validações
- Manter estrutura de resposta

#### ✅ **3.6 Migrar ContactHandler**
- Converter operações de contato
- Adaptar queries
- Manter compatibilidade

#### ✅ **3.7 Migrar ChatHandler**
- Converter operações de chat
- Adaptar downloads
- Manter funcionalidades

#### ✅ **3.8 Migrar demais handlers**
- **PrivacyHandler**: Configurações de privacidade
- **CommunityHandler**: Operações de comunidade
- **NewsletterHandler**: Newsletters
- **WebhookHandler**: Webhooks
- **MediaHandler**: Upload/download de mídia

---

### 🧪 **Fase 4: Testes e Validação** (2-3 dias)

#### ✅ **4.1 Criar testes unitários**
- Testes para todos os handlers migrados
- Mocks para dependências
- Coverage > 80%

#### ✅ **4.2 Testes de integração**
- End-to-end testing
- Validação de APIs
- Testes de autenticação

#### ✅ **4.3 Benchmarks de performance**
```bash
# Comparação Gin vs Fiber
go test -bench=. -benchmem ./benchmarks/
```

#### ✅ **4.4 Testes de carga**
- Load testing com múltiplas sessões
- Stress testing
- Memory leak detection

#### ✅ **4.5 Validação de APIs**
- Swagger compatibility
- Response format validation
- Error handling verification

#### ✅ **4.6 Testes de regressão**
- Funcionalidades críticas
- WhatsApp integration
- Database operations

---

### 🚀 **Fase 5: Deploy e Monitoramento** (1-2 dias)

#### ✅ **5.1 Deploy em ambiente de teste**
- Staging environment
- Integration testing
- Performance validation

#### ✅ **5.2 Monitoramento de performance**
- Setup de métricas
- Alertas de performance
- Dashboard de monitoramento

#### ✅ **5.3 Deploy gradual em produção**
- Blue-green deployment
- Canary release (10% → 50% → 100%)
- Rollback plan

#### ✅ **5.4 Monitoramento pós-deploy**
- 48h de monitoramento intensivo
- Análise de métricas
- Validação de performance

#### ✅ **5.5 Documentação final**
- Atualizar README
- API documentation
- Migration notes

#### ✅ **5.6 Cleanup do código Gin**
- Remover dependências Gin
- Cleanup de imports
- Archive old code

---

## 📊 **Métricas de Sucesso**

### 🎯 **Performance Targets**
- **Throughput**: +40% requests/second
- **Latência P99**: -30% reduction
- **Memory Usage**: -25% reduction
- **CPU Usage**: -20% reduction

### ✅ **Quality Gates**
- **Test Coverage**: > 80%
- **Zero Breaking Changes**: APIs mantêm compatibilidade
- **Zero Downtime**: Deploy sem interrupção
- **Performance Regression**: < 5% degradation allowed

---

## ⚠️ **Riscos e Mitigações**

### 🚨 **Riscos Identificados**
1. **Breaking Changes**: API incompatibilities
2. **Performance Regression**: Unexpected slowdowns
3. **Integration Issues**: Third-party compatibility
4. **Data Loss**: Session/cache issues

### 🛡️ **Mitigações**
1. **Extensive Testing**: Unit + Integration + E2E
2. **Gradual Rollout**: Canary deployment
3. **Monitoring**: Real-time metrics
4. **Rollback Plan**: Quick revert capability

---

## 📅 **Timeline**

| Fase | Duração | Dependências |
|------|---------|--------------|
| **Preparação** | 1-2 dias | - |
| **Core Migration** | 2-3 dias | Fase 1 |
| **Handlers** | 3-4 dias | Fase 2 |
| **Testes** | 2-3 dias | Fase 3 |
| **Deploy** | 1-2 dias | Fase 4 |
| **Total** | **9-14 dias** | - |

---

## 💡 **Exemplo Prático: SessionHandler**

### **Antes (Gin)**
```go
func (h *SessionHandler) GetSession(c *gin.Context) {
    sessionID := c.Param("sessionId")

    session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
        return
    }

    c.JSON(http.StatusOK, session)
}
```

### **Depois (Fiber)**
```go
func (h *SessionHandler) GetSession(c *fiber.Ctx) error {
    sessionID := c.Params("sessionId")

    session, err := h.sessionService.GetSession(c.Context(), sessionID)
    if err != nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
    }

    return c.Status(fiber.StatusOK).JSON(session)
}
```

### **Principais Mudanças**
1. **Return Type**: `void` → `error`
2. **Parameters**: `c.Param()` → `c.Params()`
3. **Context**: `c.Request.Context()` → `c.Context()`
4. **Response**: `c.JSON()` → `c.Status().JSON()`
5. **Error Handling**: `return` → `return error`

---

## 🎯 **Próximos Passos**

1. **Aprovação**: Review e aprovação do plano
2. **Resource Allocation**: Alocar desenvolvedor(es)
3. **Kickoff**: Iniciar Fase 1
4. **Daily Standups**: Acompanhamento diário
5. **Milestone Reviews**: Review ao final de cada fase
