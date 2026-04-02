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

Projetos `api`, `service` e `worker` gerados pelo Ginger vêm com um `Dockerfile` multi-stage otimizado:

```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build (example for API project type)
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
docker build -t foobar:latest .

# Run
docker run -p 8080:8080 \
  -e DATABASE_DSN="postgres://user:pass@host/db" \
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

### docker-compose.yml Gerado (projetos API/service)

```yaml
version: "3.9"

services:
  foobar:
    build: .
    ports:
      - "8080:8080"
    environment:
      APP_ENV: development
      HTTP_PORT: 8080
      DATABASE_DSN: postgres://user:pass@postgres:5432/foobar?sslmode=disable
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
      POSTGRES_DB: foobar
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./scripts/prometheus.yml:/etc/prometheus/prometheus.yml
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    depends_on:
      - prometheus

volumes:
  pgdata:
```

### Comandos

```bash
# Start all services
docker compose up -d

# View logs
docker compose logs -f foobar

# Stop all services
docker compose down

# Rebuild and restart
docker compose up -d --build

# Access database
docker compose exec postgres psql -U user -d foobar
```

---

## Kubernetes

### Deployment YAML Gerado

```yaml
# kubernetes/deployment.yaml
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
# kubernetes/secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: foobar-secrets
type: Opaque
stringData:
  database-dsn: "postgres://user:pass@postgres:5432/foobar?sslmode=disable"
```

```bash
# Criar secret
kubectl apply -f kubernetes/secrets.yaml

# Ou via CLI
kubectl create secret generic foobar-secrets \
  --from-literal=database-dsn="postgres://user:pass@host/db"
```

### ConfigMap

```yaml
# kubernetes/configmap.yaml
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
kubectl apply -f kubernetes/

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
# helm/Chart.yaml
apiVersion: v2
name: foobar
description: A Ginger-based API
type: application
version: 0.1.0
appVersion: "1.0.0"
```

### values.yaml Gerado

```yaml
# helm/values.yaml
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
  DATABASE_DSN: "postgres://user:pass@postgres:5432/foobar"
```

### Deployment Template

```yaml
# helm/templates/deployment.yaml
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
helm install foobar ./helm

# Install with custom values
helm install foobar ./helm -f values-prod.yaml

# Upgrade
helm upgrade foobar ./helm

# Rollback
helm rollback foobar 1

# Uninstall
helm uninstall foobar

# Dry run
helm install foobar ./helm --dry-run --debug

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
helm install foobar-dev ./helm -f values-dev.yaml
helm install foobar-staging ./helm -f values-staging.yaml
helm install foobar-prod ./helm -f values-prod.yaml
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
  format: text

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
