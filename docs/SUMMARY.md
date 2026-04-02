# Sumário da Documentação Ginger

```
📚 Documentação Ginger Framework
│
├── 📖 README.md (Índice Geral)
│   └── Visão geral de toda a documentação
│       Fluxo de aprendizado recomendado
│       Busca rápida por funcionalidade
│       Dicas e truques
│
├── 🏗️ ARCHITECTURE.md
│   ├── Visão Geral
│   │   ├── Diagrama de componentes
│   │   └── Três pilares fundamentais
│   ├── Filosofia de Design
│   │   ├── Separação de responsabilidades
│   │   ├── Dependency injection manual
│   │   ├── Interfaces no consumidor
│   │   └── Erros tipados
│   ├── Estrutura de Diretórios
│   │   ├── Layout completo
│   │   └── Convenções de nomenclatura
│   ├── Fluxo de Requisição
│   │   ├── Ciclo de vida completo (9 etapas)
│   │   └── Exemplo concreto: POST /api/v1/users
│   ├── Camadas da Aplicação
│   │   ├── Handler Layer (HTTP I/O)
│   │   ├── Service Layer (Lógica de negócio)
│   │   └── Repository Layer (Acesso a dados)
│   └── Padrões de Código
│       ├── Constructor pattern
│       ├── Interface segregation
│       ├── Error wrapping
│       ├── Context propagation
│       └── Table-driven tests
│
├── 📦 PACKAGES.md
│   ├── pkg/app
│   │   ├── API completa
│   │   ├── Lifecycle hooks
│   │   └── Graceful shutdown
│   ├── pkg/router
│   │   ├── Registro de rotas
│   │   ├── Grupos de rotas
│   │   ├── Path parameters
│   │   └── JSON helpers (JSON, Error, Decode)
│   ├── pkg/middleware
│   │   ├── Logger
│   │   ├── Recover
│   │   ├── RequestID
│   │   ├── CORS (com CORSConfig avançado)
│   │   ├── Chain
│   │   └── Middleware customizado
│   ├── pkg/errors
│   │   ├── Códigos de erro
│   │   ├── Construtores
│   │   ├── Mapeamento HTTP
│   │   ├── Error wrapping
│   │   └── Erros customizados
│   ├── pkg/response
│   │   ├── OK (200)
│   │   ├── Created (201)
│   │   ├── Paginated (200 + pagination)
│   │   ├── NoContent (204)
│   │   └── Frontend integration (TypeScript)
│   ├── pkg/sse
│   │   ├── API completa
│   │   ├── Exemplo servidor
│   │   ├── Frontend (JavaScript)
│   │   ├── Casos de uso
│   │   └── Nginx configuration
│   ├── pkg/ws
│   │   ├── API completa
│   │   ├── Exemplo servidor
│   │   ├── Frontend (JavaScript)
│   │   └── Broadcast pattern
│   └── [+ 6 outros pacotes documentados]
│
├── 🔌 INTEGRATIONS.md
│   ├── Visão Geral
│   │   └── Tabela completa de integrações
│   ├── Bancos de Dados
│   │   ├── PostgreSQL (DSN, config, health check)
│   │   ├── MySQL
│   │   ├── SQLite
│   │   └── SQL Server
│   ├── Cache
│   │   └── Redis (comandos comuns, health check)
│   ├── NoSQL
│   │   ├── MongoDB (CRUD completo)
│   │   ├── Couchbase (N1QL queries)
│   │   └── ClickHouse (analytical)
│   ├── Mensageria
│   │   ├── Kafka (producer + consumer)
│   │   ├── RabbitMQ (publish + consume)
│   │   ├── NATS (pub/sub)
│   │   └── Google Pub/Sub
│   ├── Protocolos
│   │   ├── gRPC (server + client + health)
│   │   └── MCP (Model Context Protocol)
│   ├── Observabilidade
│   │   ├── OpenTelemetry (setup + tracing)
│   │   └── Prometheus (metrics + custom)
│   └── Real-time
│       ├── SSE (handler example)
│       └── WebSocket (handler example)
│
├── 🧪 TESTING.md
│   ├── Filosofia de Testes
│   │   ├── Pirâmide de testes
│   │   └── 5 princípios
│   ├── Estrutura de Testes
│   │   ├── Convenções de nomenclatura
│   │   └── Padrão de nome de teste
│   ├── Testes Unitários
│   │   ├── Handler tests
│   │   ├── Service tests
│   │   └── Table-driven tests
│   ├── Testes de Integração
│   │   ├── Database integration
│   │   └── API integration
│   ├── Mocks e Stubs
│   │   ├── Manual mocks (recomendado)
│   │   └── Testify/mock (opcional)
│   ├── Test Helpers
│   │   ├── pkg/testhelper
│   │   └── Custom helpers
│   ├── Coverage
│   │   ├── Comandos
│   │   ├── Coverage por pacote
│   │   └── Makefile targets
│   ├── CI/CD
│   │   ├── GitHub Actions
│   │   └── GitLab CI
│   └── Boas Práticas
│       ├── Teste comportamento
│       ├── Use t.Helper()
│       ├── Cleanup com t.Cleanup()
│       ├── Parallel tests
│       └── Skip slow tests
│
└── 🚀 DEPLOYMENT.md
    ├── Docker
    │   ├── Dockerfile gerado
    │   ├── Build e run
    │   └── Otimizações (cache, multi-platform, distroless)
    ├── Docker Compose
    │   ├── docker-compose.yml gerado
    │   └── Comandos úteis
    ├── Kubernetes
    │   ├── Deployment YAML
    │   ├── Secrets
    │   ├── ConfigMap
    │   └── Deploy commands
    ├── Helm Charts
    │   ├── Chart.yaml
    │   ├── values.yaml
    │   ├── Deployment template
    │   ├── Helm commands
    │   └── Ambientes múltiplos
    ├── CI/CD
    │   ├── GitHub Actions (build + deploy)
    │   └── GitLab CI (stages)
    ├── Ambientes
    │   ├── Configuração por ambiente
    │   └── Variáveis de ambiente
    ├── Monitoramento
    │   ├── Health checks
    │   ├── Prometheus metrics
    │   └── OpenTelemetry
    └── Troubleshooting
        ├── Logs
        ├── Debug container
        ├── Port forward
        └── Common issues

```

## Estatísticas

- **Escopo:** documentação ampla dos fluxos principais, pacotes centrais e operação em produção
- **Idiomas:** Português (primário) + Inglês (README principal)
- **Manutenção:** exemplos e referências devem acompanhar a CLI e os templates atuais

## Estrutura de Links

Todos os documentos estão interligados:

```
README.md (principal)
    ↓
    ├─→ docs/README.md (índice)
    │       ↓
    │       ├─→ docs/ARCHITECTURE.md
    │       ├─→ docs/PACKAGES.md
    │       ├─→ docs/INTEGRATIONS.md
    │       ├─→ docs/TESTING.md
    │       └─→ docs/DEPLOYMENT.md
    │
    └─→ Cada documento tem:
        ├─ Link de volta ao README principal
        ├─ Links para outros documentos relacionados
        └─ Índice interno com âncoras

```

## Navegação Recomendada

### Para Iniciantes
```
README.md → docs/ARCHITECTURE.md → docs/PACKAGES.md
```

### Para Desenvolvedores
```
docs/PACKAGES.md ⇄ docs/INTEGRATIONS.md ⇄ docs/TESTING.md
```

### Para DevOps
```
docs/DEPLOYMENT.md → docs/INTEGRATIONS.md (Observability)
```

## Conteúdo por Documento

| Documento | Foco | Público |
|-----------|------|---------|
| **README.md** (índice) | Navegação e busca | Todos |
| **ARCHITECTURE.md** | Estrutura e padrões | Desenvolvedores |
| **PACKAGES.md** | API reference | Desenvolvedores |
| **INTEGRATIONS.md** | Bancos, cache, mensageria | Desenvolvedores |
| **TESTING.md** | Testes e qualidade | Desenvolvedores |
| **DEPLOYMENT.md** | Deploy e produção | DevOps |

## Recursos Especiais

### Diagramas ASCII
- Arquitetura de componentes
- Fluxo de requisição
- Pirâmide de testes

### Exemplos de Código
- Mais de 150 exemplos práticos
- Código completo e executável
- Comentários explicativos

### Tabelas de Referência
- Mapeamento de erros → HTTP status
- Integrações disponíveis
- Comandos CLI
- Variáveis de ambiente

### Snippets Frontend
- JavaScript (SSE, WebSocket)
- TypeScript (tipos para envelopes)
- Configuração Nginx

### Configurações Prontas
- GitHub Actions workflows
- GitLab CI pipelines
- Kubernetes manifests
- Helm charts
- Docker Compose

---

<div align="center">
  <p><strong>Documentação completa e profunda do Ginger Framework</strong></p>
  <p>Criada com atenção aos detalhes e foco na experiência do desenvolvedor</p>
  <p><a href="./README.md">← Voltar ao Índice</a></p>
</div>
