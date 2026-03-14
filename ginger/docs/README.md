# Documentação Ginger Framework

[← Voltar ao README Principal](../README.md)

Bem-vindo à documentação completa do Ginger Framework. Esta documentação cobre todos os aspectos do framework, desde conceitos básicos até deploy em produção.

---

## 📖 Guias Principais

### 🚀 [Guia de Início Rápido](./GETTING_STARTED.md)
**Comece em 5 minutos**

Tutorial prático e direto:
- Instalar o Ginger
- Criar primeiro projeto
- Gerar CRUD completo
- Adicionar banco de dados
- Deploy com Docker
- Exemplos de código prontos
- Dúvidas comuns respondidas

**Ideal para:** Quem quer começar imediatamente.

---

### 🏗️ [Guia de Arquitetura](./ARCHITECTURE.md)
**Entenda a estrutura e filosofia do Ginger**

Mergulho profundo na arquitetura do framework:
- Visão geral e diagrama de componentes
- Filosofia de design (stdlib-first, opinativo mas flexível, zero mágica)
- Estrutura de diretórios completa com convenções
- Fluxo de requisição detalhado (Router → Middleware → Handler → Service → Repository)
- Camadas da aplicação e suas responsabilidades
- Padrões de código (constructors, interfaces, error wrapping, context propagation)

**Ideal para:** Novos desenvolvedores que querem entender como o Ginger funciona internamente.

---

### 📦 [Referência de Pacotes](./PACKAGES.md)
**API completa de todos os pacotes core**

Documentação detalhada com exemplos práticos:

#### Pacotes HTTP
- **pkg/app** — Bootstrap da aplicação com lifecycle management
- **pkg/router** — Roteamento HTTP, grupos, path params, JSON helpers
- **pkg/middleware** — Logger, Recover, RequestID, CORS (com CORSConfig avançado)
- **pkg/response** — Envelopes JSON padronizados (OK, Created, Paginated, NoContent)

#### Pacotes Real-time
- **pkg/sse** — Server-Sent Events para streaming unidirecional
- **pkg/ws** — WebSocket para comunicação bidirecional (RFC 6455, zero deps)

#### Pacotes Core
- **pkg/errors** — Erros tipados com códigos HTTP (BadRequest, NotFound, Conflict, etc.)
- **pkg/config** — Carregamento de YAML + override por env vars
- **pkg/logger** — Logging estruturado com log/slog
- **pkg/database** — Conexão e health check para SQL
- **pkg/health** — Health checks concorrentes com interface Checker
- **pkg/telemetry** — OpenTelemetry com OTLP e stdout exporters
- **pkg/testhelper** — Utilitários para testes HTTP

**Ideal para:** Referência rápida durante desenvolvimento.

---

### 🔌 [Guia de Integrações](./INTEGRATIONS.md)
**Como adicionar bancos, cache, mensageria e mais**

Guia completo do comando `ginger add <integration>`:

#### Bancos de Dados
- **PostgreSQL** — `github.com/lib/pq`
- **MySQL** — `github.com/go-sql-driver/mysql`
- **SQLite** — `github.com/mattn/go-sqlite3`
- **SQL Server** — `github.com/microsoft/go-mssqldb`
- **ClickHouse** — `github.com/ClickHouse/clickhouse-go/v2` (analytical)

#### NoSQL
- **MongoDB** — `go.mongodb.org/mongo-driver`
- **Couchbase** — `github.com/couchbase/gocb/v2`

#### Cache
- **Redis** — `github.com/redis/go-redis/v9`

#### Mensageria
- **Kafka** — `github.com/segmentio/kafka-go`
- **RabbitMQ** — `github.com/rabbitmq/amqp091-go`
- **NATS** — `github.com/nats-io/nats.go`
- **Google Pub/Sub** — `cloud.google.com/go/pubsub`

#### Protocolos
- **gRPC** — `google.golang.org/grpc` (server + client + health check)
- **MCP** — Model Context Protocol (stdlib only)

#### Real-time
- **SSE** — Server-Sent Events (stdlib only)
- **WebSocket** — Bidirectional communication (stdlib only)

#### Observabilidade
- **OpenTelemetry** — `go.opentelemetry.io/otel`
- **Prometheus** — `github.com/prometheus/client_golang`

Cada integração inclui:
- Comando de instalação
- Exemplo de uso completo
- Configuração de health check
- Variáveis de ambiente
- Comandos comuns

**Ideal para:** Adicionar novas funcionalidades ao projeto.

---

### 🧪 [Guia de Testes](./TESTING.md)
**Estratégias de teste e melhores práticas**

Cobertura completa de testes:

#### Filosofia
- Pirâmide de testes (unitários → integração → E2E)
- Princípios (rápidos, isolados, determinísticos, legíveis, mantíveis)

#### Tipos de Teste
- **Testes Unitários** — Handler, Service, Repository
- **Testes de Integração** — Database, API completa
- **Table-Driven Tests** — Parametrização com subtests

#### Mocks e Stubs
- Mocks manuais (recomendado)
- Testify/mock (opcional)
- Padrões de interface segregation

#### Test Helpers
- `pkg/testhelper` — NewRequest, AssertStatus, DecodeJSON
- Custom helpers — NewTestDB, SeedUsers

#### Coverage e CI/CD
- Comandos de coverage
- Makefile targets
- GitHub Actions
- GitLab CI

#### Boas Práticas
- Teste comportamento, não implementação
- Use `t.Helper()` em helpers
- Cleanup com `t.Cleanup()`
- Parallel tests com `t.Parallel()`
- Skip slow tests com `testing.Short()`

**Ideal para:** Garantir qualidade e confiabilidade do código.

---

### 🚀 [Guia de Deploy](./DEPLOYMENT.md)
**Produção com Docker, Kubernetes e Helm**

Deploy completo do desenvolvimento à produção:

#### Docker
- Dockerfile multi-stage gerado automaticamente
- Otimizações (build cache, multi-platform, distroless)
- Build e run local

#### Docker Compose
- docker-compose.yml com app + postgres + redis
- Health checks
- Volumes persistentes
- Comandos úteis

#### Kubernetes
- Deployment + Service manifests
- Secrets e ConfigMaps
- Probes (liveness, readiness)
- Resource limits
- Comandos kubectl

#### Helm Charts
- Chart.yaml e values.yaml gerados
- Templates parametrizados
- Multi-ambiente (dev, staging, prod)
- Autoscaling
- Ingress

#### CI/CD
- GitHub Actions (build + deploy)
- GitLab CI (stages + environments)
- Container registry integration
- Automated rollouts

#### Monitoramento
- Health checks
- Prometheus metrics
- OpenTelemetry traces
- ServiceMonitor (Prometheus Operator)

#### Troubleshooting
- Logs (Docker, Kubernetes)
- Debug containers
- Port forwarding
- Common issues (CrashLoopBackOff, ImagePullBackOff, Probe failures)

**Ideal para:** Levar aplicação para produção com confiança.

---

## 🎯 Fluxo de Aprendizado Recomendado

### 1. Iniciante
1. Leia o [README principal](../README.md) — visão geral
2. Siga o [Guia de Início Rápido](./GETTING_STARTED.md) — tutorial prático
3. Explore o [Guia de Arquitetura](./ARCHITECTURE.md) — entenda a estrutura
4. Consulte a [Referência de Pacotes](./PACKAGES.md) conforme necessário

### 2. Intermediário
1. Adicione integrações com o [Guia de Integrações](./INTEGRATIONS.md)
2. Implemente testes seguindo o [Guia de Testes](./TESTING.md)
3. Gere código com `ginger generate crud <resource>`
4. Use `ginger doctor` para validar o projeto

### 3. Avançado
1. Deploy local com Docker Compose
2. Deploy em Kubernetes seguindo o [Guia de Deploy](./DEPLOYMENT.md)
3. Configure CI/CD
4. Adicione observabilidade (OpenTelemetry + Prometheus)
5. Customize middlewares e integrações

---

## 🔍 Busca Rápida

### Por Funcionalidade

| Funcionalidade | Documento | Seção |
|----------------|-----------|-------|
| Criar novo projeto | README | Getting Started |
| Estrutura de pastas | ARCHITECTURE | Estrutura de Diretórios |
| Roteamento HTTP | PACKAGES | pkg/router |
| Middlewares | PACKAGES | pkg/middleware |
| Erros tipados | PACKAGES | pkg/errors |
| JSON responses | PACKAGES | pkg/response |
| Server-Sent Events | PACKAGES | pkg/sse |
| WebSocket | PACKAGES | pkg/ws |
| Adicionar banco | INTEGRATIONS | Bancos de Dados |
| Adicionar cache | INTEGRATIONS | Cache |
| Adicionar mensageria | INTEGRATIONS | Mensageria |
| Testes unitários | TESTING | Testes Unitários |
| Testes de integração | TESTING | Testes de Integração |
| Mocks | TESTING | Mocks e Stubs |
| Docker | DEPLOYMENT | Docker |
| Kubernetes | DEPLOYMENT | Kubernetes |
| Helm | DEPLOYMENT | Helm Charts |
| CI/CD | DEPLOYMENT | CI/CD |

### Por Comando CLI

| Comando | Descrição | Documento |
|---------|-----------|-----------|
| `ginger new <name>` | Criar projeto | README |
| `ginger run` | Executar app | README |
| `ginger build` | Compilar binário | README |
| `ginger generate handler <name>` | Gerar handler | README, ARCHITECTURE |
| `ginger generate service <name>` | Gerar service | README, ARCHITECTURE |
| `ginger generate repository <name>` | Gerar repository | README, ARCHITECTURE |
| `ginger generate crud <name>` | Gerar CRUD completo | README, ARCHITECTURE |
| `ginger add <integration>` | Adicionar integração | INTEGRATIONS |
| `ginger doctor` | Diagnosticar projeto | README |

---

## 💡 Dicas e Truques

### Performance
- Use `middleware.Chain()` para compor middlewares eficientemente
- Configure `MaxOpen` e `MaxIdle` no pool de conexões do banco
- Use `context.Context` para cancelamento e timeouts
- Implemente caching com Redis para queries frequentes

### Segurança
- Sempre valide input no service layer
- Use `apperrors.Internal()` para não vazar detalhes de erros internos
- Configure CORS adequadamente com `CORSConfig`
- Use secrets do Kubernetes para credenciais sensíveis

### Observabilidade
- Adicione `X-Request-ID` com `middleware.RequestID()`
- Use structured logging com `pkg/logger`
- Implemente health checks para todas as dependências
- Configure OpenTelemetry para traces distribuídos

### Desenvolvimento
- Use `ginger doctor` regularmente para validar o projeto
- Rode `go vet ./...` antes de commit
- Mantenha coverage > 70%
- Use table-driven tests para múltiplos cenários

---

## 🤝 Contribuindo

Encontrou um erro na documentação? Quer adicionar exemplos?

1. Fork o repositório
2. Crie uma branch: `git checkout -b docs/minha-melhoria`
3. Faça suas alterações
4. Commit: `git commit -m "docs: adiciona exemplo de X"`
5. Push: `git push origin docs/minha-melhoria`
6. Abra um Pull Request

---

## 📞 Suporte

- **Issues:** [GitHub Issues](https://github.com/ginger-framework/ginger/issues)
- **Discussions:** [GitHub Discussions](https://github.com/ginger-framework/ginger/discussions)
- **Email:** support@ginger-framework.dev

---

<div align="center">
  <p>Documentação mantida com ❤️ pela comunidade Ginger</p>
  <p><a href="../README.md">← Voltar ao README Principal</a></p>
</div>
