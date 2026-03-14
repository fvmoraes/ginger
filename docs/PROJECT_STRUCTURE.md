# Estrutura do Projeto Ginger

[← Voltar ao Índice](./README.md)

Estrutura completa e organizada do framework Ginger.

---

## 📁 Estrutura Raiz

```
ginger/                          # Raiz do projeto
├── .git/                        # Controle de versão
├── .gitignore                   # Arquivos ignorados
├── README.md                    # Documentação principal
├── go.mod                       # Módulo Go
├── go.sum                       # Checksums de dependências
│
├── cmd/                         # Aplicações executáveis
│   └── ginger/                  # CLI do Ginger
│       └── main.go
│
├── internal/                    # Código privado do framework
│   ├── cli/                     # Comandos da CLI
│   │   ├── root.go
│   │   ├── new.go
│   │   ├── generate.go
│   │   ├── add.go
│   │   ├── run.go
│   │   └── doctor.go
│   ├── generator/               # Gerador de código
│   │   ├── generator.go
│   │   └── templates.go
│   ├── scaffold/                # Scaffold de projetos
│   │   ├── scaffold.go
│   │   └── templates.go
│   ├── integrations/            # Sistema de integrações
│   │   ├── integrations.go
│   │   └── templates.go
│   └── doctor/                  # Diagnóstico de projetos
│       └── doctor.go
│
├── pkg/                         # Código público reutilizável
│   ├── app/                     # Bootstrap da aplicação
│   │   └── app.go
│   ├── router/                  # Roteamento HTTP
│   │   └── router.go
│   ├── middleware/              # Middlewares HTTP
│   │   ├── middleware.go
│   │   └── context.go
│   ├── errors/                  # Erros tipados
│   │   └── errors.go
│   ├── response/                # Envelopes JSON
│   │   └── response.go
│   ├── config/                  # Configuração
│   │   └── config.go
│   ├── logger/                  # Logging estruturado
│   │   └── logger.go
│   ├── database/                # Conexão de banco
│   │   └── database.go
│   ├── health/                  # Health checks
│   │   └── health.go
│   ├── telemetry/               # OpenTelemetry
│   │   └── telemetry.go
│   ├── sse/                     # Server-Sent Events
│   │   └── sse.go
│   ├── ws/                      # WebSocket
│   │   ├── ws.go
│   │   └── frame.go
│   └── testhelper/              # Utilitários de teste
│       └── testhelper.go
│
├── docs/                        # Documentação completa
│   ├── README.md                # Índice da documentação
│   ├── GETTING_STARTED.md       # Tutorial de 5 minutos
│   ├── COPY_PASTE.md            # Código pronto
│   ├── ARCHITECTURE.md          # Arquitetura
│   ├── PACKAGES.md              # Referência de pacotes
│   ├── INTEGRATIONS.md          # Guia de integrações
│   ├── TESTING.md               # Guia de testes
│   ├── DEPLOYMENT.md            # Guia de deploy
│   ├── QUICK_REFERENCE.md       # Referência rápida
│   ├── QUALITY_CHECKLIST.md     # Checklist de qualidade
│   ├── VALIDATION_REPORT.md     # Relatório de validação
│   ├── SUMMARY.md               # Sumário visual
│   └── CHANGELOG.md             # Histórico
│
├── templates/                   # Templates de projeto
│   ├── k8s/                     # Kubernetes
│   │   └── deployment.yaml
│   └── project/                 # Templates de scaffold
│
├── example/                     # Projeto de exemplo
│   ├── cmd/app/main.go          # Entrypoint
│   ├── internal/                # Código da aplicação
│   │   ├── api/
│   │   │   ├── handlers/
│   │   │   ├── services/
│   │   │   └── repositories/
│   │   └── models/
│   ├── configs/app.yaml         # Configuração
│   ├── go.mod                   # Módulo separado
│   └── go.sum
│
├── bin/                         # Binários compilados
│   └── ginger                   # CLI compilada
│
└── my-local/                    # Arquivos locais (gitignored)
    ├── análises/
    ├── livros/
    └── documentação/
```

---

## 📦 Módulos Go

### Módulo Principal

```
module github.com/fvmoraes/ginger
go 1.25.0
```

**Localização:** `go.mod` na raiz

**Contém:**
- CLI (`cmd/ginger`)
- Framework (`pkg/*`)
- Geradores (`internal/*`)

### Módulo de Exemplo

```
module github.com/fvmoraes/ginger/example
go 1.25.0

replace github.com/fvmoraes/ginger => ../
```

**Localização:** `example/go.mod`

**Contém:**
- Aplicação de exemplo completa
- Demonstração de uso do framework

---

## 🎯 Convenções

### Nomenclatura de Arquivos

| Tipo | Padrão | Exemplo |
|------|--------|---------|
| Handler | `<resource>_handler.go` | `user_handler.go` |
| Service | `<resource>_service.go` | `user_service.go` |
| Repository | `<resource>_repository.go` | `user_repository.go` |
| Model | `<resource>.go` | `user.go` |
| Test | `<file>_test.go` | `user_handler_test.go` |

### Nomenclatura de Pacotes

| Tipo | Padrão | Exemplo |
|------|--------|---------|
| Pacote público | `pkg/<nome>` | `pkg/router` |
| Pacote privado | `internal/<nome>` | `internal/cli` |
| Comando | `cmd/<nome>` | `cmd/ginger` |

### Imports

```go
// Stdlib primeiro
import (
    "context"
    "net/http"
    
    // Dependências externas
    "github.com/external/package"
    
    // Pacotes do Ginger
    "github.com/fvmoraes/ginger/pkg/router"
    "github.com/fvmoraes/ginger/pkg/middleware"
    
    // Pacotes locais
    "yourmodule/internal/api/handlers"
)
```

---

## 🚀 Build e Deploy

### Build da CLI

```bash
# Desenvolvimento
go build -o bin/ginger ./cmd/ginger

# Produção
CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/ginger ./cmd/ginger

# Instalar globalmente
go build -o /usr/local/bin/ginger ./cmd/ginger
```

### Build do Exemplo

```bash
cd example
go build -o bin/app ./cmd/app
```

### Estrutura de Projeto Gerado

Quando você executa `ginger new my-api`, a estrutura criada é:

```
my-api/
├── cmd/app/main.go
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   ├── services/
│   │   ├── repositories/
│   │   └── middlewares/
│   ├── models/
│   └── config/
├── pkg/
├── platform/
├── configs/app.yaml
├── scripts/
├── tests/
├── docs/
├── Dockerfile
├── docker-compose.yml
├── kubernetes/
├── helm/
├── Makefile
├── .env.example
├── .gitignore
├── go.mod
└── README.md
```

---

## 📝 Notas Importantes

### ✅ Estrutura Correta

A estrutura atual está **correta** e **otimizada**:

```
/Users/fvmoraes/Go/src/ginger/  ← Raiz do workspace
├── cmd/                         ← Direto na raiz
├── pkg/                         ← Direto na raiz
├── internal/                    ← Direto na raiz
└── docs/                        ← Direto na raiz
```

### ❌ Estrutura Antiga (Removida)

A estrutura duplicada foi **removida**:

```
/Users/fvmoraes/Go/src/ginger/
└── ginger/                      ← REMOVIDO
    ├── cmd/                     ← Movido para raiz
    ├── pkg/                     ← Movido para raiz
    └── ...
```

### 🔍 Verificação

Para verificar que tudo está correto:

```bash
# Build deve funcionar
go build ./...

# Vet deve passar
go vet ./...

# CLI deve funcionar
go build -o bin/ginger ./cmd/ginger
./bin/ginger version

# Exemplo deve compilar
cd example && go build ./...
```

---

## 🎯 Próximos Passos

- [🚀 Começar a usar](./GETTING_STARTED.md)
- [📦 Ver pacotes](./PACKAGES.md)
- [🏗️ Entender arquitetura](./ARCHITECTURE.md)

---

<div align="center">
  <p><strong>Estrutura limpa e organizada!</strong></p>
  <p><a href="./README.md">← Voltar ao Índice</a></p>
</div>
