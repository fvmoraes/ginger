# Checklist de Qualidade Ginger

Use este checklist para validar projetos criados com Ginger.

---

## ✅ Código

### Estrutura
- [ ] Diretórios seguem convenção (`cmd/`, `internal/`, `pkg/`, `platform/`)
- [ ] Arquivos nomeados corretamente (`*_handler.go`, `*_service.go`, `*_repository.go`)
- [ ] go.mod presente e atualizado
- [ ] .gitignore configurado

### Qualidade
- [ ] `go build ./...` passa sem erros
- [ ] `go vet ./...` passa sem warnings
- [ ] `go test ./...` passa (se testes existem)
- [ ] Funções < 50 linhas
- [ ] Complexidade ciclomática < 10
- [ ] Sem código duplicado

### Boas Práticas
- [ ] Erros sempre verificados
- [ ] Context propagado corretamente
- [ ] Interfaces definidas no consumidor
- [ ] Construtores `New<Type>` para todos os tipos
- [ ] Uso de stdlib ao invés de código customizado

---

## ✅ Arquitetura

### Camadas
- [ ] Handler apenas faz HTTP I/O
- [ ] Service contém lógica de negócio
- [ ] Repository acessa dados
- [ ] Models são structs simples

### Dependências
- [ ] Injeção manual (sem DI framework)
- [ ] Dependências injetadas via construtor
- [ ] Interfaces pequenas e focadas
- [ ] Zero dependências circulares

### Erros
- [ ] Erros de domínio são `*apperrors.AppError`
- [ ] Erros internos não vazam para cliente
- [ ] Error wrapping com `%w`
- [ ] Mensagens de erro claras

---

## ✅ HTTP

### Rotas
- [ ] Rotas agrupadas logicamente
- [ ] Path params usados corretamente
- [ ] Métodos HTTP corretos (GET, POST, PUT, DELETE)
- [ ] Versionamento de API (`/api/v1`)

### Middlewares
- [ ] Logger configurado
- [ ] Recover configurado
- [ ] RequestID configurado
- [ ] CORS configurado (se necessário)

### Responses
- [ ] Respostas consistentes (usar `response.*`)
- [ ] Status codes corretos
- [ ] Content-Type: application/json
- [ ] Erros padronizados

---

## ✅ Banco de Dados

### Conexão
- [ ] Pool configurado (MaxOpen, MaxIdle)
- [ ] Timeout configurado
- [ ] Health check implementado
- [ ] Graceful shutdown (db.Close)

### Queries
- [ ] Prepared statements ou parametrizadas
- [ ] Context propagado
- [ ] Erros tratados
- [ ] Transactions quando necessário

---

## ✅ Configuração

### Arquivo
- [ ] `configs/app.yaml` presente
- [ ] Valores sensíveis não commitados
- [ ] `.env.example` presente
- [ ] Documentação de variáveis

### Carregamento
- [ ] YAML carregado primeiro
- [ ] Env vars sobrescrevem YAML
- [ ] Validação de config obrigatória
- [ ] Defaults sensatos

---

## ✅ Testes

### Cobertura
- [ ] Handlers testados
- [ ] Services testados
- [ ] Repositories testados (se possível)
- [ ] Coverage > 70%

### Qualidade
- [ ] Testes isolados (mocks)
- [ ] Testes determinísticos
- [ ] Testes rápidos (< 10s total)
- [ ] Table-driven tests para múltiplos casos

### Organização
- [ ] Arquivos `*_test.go` ao lado do código
- [ ] Mocks em arquivos separados
- [ ] Test helpers reutilizáveis
- [ ] Nomes descritivos

---

## ✅ Observabilidade

### Logs
- [ ] Structured logging (slog)
- [ ] Níveis corretos (debug, info, warn, error)
- [ ] Contexto suficiente
- [ ] Sem logs de dados sensíveis

### Health Checks
- [ ] Endpoint `/health` implementado
- [ ] Todas as dependências verificadas
- [ ] Resposta rápida (< 1s)
- [ ] Status codes corretos (200/503)

### Métricas (opcional)
- [ ] Prometheus configurado
- [ ] Métricas de negócio expostas
- [ ] Endpoint `/metrics` protegido

### Traces (opcional)
- [ ] OpenTelemetry configurado
- [ ] Spans em operações críticas
- [ ] Context propagado

---

## ✅ Segurança

### Input
- [ ] Validação de input no service
- [ ] Sanitização quando necessário
- [ ] Limites de tamanho (request body)
- [ ] Rate limiting (se necessário)

### Autenticação (se aplicável)
- [ ] Tokens validados
- [ ] Sessões gerenciadas
- [ ] Logout implementado
- [ ] Refresh tokens (se necessário)

### Autorização (se aplicável)
- [ ] Permissões verificadas
- [ ] RBAC implementado
- [ ] Recursos protegidos

### CORS
- [ ] Origins permitidas configuradas
- [ ] Credentials apenas se necessário
- [ ] Headers permitidos mínimos

---

## ✅ Deploy

### Docker
- [ ] Dockerfile multi-stage
- [ ] Imagem < 50MB
- [ ] Non-root user
- [ ] Health check no Dockerfile

### Kubernetes
- [ ] Deployment configurado
- [ ] Service configurado
- [ ] Probes configuradas (liveness, readiness)
- [ ] Resources configurados (requests, limits)

### CI/CD
- [ ] Pipeline de build
- [ ] Pipeline de testes
- [ ] Pipeline de deploy
- [ ] Rollback automático

---

## ✅ Documentação

### Código
- [ ] Pacotes documentados (package comment)
- [ ] Funções públicas documentadas
- [ ] Exemplos em godoc
- [ ] README.md no projeto

### API
- [ ] Endpoints documentados
- [ ] Request/Response examples
- [ ] Error codes documentados
- [ ] Postman/OpenAPI (opcional)

### Deploy
- [ ] Instruções de build
- [ ] Instruções de deploy
- [ ] Variáveis de ambiente documentadas
- [ ] Troubleshooting guide

---

## 🎯 Checklist Rápido (Mínimo Viável)

Para um projeto básico funcional:

- [ ] `go build ./...` passa
- [ ] `go vet ./...` passa
- [ ] `ginger doctor` passa
- [ ] Endpoint `/health` responde 200
- [ ] Pelo menos 1 endpoint de negócio funciona
- [ ] Dockerfile presente
- [ ] README.md com instruções básicas

---

## 📊 Scoring

| Categoria | Peso | Sua Nota | Ponderado |
|-----------|------|----------|-----------|
| Código | 25% | __/10 | __ |
| Arquitetura | 20% | __/10 | __ |
| Testes | 15% | __/10 | __ |
| Segurança | 15% | __/10 | __ |
| Observabilidade | 10% | __/10 | __ |
| Deploy | 10% | __/10 | __ |
| Documentação | 5% | __/10 | __ |
| **TOTAL** | **100%** | | **__/10** |

### Interpretação

- **9-10:** Excelente — Pronto para produção
- **7-8:** Bom — Pequenos ajustes necessários
- **5-6:** Aceitável — Melhorias recomendadas
- **< 5:** Insuficiente — Revisão necessária

---

## 🚀 Comandos Úteis

```bash
# Validar código
go build ./...
go vet ./...
go test ./...

# Diagnosticar projeto
ginger doctor

# Verificar coverage
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Lint (se instalado)
golangci-lint run

# Build para produção
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/app ./cmd/app

# Testar Docker
docker build -t myapp .
docker run --rm myapp /health

# Testar Kubernetes
kubectl apply -f kubernetes/ --dry-run=client
```

---

<div align="center">
  <p><strong>Use este checklist em cada release!</strong></p>
  <p><a href="./README.md">← Voltar ao Índice</a></p>
</div>
