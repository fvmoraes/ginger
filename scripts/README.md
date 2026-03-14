# Ginger Scripts

Scripts para desenvolvimento e release do Ginger Framework.

---

## 📜 Scripts Disponíveis

### 🔨 `build.sh` - Build de Desenvolvimento

**Quando usar:**
- Testando mudanças localmente
- Iterações rápidas durante desenvolvimento
- Build para plataforma específica
- Não quer criar uma release oficial

**Uso:**
```bash
./scripts/build.sh              # Build todas as plataformas
./scripts/build.sh local        # Build apenas para seu OS
./scripts/build.sh linux        # Build apenas Linux
./scripts/build.sh darwin       # Build apenas macOS
./scripts/build.sh windows      # Build apenas Windows
```

**O que faz:**
- ✅ Compila binários
- ✅ Coloca em `bin/`
- ❌ NÃO atualiza versão
- ❌ NÃO cria tag git
- ❌ NÃO faz push

---

### 🚀 `release.sh` - Release Oficial

**Quando usar:**
- Criar uma release oficial
- Publicar nova versão
- Atualizar CHANGELOG
- Fazer deploy

**Uso:**
```bash
./scripts/release.sh <version> <type> [message]
```

**Tipos:**
- `major` - Mudanças breaking (1.0.0 → 2.0.0)
- `minor` - Novas features (1.1.0 → 1.2.0)
- `patch` - Bug fixes (1.1.1 → 1.1.2)

**Exemplos:**
```bash
./scripts/release.sh 1.2.0 minor "Add WebSocket support"
./scripts/release.sh 1.1.5 patch "Fix CORS middleware"
./scripts/release.sh 2.0.0 major "Complete rewrite"
```

**O que faz:**
- ✅ Atualiza versão no README.md
- ✅ Atualiza CHANGELOG.md
- ✅ Compila binários (5 plataformas)
- ✅ Gera checksums SHA256
- ✅ Cria release notes
- ✅ Cria tag git
- ✅ Faz commit e push
- ✅ Mostra instruções para GitHub release

---

## 🎯 Workflow Recomendado

### Durante Desenvolvimento
```bash
# Fazer mudanças no código
vim pkg/router/router.go

# Testar localmente
./scripts/build.sh local
./bin/ginger version

# Commit normal (sem release)
git add .
git commit -m "feat: improve router performance"
git push
```

### Quando Pronto para Release
```bash
# Criar release oficial
./scripts/release.sh 1.2.0 minor "Improve router performance"

# Seguir instruções para criar GitHub release
# ou usar GitHub CLI:
gh release create v1.2.0 releases/v1.2.0/ginger-* \
  --title "Ginger Framework v1.2.0" \
  --notes-file releases/v1.2.0/RELEASE_NOTES.md \
  --latest
```

---

## 📝 Convenções de Versionamento

Seguimos [Semantic Versioning](https://semver.org/):

- **MAJOR** (X.0.0) - Breaking changes
  - Mudanças incompatíveis na API
  - Remoção de features
  - Mudanças na estrutura de projeto

- **MINOR** (1.X.0) - Novas features
  - Novas funcionalidades
  - Novos pacotes
  - Novas integrações
  - Melhorias significativas

- **PATCH** (1.1.X) - Bug fixes
  - Correções de bugs
  - Pequenas melhorias
  - Atualizações de documentação
  - Ajustes de performance

---

## 🔍 Exemplos Práticos

### Exemplo 1: Fix de Bug
```bash
# Corrigir bug no middleware CORS
vim pkg/middleware/middleware.go

# Build e testar
./scripts/build.sh local
./bin/ginger new test-api
cd test-api && ginger run

# Tudo OK? Criar release patch
cd ..
./scripts/release.sh 1.1.5 patch "Fix CORS preflight handling"
```

### Exemplo 2: Nova Feature
```bash
# Adicionar suporte a GraphQL
vim internal/integrations/graphql.go

# Build e testar
./scripts/build.sh local

# Criar release minor
./scripts/release.sh 1.2.0 minor "Add GraphQL integration"
```

### Exemplo 3: Breaking Change
```bash
# Refatorar API do router
vim pkg/router/router.go

# Build e testar
./scripts/build.sh all

# Criar release major
./scripts/release.sh 2.0.0 major "Redesign router API"
```

---

## 💡 Dicas

1. **Use `build.sh` frequentemente** durante desenvolvimento
2. **Use `release.sh` apenas** quando pronto para publicar
3. **Teste antes de release** com `build.sh local`
4. **Siga semantic versioning** para escolher o tipo
5. **Escreva mensagens claras** no release

---

## 🛠️ Requisitos

- Bash 4.0+
- Go 1.25+
- Git
- shasum (para checksums)
- GitHub CLI (opcional, para `gh release`)

---

## 📞 Ajuda

Se algo der errado:
- Verifique se está na branch `main`
- Verifique se não há mudanças uncommitted
- Verifique se a versão não existe já
- Veja os logs de erro para detalhes
