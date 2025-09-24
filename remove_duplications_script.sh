#!/bin/bash

# Script para remover duplicações do service.go
# Os métodos duplicados já estão nos arquivos especializados
# Criado em: 24/09/2025

echo "🔧 Removendo duplicações do service.go - mantendo apenas métodos únicos"
echo "======================================================================="

SERVICE_FILE="internal/infra/wmeow/service.go"
BACKUP_FILE="internal/infra/wmeow/service.go.backup"

# Fazer backup
echo "📋 Criando backup do service.go..."
cp "$SERVICE_FILE" "$BACKUP_FILE"

echo "🔍 Identificando métodos duplicados..."

# Métodos que estão em service_sessions.go
echo "   • StartClient, StopClient, LogoutClient, GetQRCode, PairPhone, IsClientConnected"

# Métodos que estão em service_messages.go  
echo "   • SendTextMessage, SendImageMessage, SendAudioMessage, SendVideoMessage"
echo "   • SendDocumentMessage, SendStickerMessage, SendContactMessage, SendLocationMessage"

# Métodos que estão em service_groups.go
echo "   • CreateGroup, ListGroups, GetGroupInfo, JoinGroup, LeaveGroup"
echo "   • AddParticipant, RemoveParticipant, PromoteParticipant, DemoteParticipant"

# Métodos que estão em service_contacts.go
echo "   • CheckUser, GetContacts, GetContactInfo, GetUserInfo, GetProfilePicture"

# Métodos que estão em service_chats.go
echo "   • ListChats, GetChatHistory, ArchiveChat, DeleteChat, MuteChat"

# Métodos que estão em service_actions.go
echo "   • MarkMessageRead, DeleteMessage, EditMessage, ReactToMessage"

# Métodos que estão em service_privacy.go
echo "   • SetAllPrivacySettings, GetPrivacySettings, UpdateBlocklist"

echo ""
echo "🎯 Estratégia: Manter apenas métodos de coordenação e helpers internos"
echo ""

# Listar métodos únicos que devem permanecer no service.go
echo "✅ Métodos que devem PERMANECER no service.go:"
echo "   • Construtores: NewMeowService, NewMeowServiceWithChatwoot"
echo "   • Helpers internos: getClient, getOrCreateClient, createNewClient"
echo "   • Configuração: loadSessionConfiguration, removeClient"
echo "   • Coordenação: ConnectOnStartup, SetChatwootIntegration"

echo ""
echo "❌ Métodos que devem ser REMOVIDOS (duplicados):"

# Buscar métodos duplicados
grep -n "^func (m \*MeowService)" "$SERVICE_FILE" | while read -r line; do
    method_name=$(echo "$line" | sed 's/.*func (m \*MeowService) \([^(]*\).*/\1/')
    line_num=$(echo "$line" | cut -d: -f1)
    
    # Verificar se método existe em arquivos especializados
    if grep -q "func (m \*MeowService) $method_name" internal/infra/wmeow/service_*.go 2>/dev/null; then
        echo "   • Linha $line_num: $method_name (duplicado)"
    fi
done

echo ""
echo "⚠️  ATENÇÃO: Este script apenas identifica duplicações."
echo "    A remoção deve ser feita manualmente para evitar quebrar o código."
echo ""
echo "📋 Próximos passos manuais:"
echo "   1. Revisar cada método duplicado identificado"
echo "   2. Confirmar que existe nos arquivos especializados"
echo "   3. Remover cuidadosamente do service.go"
echo "   4. Testar compilação após cada remoção"
echo "   5. Manter apenas métodos de coordenação e helpers"

echo ""
echo "🎯 Meta: Reduzir service.go de 2,975 linhas para ~500 linhas"
echo "======================================================================="
