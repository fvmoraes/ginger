# Fase 2 — Developer Experience

> 5 features para acelerar o ciclo de desenvolvimento e tornar o Ginger auto-suficiente.

---

## 1. Hot Reload — `ginger dev`

**Status**: Inexistente | **Prioridade**: ALTA | **Dependências**: `internal/cli`

### Motivação
Todo ciclo "edita → compila → roda → testa" custa tempo. Ferramentas como Air, CompileDaemon e nodemon existem, mas são externas.

### Design

```bash
# Inicia o app com hot reload
ginger dev

# Com flags
ginger dev --port 8080 --exclude "vendor,node_modules"
```

### Comportamento
1. Constrói e inicia o binário (`go build && ./bin/<name>`)
2. Monitora o diretório por mudanças (`.go`, `.yaml`, `.env`)
3. Ao detectar mudança, mata o processo atual e reinicia
4. Reinicia também se o processo morrer sozinho
5. Proxy opcional na porta 3000 com injeção de LiveReload para frontend

### Componentes
- `internal/dev/` — novo pacote: watcher (fsnotify wrapper), runner (start/stop/restart), proxy (opcional)
- `internal/cli/dev.go` — comando `ginger dev`

---

## 2. Migration Tooling — `ginger migrate`

**Status**: Inexistente (pasta `migrations/` vazia) | **Prioridade**: ALTA | **Dependências**: `internal/cli`

### Motivação
Hoje a pasta `migrations/` é criada vazia. O usuário precisa buscar ferramentas externas (golang-migrate, goose, atlas). O Ginger deveria oferecer o básico nativamente.

### Design

```bash
# Criar migration
ginger migrate create add_users_table
# → migrations/20260628120000_add_users_table.up.sql
# → migrations/20260628120000_add_users_table.down.sql

# Aplicar
ginger migrate up
ginger migrate up --steps 1

# Reverter
ginger migrate down
ginger migrate down --steps 1

# Status
ginger migrate status
```

### Arquitetura
- Arquivos SQL puros (sem DSL, sem Go migrations)
- Tabela `schema_migrations` para tracking (criada automaticamente)
- Driver configurável: postgres, mysql, sqlite
- Lock para evitar migrações concorrentes
- Embed das migrations no binário via `//go:embed`

---

## 3. CRUD Generator v2

**Status**: v1 existe (básico) | **Prioridade**: ALTA | **Dependências**: `pkg/validate`, `pkg/pagination`

### Motivação
O CRUD atual gera código básico sem validação, paginação, filtros ou ordenação. A versão 2 deve gerar código próximo do que um desenvolvedor experiente escreveria.

### O que muda

| Aspecto | v1 (atual) | v2 (planejado) |
|---------|-----------|-----------------|
| Validação | Manual no service | Tags `validate` no model |
| Paginação | Lista completa | Paginação cursor + offset |
| Filtros | Nenhum | Query params → filtros SQL |
| Ordenação | Nenhuma | `?sort=name&order=asc` seguro |
| Busca | Apenas GET /{id} | `?q=searchterm` com ILIKE |
| Soft delete | Não | Suporte opcional (`--soft-delete`) |
| Timestamps | Manual | Automáticos (CreatedAt/UpdatedAt/DeletedAt) |
| Testes | 1 teste básico | Handler + Service + Repository + Integration |

### Comando

```bash
ginger generate crud product \
  --fields "name:string:required,price:float64:min=0,category:string:oneof=books electronics food" \
  --soft-delete \
  --pagination cursor \
  --search name
```

---

## 4. Test Generator v2

**Status**: v1 existe | **Prioridade**: MÉDIA | **Dependências**: `pkg/testhelper`

### Melhorias

```bash
# Gerar TODOS os testes de uma vez
ginger generate test --all

# Gerar com benchmarks
ginger generate test product --benchmarks

# Gerar fuzz tests para endpoints de entrada
ginger generate test product --fuzz

# Gerar testes de contrato (API snapshot)
ginger generate test --contract
```

### O que gera

| Tipo de teste | Conteúdo |
|---------------|----------|
| Unit (handler) | Table-driven com middleware chain |
| Unit (service) | Mock repository, casos de erro |
| Unit (repository) | In-memory adapter |
| Integration | Fluxo CRUD completo com `app.New()` |
| Contract | Snapshot de respostas JSON |
| Benchmark | `BenchmarkGetUser`, `BenchmarkListUsers` |
| Fuzz | Fuzzing de campos de entrada |

---

## 5. Templates Customizáveis — `.ginger.yaml`

**Status**: Inexistente | **Prioridade**: MÉDIA | **Dependências**: `internal/scaffold`, `internal/generator`, Fase 1 completa

### Motivação
Cada empresa tem convenções diferentes (nomes de pastas, padrões de erro, libs de log, ORMs). Templates hardcoded em Go não permitem customização.

### Design

```yaml
# .ginger.yaml (na raiz do projeto ou ~/.config/ginger/)
version: "1"
project:
  module_prefix: "github.com/mycompany"
  default_type: service
  go_version: "1.25"

templates:
  handler:
    use_ginger_errors: true
    response_envelope: true
  model:
    use_uuid: true
    add_soft_delete: true
    validation: true
  service:
    interface_first: true
    use_context: true

paths:
  models: "internal/domain"
  handlers: "internal/api/handlers"
  services: "internal/core/services"
  repositories: "internal/data/repositories"
  ports: "internal/ports"
  adapters: "internal/adapters"

integrations:
  default_database: postgres
  cache: redis
  messaging: kafka
```

### Implementação
- `.ginger.yaml` no projeto > `~/.config/ginger/config.yaml` > defaults
- Templates Go usam dados do config para modificar comportamento
- Chaves desconhecidas são ignoradas (não quebram com configs antigas/novas)
- Sem suporte a templates customizados pelo usuário (evita complexidade)

### Nota pós-revisão
As chaves de config só referenciam features que já existem. Ao adicionar uma feature nova na Fase 1, sua chave de config correspondente é documentada. O parser ignora chaves desconhecidas para forward-compat.

---

## Nota: GraphQL movido para Fase 3

O plano original incluía `ginger generate graphql` na Fase 2. Após revisão crítica, foi rebaixado para template de integração (`ginger add graphql`) na Fase 3 — gerar código GraphQL de qualidade exigiria um codegen complexo (gqlgen). O template de integração gera schema + gqlgen.yml + resolver stub, delegando a geração pesada ao gqlgen.
