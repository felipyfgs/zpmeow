#!/bin/bash

# Script para renomear arquivos seguindo a convenção [categoria]_[especificação].go
# Criado em: 24/09/2025
# Convenção: [categoria]_[especificação].go

echo "🎯 Iniciando renomeação de arquivos seguindo convenção [categoria]_[especificação].go"
echo "=================================================================="

# Função para renomear arquivo com verificação
rename_file() {
    local old_path="$1"
    local new_path="$2"
    local description="$3"
    
    if [ -f "$old_path" ]; then
        echo "✅ Renomeando: $(basename "$old_path") → $(basename "$new_path") ($description)"
        mv "$old_path" "$new_path"
    else
        echo "⚠️  Arquivo não encontrado: $old_path"
    fi
}

# =============================================================================
# CHATWOOT MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO CHATWOOT:"

rename_file "./internal/infra/chatwoot/mediaprocessor.go" "./internal/infra/chatwoot/processor_media.go" "Processador de mídia"
rename_file "./internal/infra/chatwoot/messagemapper.go" "./internal/infra/chatwoot/mapper_message.go" "Mapeador de mensagens"
rename_file "./internal/infra/chatwoot/ratelimiter.go" "./internal/infra/chatwoot/limiter_rate.go" "Limitador de taxa"
rename_file "./internal/infra/chatwoot/validator.go" "./internal/infra/chatwoot/validation_chatwoot.go" "Validação Chatwoot"
rename_file "./internal/infra/chatwoot/client.go" "./internal/infra/chatwoot/client_chatwoot.go" "Cliente Chatwoot"
rename_file "./internal/infra/chatwoot/service.go" "./internal/infra/chatwoot/service_chatwoot.go" "Serviço Chatwoot"
rename_file "./internal/infra/chatwoot/integration.go" "./internal/infra/chatwoot/service_integration.go" "Serviço de integração"

# =============================================================================
# DATABASE MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO DATABASE:"

rename_file "./internal/infra/database/connection.go" "./internal/infra/database/client_database.go" "Cliente de banco"
rename_file "./internal/infra/database/models/models.go" "./internal/infra/database/models/entity_models.go" "Entidades do banco"

# Repositories
rename_file "./internal/infra/database/repository/chat.go" "./internal/infra/database/repository/repo_chat.go" "Repositório de chat"
rename_file "./internal/infra/database/repository/chatwoot.go" "./internal/infra/database/repository/repo_chatwoot.go" "Repositório Chatwoot"
rename_file "./internal/infra/database/repository/message.go" "./internal/infra/database/repository/repo_message.go" "Repositório de mensagem"
rename_file "./internal/infra/database/repository/session.go" "./internal/infra/database/repository/repo_session.go" "Repositório de sessão"
rename_file "./internal/infra/database/repository/webhook.go" "./internal/infra/database/repository/repo_webhook.go" "Repositório de webhook"
rename_file "./internal/infra/database/repository/zpcwmessage.go" "./internal/infra/database/repository/repo_zpcwmessage.go" "Repositório ZP mensagem"

# =============================================================================
# HTTP MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO HTTP:"

# DTOs
rename_file "./internal/infra/http/dto/chat.go" "./internal/infra/http/dto/dto_chat.go" "DTO de chat"
rename_file "./internal/infra/http/dto/chatwoot.go" "./internal/infra/http/dto/dto_chatwoot.go" "DTO Chatwoot"
rename_file "./internal/infra/http/dto/common.go" "./internal/infra/http/dto/dto_common.go" "DTO comum"
rename_file "./internal/infra/http/dto/community.go" "./internal/infra/http/dto/dto_community.go" "DTO comunidade"
rename_file "./internal/infra/http/dto/contact.go" "./internal/infra/http/dto/dto_contact.go" "DTO contato"
rename_file "./internal/infra/http/dto/group.go" "./internal/infra/http/dto/dto_group.go" "DTO grupo"
rename_file "./internal/infra/http/dto/media.go" "./internal/infra/http/dto/dto_media.go" "DTO mídia"
rename_file "./internal/infra/http/dto/messages.go" "./internal/infra/http/dto/dto_messages.go" "DTO mensagens"
rename_file "./internal/infra/http/dto/newsletter.go" "./internal/infra/http/dto/dto_newsletter.go" "DTO newsletter"
rename_file "./internal/infra/http/dto/privacy.go" "./internal/infra/http/dto/dto_privacy.go" "DTO privacidade"
rename_file "./internal/infra/http/dto/session.go" "./internal/infra/http/dto/dto_session.go" "DTO sessão"
rename_file "./internal/infra/http/dto/webhook.go" "./internal/infra/http/dto/dto_webhook.go" "DTO webhook"

# Handlers
rename_file "./internal/infra/http/handlers/chat.go" "./internal/infra/http/handlers/handler_chat.go" "Handler de chat"
rename_file "./internal/infra/http/handlers/chatwoot.go" "./internal/infra/http/handlers/handler_chatwoot.go" "Handler Chatwoot"
rename_file "./internal/infra/http/handlers/common.go" "./internal/infra/http/handlers/handler_common.go" "Handler comum"
rename_file "./internal/infra/http/handlers/community.go" "./internal/infra/http/handlers/handler_community.go" "Handler comunidade"
rename_file "./internal/infra/http/handlers/contact.go" "./internal/infra/http/handlers/handler_contact.go" "Handler contato"
rename_file "./internal/infra/http/handlers/group.go" "./internal/infra/http/handlers/handler_group.go" "Handler grupo"
rename_file "./internal/infra/http/handlers/health.go" "./internal/infra/http/handlers/handler_health.go" "Handler saúde"
rename_file "./internal/infra/http/handlers/media.go" "./internal/infra/http/handlers/handler_media.go" "Handler mídia"
rename_file "./internal/infra/http/handlers/message.go" "./internal/infra/http/handlers/handler_message.go" "Handler mensagem"
rename_file "./internal/infra/http/handlers/newsletter.go" "./internal/infra/http/handlers/handler_newsletter.go" "Handler newsletter"
rename_file "./internal/infra/http/handlers/privacy.go" "./internal/infra/http/handlers/handler_privacy.go" "Handler privacidade"
rename_file "./internal/infra/http/handlers/session.go" "./internal/infra/http/handlers/handler_session.go" "Handler sessão"
rename_file "./internal/infra/http/handlers/webhook.go" "./internal/infra/http/handlers/handler_webhook.go" "Handler webhook"

# Middleware
rename_file "./internal/infra/http/middleware/auth.go" "./internal/infra/http/middleware/middleware_auth.go" "Middleware auth"
rename_file "./internal/infra/http/middleware/correlation.go" "./internal/infra/http/middleware/middleware_correlation.go" "Middleware correlação"
rename_file "./internal/infra/http/middleware/cors.go" "./internal/infra/http/middleware/middleware_cors.go" "Middleware CORS"
rename_file "./internal/infra/http/middleware/logging.go" "./internal/infra/http/middleware/middleware_logging.go" "Middleware logging"
rename_file "./internal/infra/http/middleware/validation.go" "./internal/infra/http/middleware/middleware_validation.go" "Middleware validação"

# Routes
rename_file "./internal/infra/http/routes/router.go" "./internal/infra/http/routes/router_main.go" "Router principal"

# =============================================================================
# CACHE MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO CACHE:"

rename_file "./internal/infra/cache/noop.go" "./internal/infra/cache/cache_noop.go" "Cache noop"
rename_file "./internal/infra/cache/redis.go" "./internal/infra/cache/cache_redis.go" "Cache Redis"
rename_file "./internal/infra/cache/repository.go" "./internal/infra/cache/repo_cache.go" "Repositório cache"

# =============================================================================
# WEBHOOKS MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO WEBHOOKS:"

rename_file "./internal/infra/webhooks/client.go" "./internal/infra/webhooks/client_webhook.go" "Cliente webhook"
rename_file "./internal/infra/webhooks/retry.go" "./internal/infra/webhooks/helper_retry.go" "Helper retry"
rename_file "./internal/infra/webhooks/service.go" "./internal/infra/webhooks/service_webhook.go" "Serviço webhook"

# =============================================================================
# LOGGING MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO LOGGING:"

rename_file "./internal/infra/logging/logger.go" "./internal/infra/logging/service_logger.go" "Serviço de logging"

# =============================================================================
# CONFIG MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO CONFIG:"

rename_file "./internal/config/config.go" "./internal/config/config_main.go" "Configuração principal"
rename_file "./internal/config/defaults.go" "./internal/config/config_defaults.go" "Configurações padrão"
rename_file "./internal/config/interfaces.go" "./internal/config/interface_config.go" "Interfaces de config"

# =============================================================================
# DOMAIN MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO DOMAIN:"

# Common
rename_file "./internal/domain/common/events.go" "./internal/domain/common/event_common.go" "Eventos comuns"
rename_file "./internal/domain/common/interfaces.go" "./internal/domain/common/interface_common.go" "Interfaces comuns"
rename_file "./internal/domain/common/value_objects.go" "./internal/domain/common/valueobject_common.go" "Value objects comuns"

# Session
rename_file "./internal/domain/session/entity.go" "./internal/domain/session/entity_session.go" "Entidade sessão"
rename_file "./internal/domain/session/errors.go" "./internal/domain/session/error_session.go" "Erros de sessão"
rename_file "./internal/domain/session/events.go" "./internal/domain/session/event_session.go" "Eventos de sessão"
rename_file "./internal/domain/session/repository.go" "./internal/domain/session/interface_repository.go" "Interface repositório"
rename_file "./internal/domain/session/service.go" "./internal/domain/session/service_session.go" "Serviço de sessão"
rename_file "./internal/domain/session/value_objects.go" "./internal/domain/session/valueobject_session.go" "Value objects sessão"

# =============================================================================
# APPLICATION MODULE - Renomeações
# =============================================================================
echo ""
echo "📁 MÓDULO APPLICATION:"

# Common
rename_file "./internal/application/common/errors.go" "./internal/application/common/error_application.go" "Erros da aplicação"

# Ports
rename_file "./internal/application/ports/events.go" "./internal/application/ports/interface_events.go" "Interface eventos"
rename_file "./internal/application/ports/interfaces.go" "./internal/application/ports/interface_ports.go" "Interfaces principais"

# Use Cases - Chat
rename_file "./internal/application/usecases/chat/history.go" "./internal/application/usecases/chat/usecase_history.go" "Use case histórico"
rename_file "./internal/application/usecases/chat/list.go" "./internal/application/usecases/chat/usecase_list.go" "Use case listar"
rename_file "./internal/application/usecases/chat/manage.go" "./internal/application/usecases/chat/usecase_manage.go" "Use case gerenciar"

# Use Cases - Contact
rename_file "./internal/application/usecases/contact/contacts.go" "./internal/application/usecases/contact/usecase_contacts.go" "Use case contatos"

# Use Cases - Group
rename_file "./internal/application/usecases/group/create.go" "./internal/application/usecases/group/usecase_create.go" "Use case criar"
rename_file "./internal/application/usecases/group/list.go" "./internal/application/usecases/group/usecase_list.go" "Use case listar"
rename_file "./internal/application/usecases/group/manage.go" "./internal/application/usecases/group/usecase_manage.go" "Use case gerenciar"
rename_file "./internal/application/usecases/group/members.go" "./internal/application/usecases/group/usecase_members.go" "Use case membros"

# Use Cases - Messaging
rename_file "./internal/application/usecases/messaging/actions.go" "./internal/application/usecases/messaging/usecase_actions.go" "Use case ações"
rename_file "./internal/application/usecases/messaging/contact.go" "./internal/application/usecases/messaging/usecase_contact.go" "Use case contato"
rename_file "./internal/application/usecases/messaging/location.go" "./internal/application/usecases/messaging/usecase_location.go" "Use case localização"
rename_file "./internal/application/usecases/messaging/media.go" "./internal/application/usecases/messaging/usecase_media.go" "Use case mídia"
rename_file "./internal/application/usecases/messaging/text.go" "./internal/application/usecases/messaging/usecase_text.go" "Use case texto"

# Use Cases - Newsletter
rename_file "./internal/application/usecases/newsletter/newsletter.go" "./internal/application/usecases/newsletter/usecase_newsletter.go" "Use case newsletter"

# Use Cases - Session
rename_file "./internal/application/usecases/session/connect.go" "./internal/application/usecases/session/usecase_connect.go" "Use case conectar"
rename_file "./internal/application/usecases/session/create.go" "./internal/application/usecases/session/usecase_create.go" "Use case criar"
rename_file "./internal/application/usecases/session/delete.go" "./internal/application/usecases/session/usecase_delete.go" "Use case deletar"
rename_file "./internal/application/usecases/session/disconnect.go" "./internal/application/usecases/session/usecase_disconnect.go" "Use case desconectar"
rename_file "./internal/application/usecases/session/get.go" "./internal/application/usecases/session/usecase_get.go" "Use case obter"
rename_file "./internal/application/usecases/session/pair.go" "./internal/application/usecases/session/usecase_pair.go" "Use case parear"
rename_file "./internal/application/usecases/session/status.go" "./internal/application/usecases/session/usecase_status.go" "Use case status"

# Use Cases - Webhook
rename_file "./internal/application/usecases/webhook/webhook.go" "./internal/application/usecases/webhook/usecase_webhook.go" "Use case webhook"

# Application main
rename_file "./internal/application/app.go" "./internal/application/app_main.go" "Aplicação principal"

echo ""
echo "🎉 Renomeação concluída!"
echo "📊 Arquivos renomeados seguindo a convenção [categoria]_[especificação].go"
echo ""
echo "📋 RESUMO DAS CATEGORIAS APLICADAS:"
echo "   • service_*        → Serviços de negócio"
echo "   • usecase_*        → Casos de uso"
echo "   • handler_*        → Handlers HTTP"
echo "   • dto_*            → Data Transfer Objects"
echo "   • repo_*           → Repositórios"
echo "   • client_*         → Clientes externos"
echo "   • middleware_*     → Middlewares HTTP"
echo "   • validation_*     → Validadores"
echo "   • interface_*      → Interfaces"
echo "   • entity_*         → Entidades"
echo "   • valueobject_*    → Value Objects"
echo "   • event_*          → Eventos"
echo "   • error_*          → Definições de erro"
echo "   • config_*         → Configurações"
echo "   • helper_*         → Funções auxiliares"
echo "   • processor_*      → Processadores"
echo "   • mapper_*         → Mapeadores"
echo "   • limiter_*        → Limitadores"
echo "   • cache_*          → Implementações de cache"
echo ""
echo "✅ Estrutura do projeto agora segue padrão consistente!"
echo "=================================================================="
