# 📋 Convenção de Nomenclatura de Arquivos

**Versão**: 1.0  
**Data**: 24/09/2025  
**Status**: ✅ **Ativo**

---

## 🎯 **PADRÃO OFICIAL**

### **Formato:**
```
[categoria]_[especificação].go
```

### **Regras:**
- ✅ **Letras minúsculas** apenas
- ✅ **Underscore** como separador
- ✅ **Nomes descritivos** e claros
- ✅ **Categoria** indica o tipo/propósito
- ✅ **Especificação** indica a funcionalidade específica

---

## 📚 **CATEGORIAS OFICIAIS**

### **🔧 Serviços e Lógica de Negócio:**
- `service_*` - Serviços de negócio
- `usecase_*` - Casos de uso da aplicação
- `manager_*` - Gerenciadores de recursos

### **🌐 HTTP e API:**
- `handler_*` - Handlers HTTP
- `dto_*` - Data Transfer Objects
- `middleware_*` - Middlewares HTTP
- `router_*` - Roteadores

### **💾 Dados e Persistência:**
- `repo_*` - Repositórios
- `entity_*` - Entidades de domínio
- `valueobject_*` - Value Objects
- `model_*` - Modelos de dados

### **🔌 Integrações e Clientes:**
- `client_*` - Clientes externos
- `adapter_*` - Adaptadores
- `integration_*` - Integrações

### **⚙️ Infraestrutura:**
- `config_*` - Configurações
- `cache_*` - Implementações de cache
- `helper_*` - Funções auxiliares
- `processor_*` - Processadores
- `mapper_*` - Mapeadores
- `limiter_*` - Limitadores

### **🔍 Validação e Controle:**
- `validation_*` - Validadores
- `interface_*` - Interfaces
- `event_*` - Eventos
- `error_*` - Definições de erro

---

## 📖 **EXEMPLOS PRÁTICOS**

### **✅ Correto:**
```
service_messages.go      → Serviço de mensagens
handler_session.go       → Handler de sessão
dto_contact.go          → DTO de contato
repo_chat.go            → Repositório de chat
client_whatsapp.go      → Cliente WhatsApp
validation_message.go   → Validação de mensagem
usecase_connect.go      → Caso de uso conectar
middleware_auth.go      → Middleware de autenticação
```

### **❌ Incorreto:**
```
messages.go             → Sem categoria
sessionHandler.go       → CamelCase
contact_dto.go          → Ordem invertida
chatRepository.go       → CamelCase
whatsapp-client.go      → Hífen em vez de underscore
messageValidation.go    → CamelCase
```

---

## 🗂️ **ESTRUTURA POR MÓDULO**

### **Application Layer:**
```
internal/application/
├── app_main.go
├── common/
│   └── error_application.go
├── ports/
│   ├── interface_events.go
│   └── interface_ports.go
└── usecases/
    ├── chat/
    │   ├── usecase_history.go
    │   ├── usecase_list.go
    │   └── usecase_manage.go
    └── session/
        ├── usecase_connect.go
        ├── usecase_create.go
        └── usecase_status.go
```

### **Infrastructure Layer:**
```
internal/infra/
├── http/
│   ├── dto/
│   │   ├── dto_chat.go
│   │   ├── dto_session.go
│   │   └── dto_message.go
│   ├── handlers/
│   │   ├── handler_chat.go
│   │   ├── handler_session.go
│   │   └── handler_message.go
│   └── middleware/
│       ├── middleware_auth.go
│       ├── middleware_cors.go
│       └── middleware_logging.go
├── database/
│   └── repository/
│       ├── repo_chat.go
│       ├── repo_session.go
│       └── repo_message.go
└── wmeow/
    ├── service_messages.go
    ├── service_sessions.go
    ├── client_wameow.go
    └── validation_message.go
```

---

## 🛠️ **FERRAMENTAS**

### **Scripts Disponíveis:**

#### **1. Script de Renomeação:**
```bash
./rename_files_script.sh
```
- Renomeia todos os arquivos para seguir a convenção
- Backup automático antes das mudanças
- Log detalhado das operações

#### **2. Script de Validação:**
```bash
./validate_naming_convention.sh
```
- Verifica conformidade com a convenção
- Relatório detalhado de status
- Taxa de conformidade em porcentagem

### **Uso Recomendado:**
```bash
# 1. Validar estado atual
./validate_naming_convention.sh

# 2. Aplicar renomeação (se necessário)
./rename_files_script.sh

# 3. Validar novamente
./validate_naming_convention.sh
```

---

## 🚫 **EXCEÇÕES**

### **Arquivos que NÃO seguem a convenção:**
- `main.go` - Ponto de entrada da aplicação
- `docs.go` - Documentação gerada
- `constants.go` - Constantes globais
- `errors.go` - Erros globais
- `types.go` - Tipos globais
- `utils.go` - Utilitários globais
- `*_test.go` - Arquivos de teste

### **Justificativa:**
Estes arquivos têm nomes convencionais estabelecidos na comunidade Go e são facilmente reconhecíveis.

---

## ✅ **BENEFÍCIOS**

### **Organização:**
- 🎯 **Agrupamento visual** de arquivos relacionados
- 🎯 **Navegação intuitiva** por funcionalidade
- 🎯 **Estrutura previsível** e consistente

### **Manutenibilidade:**
- 🔧 **Localização rápida** de código específico
- 🔧 **Propósito claro** de cada arquivo
- 🔧 **Refatoração facilitada**

### **Colaboração:**
- 👥 **Onboarding acelerado** para novos desenvolvedores
- 👥 **Code review** mais eficiente
- 👥 **Padrão uniforme** em toda a equipe

---

## 📋 **CHECKLIST DE CONFORMIDADE**

### **Antes de Criar Novo Arquivo:**
- [ ] Nome segue padrão `[categoria]_[especificação].go`
- [ ] Categoria está na lista oficial
- [ ] Nome é descritivo e claro
- [ ] Não há conflito com arquivos existentes
- [ ] Arquivo está no diretório correto

### **Durante Code Review:**
- [ ] Novos arquivos seguem a convenção
- [ ] Renomeações mantêm consistência
- [ ] Imports foram atualizados corretamente
- [ ] Testes continuam funcionando

---

## 🔄 **VERSIONAMENTO**

### **v1.0 (24/09/2025):**
- ✅ Convenção inicial estabelecida
- ✅ Scripts de automação criados
- ✅ Documentação completa
- ✅ Aplicação no módulo wmeow

### **Próximas Versões:**
- 🔮 Extensão para outros tipos de arquivo
- 🔮 Integração com CI/CD
- 🔮 Validação automática em PRs

---

**Mantido por**: Equipe de Desenvolvimento  
**Última atualização**: 24/09/2025
