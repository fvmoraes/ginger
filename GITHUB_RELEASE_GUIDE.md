# GitHub Release Guide - Ginger Framework

Guia para criar os releases v1.1.2 e v1.1.3 no GitHub.

---

## 📦 Release v1.1.2

### 1. Acessar GitHub Releases
https://github.com/fvmoraes/ginger/releases/new

### 2. Configurar Release

**Tag:** `v1.1.2`  
**Target:** `main` branch  
**Title:** `Ginger Framework v1.1.2 - Bug Fix Release`

### 3. Descrição

Copie o conteúdo de: `releases/v1.1.2/RELEASE_NOTES.md`

### 4. Upload de Binários

Faça upload dos seguintes arquivos de `releases/v1.1.2/`:

- ✅ `ginger-linux-amd64`
- ✅ `ginger-linux-arm64`
- ✅ `ginger-darwin-amd64`
- ✅ `ginger-darwin-arm64`
- ✅ `ginger-windows-amd64.exe`
- ✅ `checksums.txt`

### 5. Opções

- ⬜ Set as the latest release (NÃO marcar - v1.1.3 será a latest)
- ⬜ Set as a pre-release (NÃO marcar)
- ✅ Create a discussion for this release (OPCIONAL)

### 6. Publicar

Clique em: **Publish release**

---

## 📦 Release v1.1.3 (LATEST)

### 1. Acessar GitHub Releases
https://github.com/fvmoraes/ginger/releases/new

### 2. Configurar Release

**Tag:** `v1.1.3`  
**Target:** `main` branch  
**Title:** `Ginger Framework v1.1.3 - Recommended Release`

### 3. Descrição

Copie o conteúdo de: `releases/v1.1.3/RELEASE_NOTES.md`

### 4. Upload de Binários

Faça upload dos seguintes arquivos de `releases/v1.1.3/`:

- ✅ `ginger-linux-amd64`
- ✅ `ginger-linux-arm64`
- ✅ `ginger-darwin-amd64`
- ✅ `ginger-darwin-arm64`
- ✅ `ginger-windows-amd64.exe`
- ✅ `checksums.txt`

### 5. Opções

- ✅ **Set as the latest release** (MARCAR - esta é a versão recomendada)
- ⬜ Set as a pre-release (NÃO marcar)
- ✅ Create a discussion for this release (OPCIONAL)

### 6. Publicar

Clique em: **Publish release**

---

## 📝 Ordem de Criação

**IMPORTANTE:** Crie os releases nesta ordem:

1. **Primeiro:** v1.1.2
2. **Depois:** v1.1.3 (marcando como "latest")

Isso garante que v1.1.3 apareça como a versão mais recente.

---

## ✅ Verificação Pós-Release

Após criar ambos os releases, verifique:

### 1. Página de Releases
https://github.com/fvmoraes/ginger/releases

Deve mostrar:
- ✅ v1.1.3 (Latest) - no topo
- ✅ v1.1.2 - abaixo
- ✅ v1.1.1 - ainda visível mas não recomendada

### 2. Testar Instalação

```bash
# Via go install (deve pegar v1.1.3)
go install github.com/fvmoraes/ginger/cmd/ginger@latest
ginger version
# Deve mostrar: ginger version 1.1.3

# Via script de instalação
curl -sSL https://raw.githubusercontent.com/fvmoraes/ginger/main/install.sh | bash
ginger version
# Deve mostrar: ginger version 1.1.3
```

### 3. Verificar pkg.go.dev

Acesse: https://pkg.go.dev/github.com/fvmoraes/ginger

Deve mostrar:
- ✅ v1.1.3 como versão mais recente
- ✅ v1.1.2 disponível
- ✅ v1.1.1 marcada como retracted (se já indexado)

### 4. Testar Download de Binários

```bash
# Linux AMD64
curl -L https://github.com/fvmoraes/ginger/releases/download/v1.1.3/ginger-linux-amd64 -o ginger
chmod +x ginger
./ginger version

# macOS Apple Silicon
curl -L https://github.com/fvmoraes/ginger/releases/download/v1.1.3/ginger-darwin-arm64 -o ginger
chmod +x ginger
./ginger version
```

---

## 🎯 Checklist Final

Antes de considerar completo:

- [ ] v1.1.2 release criado
- [ ] v1.1.3 release criado e marcado como "latest"
- [ ] Todos os binários (6 arquivos) em cada release
- [ ] checksums.txt em cada release
- [ ] Release notes completas em cada release
- [ ] `go install @latest` instala v1.1.3
- [ ] Script install.sh funciona
- [ ] Binários executam e mostram versão correta
- [ ] pkg.go.dev indexado (pode levar alguns minutos)

---

## 📊 Estatísticas dos Releases

### v1.1.2
- **Binários:** 5 plataformas
- **Tamanho total:** ~14.5 MB
- **Checksums:** SHA256
- **Status:** Stable (não latest)

### v1.1.3
- **Binários:** 5 plataformas
- **Tamanho total:** ~14.5 MB
- **Checksums:** SHA256
- **Status:** Latest (recomendada)

---

## 🚀 Após os Releases

### 1. Atualizar install.sh (se necessário)

O script já usa `${GINGER_VERSION:-v1.1.1}`, então deve funcionar automaticamente com v1.1.3.

Mas você pode atualizar a versão padrão:

```bash
VERSION="${GINGER_VERSION:-v1.1.3}"  # Atualizar aqui
```

### 2. Divulgar

- Reddit: r/golang
- Twitter/X
- Dev.to
- LinkedIn
- Hacker News
- Go Forum

### 3. Monitorar

- Issues no GitHub
- Discussions
- Downloads dos releases
- Estatísticas do pkg.go.dev

---

## 💡 Dicas

1. **Ordem importa:** Sempre crie releases mais antigos primeiro, depois os mais novos
2. **Latest badge:** Apenas uma release pode ser "latest"
3. **Binários:** Teste pelo menos um binário de cada release antes de publicar
4. **Checksums:** Sempre inclua para segurança
5. **Release notes:** Seja claro sobre o que mudou

---

## 📞 Suporte

Se algo der errado:
- Releases podem ser editados depois de criados
- Binários podem ser adicionados/removidos
- Descrição pode ser atualizada
- Tag "latest" pode ser movida

---

**Boa sorte com os releases! 🎉**
