# 🗄️ Cache System - zpmeow

Sistema de cache implementado com Redis para melhorar a performance da API zpmeow, evitando consultas desnecessárias ao banco de dados.

## 🎯 **Funcionalidades**

### ✅ **Implementado**

- **Cache de Sessões** - Sessões WhatsApp (TTL: 24h)
- **Cache de QR Codes** - QR codes temporários (TTL: 60s)
- **Cache de Credenciais** - Device JIDs (TTL: 6h)
- **Cache de Status** - Status das sessões (TTL: 5m)
- **Health Checks** - Monitoramento do Redis
- **Fallback Automático** - Funciona mesmo com Redis offline

## 🏗️ **Arquitetura**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP Handler  │───▶│ CachedRepository│───▶│  BaseRepository │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Redis Service  │
                       └─────────────────┘
```

### **Padrão Cache-Aside**

1. **Cache Hit**: Dados retornados diretamente do cache
2. **Cache Miss**: Busca no banco → Armazena no cache → Retorna dados
3. **Cache Invalidation**: Remove dados do cache quando atualizados

## 📁 **Estrutura de Arquivos**

```
internal/infra/cache/
├── redis.go          # Implementação Redis
├── noop.go           # Implementação No-Op (cache desabilitado)
├── repository.go     # Repository com cache
└── README.md         # Esta documentação
```

## 🔧 **Configuração**

### **Variáveis de Ambiente (Opcionais)**

```bash
# Configurações básicas (defaults funcionam bem)
CACHE_ENABLED=true
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
```

### **Configurações Avançadas (Código)**

```go
// Configurações automáticas com defaults sensatos
DefaultCacheConfig() CacheConfig {
    return CacheConfig{
        Enabled:       true,
        RedisHost:     "localhost",
        RedisPort:     "6379",
        PoolSize:      10,
        SessionTTL:    24 * time.Hour,
        QRCodeTTL:     60 * time.Second,
        CredentialTTL: 6 * time.Hour,
        StatusTTL:     5 * time.Minute,
    }
}
```

## 🚀 **Uso**

### **Automático**

O cache funciona automaticamente quando habilitado. Não requer mudanças no código existente.

### **Health Check**

```bash
curl http://localhost:8080/health
```

**Resposta com cache:**

```json
{
  "success": true,
  "data": {
    "status": "ok",
    "dependencies": {
      "database": "healthy",
      "cache": "healthy"
    }
  }
}
```

### **Métricas**

```bash
curl http://localhost:8080/metrics
```

**Resposta:**

```json
{
  "success": true,
  "data": {
    "cache": {
      "connected": true,
      "total_keys": 42,
      "version": "Redis"
    }
  }
}
```

## 📊 **Performance**

### **Benefícios Esperados**

- **70-80% redução** nas consultas ao banco
- **Respostas 5-10x mais rápidas** para dados em cache
- **Menor carga** no PostgreSQL
- **Melhor experiência** do usuário

### **TTL Otimizado**

- **Sessions (24h)**: Dados raramente mudam
- **QR Codes (60s)**: Dados temporários por natureza
- **Credentials (6h)**: Balanceio entre performance e segurança
- **Status (5m)**: Dados que mudam frequentemente

## 🔄 **Estratégias de Cache**

### **1. Session Cache**

```go
// Cache Hit - Retorna imediatamente
session := cache.GetSession(sessionID)

// Cache Miss - Busca no banco e cacheia
session := database.GetSession(sessionID)
cache.SetSession(sessionID, session, 24*time.Hour)
```

### **2. QR Code Cache**

```go
// Cache temporário para QR codes
cache.SetQRCode(sessionID, qrCode) // TTL: 60s
```

### **3. Credential Cache**

```go
// Cache de credenciais WhatsApp
cache.SetDeviceJID(sessionID, deviceJID, 6*time.Hour)
```

## 🛡️ **Resiliência**

### **Fallback Automático**

- **Redis offline**: Funciona normalmente (sem cache)
- **Redis lento**: Timeout automático → fallback para banco
- **Dados corrompidos**: Ignora cache → busca no banco

### **No-Op Service**

Quando cache está desabilitado, usa implementação no-op que não faz nada.

## 🧪 **Testes**

### **Teste Manual**

```bash
# 1. Inicie o Redis
docker compose up -d redis

# 2. Inicie a aplicação
make run

# 3. Teste uma sessão
curl -X POST http://localhost:8080/sessions/create \
  -H "Content-Type: application/json" \
  -d '{"name": "test-session"}'

# 4. Verifique o cache
curl http://localhost:8080/metrics
```

### **Logs de Debug**

```bash
# Ative logs de debug
LOG_LEVEL=debug make run

# Observe os logs de cache
# Cache HIT: "Retrieved session test-session from cache"
# Cache MISS: "Cache MISS for session test-session, fetching from database"
```

## 🔧 **Troubleshooting**

### **Redis não conecta**

```bash
# Verifique se Redis está rodando
docker compose ps redis

# Inicie o Redis
docker compose up -d redis

# Teste conexão
redis-cli ping
```

### **Cache não funciona**

```bash
# Verifique configuração
curl http://localhost:8080/health

# Verifique logs
tail -f log/app.log | grep cache
```

### **Performance não melhora**

- Verifique se TTL não está muito baixo
- Confirme que dados estão sendo cacheados
- Monitore hit rate nas métricas

## 📈 **Monitoramento**

### **Métricas Importantes**

- **Hit Rate**: % de requests que usam cache
- **Total Keys**: Número de itens em cache
- **Connected**: Status da conexão Redis

### **Logs Estruturados**

```json
{
  "level": "debug",
  "message": "Cache HIT for session test-session",
  "module": "cached-repo"
}
```

## 🎯 **Próximos Passos**

### **Melhorias Futuras**

- Cache de contatos WhatsApp
- Cache de grupos
- Métricas avançadas (hit rate, latência)
- Cache warming automático
- Compressão de dados grandes

### **Otimizações**

- Pipeline Redis para operações em lote
- Clustering Redis para alta disponibilidade
- Monitoramento com Prometheus/Grafana
