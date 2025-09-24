#!/bin/bash

# Script para simplificar nomes de arquivos onde faz sentido
# Princípio: Simplicidade quando não há ambiguidade
# Criado em: 24/09/2025

echo "🎯 Simplificando nomes de arquivos - removendo redundâncias desnecessárias"
echo "=========================================================================="

# Função para renomear arquivo com verificação
simplify_file() {
    local old_path="$1"
    local new_path="$2"
    local reason="$3"
    
    if [ -f "$old_path" ]; then
        echo "✅ Simplificando: $(basename "$old_path") → $(basename "$new_path") ($reason)"
        mv "$old_path" "$new_path"
    else
        echo "⚠️  Arquivo não encontrado: $old_path"
    fi
}

# =============================================================================
# APPLICATION LAYER - Simplificações
# =============================================================================
echo ""
echo "📁 APPLICATION LAYER:"

simplify_file "./internal/application/app_main.go" "./internal/application/app.go" "único no diretório"
simplify_file "./internal/application/common/error_application.go" "./internal/application/common/errors.go" "padrão Go"
simplify_file "./internal/application/ports/interface_events.go" "./internal/application/ports/events.go" "contexto claro"
simplify_file "./internal/application/ports/interface_ports.go" "./internal/application/ports/interfaces.go" "contexto claro"

# =============================================================================
# CONFIG - Simplificações
# =============================================================================
echo ""
echo "📁 CONFIG:"

simplify_file "./internal/config/config_main.go" "./internal/config/config.go" "arquivo principal"
simplify_file "./internal/config/config_defaults.go" "./internal/config/defaults.go" "contexto claro"
simplify_file "./internal/config/interface_config.go" "./internal/config/interfaces.go" "contexto claro"

# =============================================================================
# DOMAIN/COMMON - Simplificações
# =============================================================================
echo ""
echo "📁 DOMAIN/COMMON:"

simplify_file "./internal/domain/common/event_common.go" "./internal/domain/common/events.go" "contexto claro"
simplify_file "./internal/domain/common/interface_common.go" "./internal/domain/common/interfaces.go" "contexto claro"
simplify_file "./internal/domain/common/valueobject_common.go" "./internal/domain/common/valueobjects.go" "contexto claro"

# =============================================================================
# DOMAIN/SESSION - Simplificações
# =============================================================================
echo ""
echo "📁 DOMAIN/SESSION:"

simplify_file "./internal/domain/session/entity_session.go" "./internal/domain/session/entity.go" "contexto claro"
simplify_file "./internal/domain/session/error_session.go" "./internal/domain/session/errors.go" "padrão Go"
simplify_file "./internal/domain/session/event_session.go" "./internal/domain/session/events.go" "contexto claro"
simplify_file "./internal/domain/session/interface_repository.go" "./internal/domain/session/repository.go" "contexto claro"
simplify_file "./internal/domain/session/service_session.go" "./internal/domain/session/service.go" "contexto claro"
simplify_file "./internal/domain/session/valueobject_session.go" "./internal/domain/session/valueobjects.go" "contexto claro"

# =============================================================================
# CACHE - Simplificações
# =============================================================================
echo ""
echo "📁 CACHE:"

simplify_file "./internal/infra/cache/cache_noop.go" "./internal/infra/cache/noop.go" "contexto claro"
simplify_file "./internal/infra/cache/cache_redis.go" "./internal/infra/cache/redis.go" "contexto claro"
simplify_file "./internal/infra/cache/repo_cache.go" "./internal/infra/cache/repository.go" "contexto claro"

# =============================================================================
# DATABASE - Simplificações
# =============================================================================
echo ""
echo "📁 DATABASE:"

simplify_file "./internal/infra/database/client_database.go" "./internal/infra/database/database.go" "único no diretório"
simplify_file "./internal/infra/database/models/entity_models.go" "./internal/infra/database/models/models.go" "padrão Go"

# =============================================================================
# HTTP/ROUTES - Simplificações
# =============================================================================
echo ""
echo "📁 HTTP/ROUTES:"

simplify_file "./internal/infra/http/routes/router_main.go" "./internal/infra/http/routes/router.go" "único no diretório"

# =============================================================================
# LOGGING - Simplificações
# =============================================================================
echo ""
echo "📁 LOGGING:"

simplify_file "./internal/infra/logging/service_logger.go" "./internal/infra/logging/logger.go" "único no diretório"

# =============================================================================
# WEBHOOKS - Simplificações
# =============================================================================
echo ""
echo "📁 WEBHOOKS:"

simplify_file "./internal/infra/webhooks/client_webhook.go" "./internal/infra/webhooks/client.go" "contexto claro"
simplify_file "./internal/infra/webhooks/helper_retry.go" "./internal/infra/webhooks/retry.go" "contexto claro"
simplify_file "./internal/infra/webhooks/service_webhook.go" "./internal/infra/webhooks/service.go" "contexto claro"

# =============================================================================
# CHATWOOT - Simplificações Específicas
# =============================================================================
echo ""
echo "📁 CHATWOOT:"

simplify_file "./internal/infra/chatwoot/adapter_chatwoot.go" "./internal/infra/chatwoot/adapters.go" "contexto claro"
simplify_file "./internal/infra/chatwoot/client_chatwoot.go" "./internal/infra/chatwoot/client.go" "contexto claro"
simplify_file "./internal/infra/chatwoot/helper_parser.go" "./internal/infra/chatwoot/parser.go" "contexto claro"
simplify_file "./internal/infra/chatwoot/limiter_rate.go" "./internal/infra/chatwoot/ratelimiter.go" "nome composto tradicional"
simplify_file "./internal/infra/chatwoot/mapper_message.go" "./internal/infra/chatwoot/messagemapper.go" "nome composto tradicional"
simplify_file "./internal/infra/chatwoot/processor_media.go" "./internal/infra/chatwoot/mediaprocessor.go" "nome composto tradicional"
simplify_file "./internal/infra/chatwoot/validation_chatwoot.go" "./internal/infra/chatwoot/validator.go" "contexto claro"

echo ""
echo "🎉 Simplificação concluída!"
echo ""
echo "📋 RESUMO DAS SIMPLIFICAÇÕES:"
echo "   • Arquivos únicos → nomes simples"
echo "   • Contexto claro → remove prefixo redundante"
echo "   • Padrões Go → config.go, errors.go, etc."
echo "   • Nomes compostos → mediaprocessor.go, ratelimiter.go"
echo ""
echo "✅ Mantém organização onde necessário:"
echo "   • DTOs → dto_*.go (múltiplos no diretório)"
echo "   • Handlers → handler_*.go (múltiplos no diretório)"
echo "   • Middlewares → middleware_*.go (múltiplos no diretório)"
echo "   • Repositórios → repo_*.go (múltiplos no diretório)"
echo "   • Use Cases → usecase_*.go (múltiplos no diretório)"
echo "   • WMeow Services → service_*.go (múltiplos no diretório)"
echo ""
echo "🎯 Resultado: Simplicidade inteligente aplicada!"
echo "=========================================================================="
