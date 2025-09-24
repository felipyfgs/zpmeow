#!/bin/bash

# Script para validar se os arquivos seguem a convenção [categoria]_[especificação].go
# Criado em: 24/09/2025

echo "🔍 Validando convenção de nomenclatura [categoria]_[especificação].go"
echo "=================================================================="

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Contadores
total_files=0
compliant_files=0
non_compliant_files=0

# Categorias válidas
valid_categories=(
    "service" "usecase" "handler" "dto" "repo" "client" "middleware" 
    "validation" "interface" "entity" "valueobject" "event" "error" 
    "config" "helper" "processor" "mapper" "limiter" "cache" "app"
    "router" "entity" "main"
)

# Exceções (arquivos que não precisam seguir a convenção)
exceptions=(
    "main.go"
    "docs.go"
    "constants.go"
    "errors.go"
    "types.go"
    "utils.go"
    "events.go"
    "cache.go"
    "connection.go"
    "service.go"
)

# Função para verificar se arquivo é exceção
is_exception() {
    local filename="$1"
    for exception in "${exceptions[@]}"; do
        if [[ "$filename" == "$exception" ]]; then
            return 0
        fi
    done
    return 1
}

# Função para verificar se categoria é válida
is_valid_category() {
    local category="$1"
    for valid_cat in "${valid_categories[@]}"; do
        if [[ "$category" == "$valid_cat" ]]; then
            return 0
        fi
    done
    return 1
}

# Função para validar nome do arquivo
validate_filename() {
    local filepath="$1"
    local filename=$(basename "$filepath")
    local dir=$(dirname "$filepath")
    
    total_files=$((total_files + 1))
    
    # Pula arquivos de teste
    if [[ "$filename" == *"_test.go" ]]; then
        echo -e "${BLUE}📝 TESTE:${NC} $filepath (arquivo de teste - OK)"
        compliant_files=$((compliant_files + 1))
        return 0
    fi
    
    # Verifica exceções
    if is_exception "$filename"; then
        echo -e "${BLUE}📋 EXCEÇÃO:${NC} $filepath (arquivo de exceção - OK)"
        compliant_files=$((compliant_files + 1))
        return 0
    fi
    
    # Verifica se segue o padrão [categoria]_[especificação].go
    if [[ "$filename" =~ ^([a-z]+)_([a-z]+)\.go$ ]]; then
        local category="${BASH_REMATCH[1]}"
        local specification="${BASH_REMATCH[2]}"
        
        if is_valid_category "$category"; then
            echo -e "${GREEN}✅ CONFORME:${NC} $filepath (${category}_${specification})"
            compliant_files=$((compliant_files + 1))
        else
            echo -e "${YELLOW}⚠️  CATEGORIA INVÁLIDA:${NC} $filepath (categoria: $category)"
            non_compliant_files=$((non_compliant_files + 1))
        fi
    else
        echo -e "${RED}❌ NÃO CONFORME:${NC} $filepath (não segue padrão [categoria]_[especificação].go)"
        non_compliant_files=$((non_compliant_files + 1))
    fi
}

echo ""
echo "🔍 Analisando arquivos .go no projeto..."
echo ""

# Encontra todos os arquivos .go e valida
while IFS= read -r -d '' file; do
    validate_filename "$file"
done < <(find . -name "*.go" -type f -not -path "./vendor/*" -not -path "./.git/*" -print0 | sort -z)

echo ""
echo "=================================================================="
echo "📊 RELATÓRIO DE VALIDAÇÃO:"
echo "=================================================================="
echo -e "📁 Total de arquivos analisados: ${BLUE}$total_files${NC}"
echo -e "✅ Arquivos conformes: ${GREEN}$compliant_files${NC}"
echo -e "❌ Arquivos não conformes: ${RED}$non_compliant_files${NC}"

# Calcula porcentagem
if [ $total_files -gt 0 ]; then
    compliance_percentage=$((compliant_files * 100 / total_files))
    echo -e "📈 Taxa de conformidade: ${BLUE}$compliance_percentage%${NC}"
else
    compliance_percentage=0
fi

echo ""
echo "=================================================================="

# Status final
if [ $non_compliant_files -eq 0 ]; then
    echo -e "${GREEN}🎉 PARABÉNS! Todos os arquivos seguem a convenção!${NC}"
    exit 0
elif [ $compliance_percentage -ge 80 ]; then
    echo -e "${YELLOW}⚠️  Boa conformidade, mas ainda há arquivos para ajustar.${NC}"
    exit 1
else
    echo -e "${RED}❌ Muitos arquivos não seguem a convenção. Execute o script de renomeação.${NC}"
    exit 2
fi
