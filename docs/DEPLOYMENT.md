# Guia de Deploy

[← Voltar ao README](../README.md)

## Índice

- [Docker](#docker)
- [Docker Compose](#docker-compose)
- [Kubernetes](#kubernetes)
- [Helm Charts](#helm-charts)
- [CI/CD](#cicd)
- [Ambientes](#ambientes)
- [Monitoramento](#monitoramento)
- [Troubleshooting](#troubleshooting)

---

## Docker

### Dockerfile Gerado

Projetos `service` e `worker` gerados pelo Ginger vêm com um `Dockerfile` multi-stage otimizado em `devops/docker/Dockerfile`:

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build (example for service project type)
COPY . .
RUN go build -o bin/foobar ./cmd/foobar

# Runtime stage
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/bin/foobar .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

ENTRYPOINT ["./foobar"]
```

### Build e Run

```bash
# Build
docker build -f devops/docker/Dockerfile -t foobar:latest .

# Run
docker run -p 8080:8080 \
  -e DATABASE_DSN="postgres://<user>:<password>@host/db" \
  -e LOG_LEVEL="info" \
  foobar:latest
```

### Otimizações

#### 1. Build Cache

```dockerfile
# Copie go.mod/go.sum primeiro para cachear dependências
COPY go.mod go.sum ./
RUN go mod download

# Depois copie o código
COPY . .
```

#### 2. Multi-Platform Build

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t foobar:latest .
```

#### 3. Distroless Image

```dockerfile
# Runtime stage com distroless (menor e mais seguro)
FROM gcr.io/distroless/static-debian11

COPY --from=builder /app/main /
COPY --from=builder /app/configs /configs

EXPOSE 8080

CMD ["/main"]
```

**Tamanho:** ~20MB vs ~50MB (alpine)

---

## Docker Compose

### docker-compose.yml Gerado

Projetos `service` e `worker` começam com o serviço principal no compose. Dependências locais entram depois, quando você roda `ginger add <integration>`.

Localização: `devops/docker/docker-compose.yml`

```yaml
version: "3.9"

services:
  foobar:
    build:
      context: ../..
      dockerfile: devops/docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      HTTP_PORT: 8080
```

### Comandos

```bash
# Start all services
docker compose -f devops/docker/docker-compose.yml up -d

# View logs
docker compose logs -f foobar

# Stop all services
docker compose down

# Rebuild and restart
docker compose -f devops/docker/docker-compose.yml up -d --build

# Access database
docker compose exec postgres psql -U user -d foobar
```

### Atualização Automática com `ginger add`

Quando o compose já existe, o Ginger também o atualiza automaticamente para integrações com infraestrutura local. Exemplos:

```bash
ginger add postgres   # adiciona serviço postgres + DATABASE_DSN
ginger add redis      # adiciona serviço redis + REDIS_ADDR
ginger add prometheus # adiciona serviço prometheus + prometheus.yml
ginger add rabbitmq   # adiciona serviço rabbitmq + RABBITMQ_URL
ginger add kafka      # adiciona serviço kafka + KAFKA_BROKERS
ginger add nats       # adiciona serviço nats + NATS_URL
```

Hoje, as integrações que também alimentam o compose são:

- `postgres`
- `mysql`
- `redis`
- `rabbitmq`
- `kafka`
- `nats`
- `mongodb`
- `clickhouse`
- `couchbase`
- `prometheus`
- `otel`

Estas integrações não alteram o compose automaticamente:

- `sqlite`
- `sqlserver`
- `pubsub`
- `grpc`
- `mcp`
- `sse`
- `websocket`
- `swagger`

Observações:

- `prometheus` adiciona o serviço e cria `devops/docker/prometheus.yml` se ele ainda não existir.
- `otel` adiciona `otel-collector` e configura `OTEL_EXPORTER_OTLP_ENDPOINT` na app.
- Se `devops/docker/docker-compose.yml` não existir, o `ginger add` só gera o código da integração.

---

## Kubernetes

### Deployment YAML Gerado

```yaml
# devops/kubernetes/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foobar
  labels:
    app: foobar
spec:
  replicas: 3
  selector:
    matchLabels:
      app: foobar
  template:
    metadata:
      labels:
        app: foobar
    spec:
      containers:
      - name: foobar
        image: foobar:latest
        ports:
        - containerPort: 8080
        env:
        - name: APP_ENV
          value: "production"
        - name: DATABASE_DSN
          valueFrom:
            secretKeyRef:
              name: foobar-secrets
              key: database-dsn
        - name: LOG_LEVEL
          value: "info"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: foobar
spec:
  selector:
    app: foobar
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
```

### Secrets

```yaml
# devops/kubernetes/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: foobar-secrets
type: Opaque
stringData:
  database-dsn: "postgres://<user>:<password>@postgres:5432/foobar?sslmode=disable"
```

```bash
# Criar secret
kubectl apply -f devops/kubernetes/secrets.yaml

# Ou via CLI
kubectl create secret generic foobar-secrets \
  --from-literal=database-dsn="postgres://<user>:<password>@host/db"
```

### ConfigMap

```yaml
# devops/kubernetes/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: foobar-config
data:
  app.yaml: |
    app:
      name: foobar
      env: production
    http:
      port: 8080
      shutdown_timeout: 30
    log:
      level: info
      format: json
```

```yaml
# Montar ConfigMap no Deployment
spec:
  containers:
  - name: foobar
    volumeMounts:
    - name: config
      mountPath: /configs
  volumes:
  - name: config
    configMap:
      name: foobar-config
```

### Deploy

```bash
# Apply all manifests
kubectl apply -f devops/kubernetes/

# Check status
kubectl get pods -l app=foobar
kubectl get svc foobar

# View logs
kubectl logs -f deployment/foobar

# Scale
kubectl scale deployment foobar --replicas=5

# Rollout
kubectl rollout status deployment/foobar
kubectl rollout history deployment/foobar
kubectl rollout undo deployment/foobar

# Port forward (local testing)
kubectl port-forward svc/foobar 8080:80
```

---

## Helm Charts

### Chart.yaml Gerado

```yaml
# devops/helm/Chart.yaml
apiVersion: v2
name: foobar
description: A Ginger-based API
type: application
version: 0.1.0
appVersion: "1.0.0"
```

### values.yaml Gerado

```yaml
# devops/helm/values.yaml
replicaCount: 3

image:
  repository: foobar
  pullPolicy: IfNotPresent
  tag: "latest"

service:
  type: ClusterIP
  port: 80
  targetPort: 8080

ingress:
  enabled: false
  className: "nginx"
  annotations: {}
  hosts:
    - host: api.example.com
      paths:
        - path: /
          pathType: Prefix
  tls: []

resources:
  requests:
    memory: "128Mi"
    cpu: "100m"
  limits:
    memory: "512Mi"
    cpu: "500m"

autoscaling:
  enabled: false
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 80

env:
  - name: APP_ENV
    value: "production"
  - name: LOG_LEVEL
    value: "info"

secrets:
  DATABASE_DSN: "postgres://<user>:<password>@postgres:5432/foobar"
```

### Deployment Template

```yaml
# devops/helm/templates/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "foobar.fullname" . }}
  labels:
    {{- include "foobar.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "foobar.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "foobar.selectorLabels" . | nindent 8 }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - name: http
          containerPort: {{ .Values.service.targetPort }}
          protocol: TCP
        env:
        {{- range .Values.env }}
        - name: {{ .name }}
          value: {{ .value | quote }}
        {{- end }}
        {{- range $key, $value := .Values.secrets }}
        - name: {{ $key }}
          valueFrom:
            secretKeyRef:
              name: {{ include "foobar.fullname" $ }}-secrets
              key: {{ $key }}
        {{- end }}
        livenessProbe:
          httpGet:
            path: /health
            port: http
        readinessProbe:
          httpGet:
            path: /health
            port: http
        resources:
          {{- toYaml .Values.resources | nindent 12 }}
```

### Helm Commands

```bash
# Install
helm install foobar ./devops/helm

# Install with custom values
helm install foobar ./devops/helm -f values-prod.yaml

# Upgrade
helm upgrade foobar ./helm

# Rollback
helm rollback foobar 1

# Uninstall
helm uninstall foobar

# Dry run
helm install foobar ./devops/helm --dry-run --debug

# Template (render locally)
helm template foobar ./helm
```

### Ambientes Múltiplos

```bash
# values-dev.yaml
replicaCount: 1
env:
  - name: APP_ENV
    value: "development"
  - name: LOG_LEVEL
    value: "debug"

# values-staging.yaml
replicaCount: 2
env:
  - name: APP_ENV
    value: "staging"

# values-prod.yaml
replicaCount: 5
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
env:
  - name: APP_ENV
    value: "production"
```

```bash
helm install foobar-dev ./devops/helm -f values-dev.yaml
helm install foobar-staging ./devops/helm -f values-staging.yaml
helm install foobar-prod ./devops/helm -f values-prod.yaml
```

---

## CI/CD

### GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]
    tags: ['v*']

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v2
      
      - name: Log in to Container Registry
        uses: docker/login-action@v2
        with:
          registry: ${{ env.REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      
      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v4
        with:
          images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
      
      - name: Build and push
        uses: docker/build-push-action@v4
        with:
          context: .
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
  
  deploy:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up kubectl
        uses: azure/setup-kubectl@v3
      
      - name: Configure kubectl
        run: |
          echo "${{ secrets.KUBECONFIG }}" | base64 -d > kubeconfig
          export KUBECONFIG=kubeconfig
      
      - name: Deploy to Kubernetes
        run: |
          kubectl set image deployment/foobar \
            foobar=${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:main
          kubectl rollout status deployment/foobar
```

### GitLab CI

```yaml
# .gitlab-ci.yml
stages:
  - build
  - deploy

variables:
  DOCKER_DRIVER: overlay2
  IMAGE_TAG: $CI_REGISTRY_IMAGE:$CI_COMMIT_SHORT_SHA

build:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build -t $IMAGE_TAG .
    - docker push $IMAGE_TAG
  only:
    - main
    - tags

deploy:production:
  stage: deploy
  image: bitnami/kubectl:latest
  before_script:
    - echo "$KUBECONFIG" | base64 -d > /tmp/kubeconfig
    - export KUBECONFIG=/tmp/kubeconfig
  script:
    - kubectl set image deployment/foobar foobar=$IMAGE_TAG
    - kubectl rollout status deployment/foobar
  only:
    - main
  environment:
    name: production
    url: https://api.example.com
```

---

## Ambientes

### Configuração por Ambiente

```yaml
# configs/app.dev.yaml
app:
  env: development
log:
  level: debug
  format: json

# configs/app.staging.yaml
app:
  env: staging
log:
  level: info
  format: json

# configs/app.prod.yaml
app:
  env: production
log:
  level: warn
  format: json
```

```go
// Carregar config baseado em env
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}

configPath := fmt.Sprintf("configs/app.%s.yaml", env)
cfg, err := config.Load(configPath)
```

### Variáveis de Ambiente

```bash
# .env.dev
APP_ENV=development
DATABASE_DSN=postgres://localhost:5432/foobar
LOG_LEVEL=debug

# .env.staging
APP_ENV=staging
DATABASE_DSN=postgres://staging-db:5432/foobar
LOG_LEVEL=info

# .env.prod
APP_ENV=production
DATABASE_DSN=postgres://prod-db:5432/foobar
LOG_LEVEL=warn
```

---

## Monitoramento

### Health Checks

```go
// Configurar health checks
h := health.New()
h.Register(database.NewChecker(db))
h.Register(cache.NewChecker(redisClient))

app.Router.HandleRaw("/health", h)
```

**Kubernetes Probes:**

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 2
```

### Prometheus Metrics

```go
// Adicionar Prometheus
m := metrics.NewDefaultMetrics("myapi")
app.Router.Use(metrics.Middleware(m))
app.Router.HandleRaw("/metrics", metrics.Handler())
```

**Kubernetes ServiceMonitor:**

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: foobar
spec:
  selector:
    matchLabels:
      app: foobar
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### OpenTelemetry

```go
// Setup telemetry
shutdown, err := telemetry.Setup(ctx, "foobar", "1.0.0")
if err != nil {
    log.Fatal(err)
}
app.OnStop(shutdown)
```

**Environment Variables:**

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318
OTEL_EXPORTER_OTLP_HEADERS="x-api-key=secret"
```

---

## Troubleshooting

### Logs

```bash
# Docker
docker logs -f foobar

# Docker Compose
docker compose logs -f foobar

# Kubernetes
kubectl logs -f deployment/foobar
kubectl logs -f deployment/foobar --previous  # logs do container anterior

# Logs de múltiplos pods
kubectl logs -l app=foobar --tail=100 -f
```

### Debug Container

```bash
# Kubernetes — exec into pod
kubectl exec -it deployment/foobar -- sh

# Docker
docker exec -it foobar sh
```

### Port Forward

```bash
# Kubernetes
kubectl port-forward svc/foobar 8080:80

# Docker Compose
# Já exposto via ports: no docker-compose.yml
```

### Database Connection

```bash
# Test connection from pod
kubectl exec -it deployment/foobar -- sh
apk add postgresql-client
psql $DATABASE_DSN
```

### Common Issues

#### 1. CrashLoopBackOff

```bash
# Ver logs
kubectl logs deployment/foobar

# Verificar eventos
kubectl describe pod <pod-name>

# Causas comuns:
# - Erro de conexão com banco
# - Variável de ambiente faltando
# - Porta já em uso
```

#### 2. ImagePullBackOff

```bash
# Verificar secret de registry
kubectl get secret

# Criar secret se necessário
kubectl create secret docker-registry regcred \
  --docker-server=ghcr.io \
  --docker-username=<username> \
  --docker-password=<token>

# Adicionar ao deployment
spec:
  imagePullSecrets:
  - name: regcred
```

#### 3. Readiness Probe Failed

```bash
# Verificar health endpoint
kubectl port-forward svc/foobar 8080:80
curl http://localhost:8080/health

# Ajustar probe timing
readinessProbe:
  initialDelaySeconds: 30  # aumentar se app demora para iniciar
  periodSeconds: 10
  timeoutSeconds: 5
```

---

## Próximos Passos

- [🏗️ Arquitetura](./ARCHITECTURE.md) — Estrutura e padrões
- [📦 Pacotes](./PACKAGES.md) — Documentação de cada pacote
- [🔌 Integrações](./INTEGRATIONS.md) — Bancos, cache, mensageria
- [🧪 Testes](./TESTING.md) — Estratégias de teste

[← Voltar ao README](../README.md)
