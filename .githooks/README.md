# Git Hooks para zpmeow

Este diretório contém hooks do Git para manter a qualidade do código.

## Configuração

Para ativar os hooks, execute:

```bash
git config core.hooksPath .githooks
```

## Hooks Disponíveis

### pre-commit

Executa verificações antes de cada commit:

- **Markdownlint**: Verifica formatação de arquivos `.md`
- **Go Build**: Verifica se o código compila

### Dependências

- `markdownlint-cli`: `npm install -g markdownlint-cli`
- `make`: Para executar o build do Go

## Uso

Após configurar, os hooks serão executados automaticamente:

```bash
git add .
git commit -m "feat: nova funcionalidade"
# Os hooks serão executados automaticamente
```

## Bypass (Emergência)

Para pular os hooks em caso de emergência:

```bash
git commit --no-verify -m "fix: correção urgente"
```

**⚠️ Use com moderação!**
