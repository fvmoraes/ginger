# Guia de Integrações

[← Voltar ao README](../README.md)

## Índice

- [Visão Geral](#visão-geral)
- [Bancos de Dados](#bancos-de-dados)
- [Cache](#cache)
- [NoSQL](#nosql)
- [Mensageria](#mensageria)
- [Protocolos](#protocolos)
- [Observabilidade](#observabilidade)
- [Real-time](#real-time)

---

## Visão Geral

O comando `ginger add <integration>` gera código boilerplate e adiciona dependências automaticamente.

### Comando

```bash
ginger add <integration>
```

### Integrações Disponíveis

| Categoria | Integração | Pacote | Arquivo Gerado |
|-----------|------------|--------|----------------|
| **Databases** | `postgres` | `github.com/lib/pq` | `platform/database/postgres.go` |
| | `mysql` | `github.com/go-sql-driver/mysql` | `platform/database/mysql.go` |
| | `sqlite` | `github.com/mattn/go-sqlite3` | `platform/database/sqlite.go` |
| | `sqlserver` | `github.com/microsoft/go-mssqldb` | `platform/database/sqlserver.go` |
| **NoSQL** | `couchbase` | `github.com/couchbase/gocb/v2` | `platform/nosql/couchbase.go` |
| | `mongodb` | `go.mongodb.org/mongo-driver` | `platform/nosql/mongo.go` |
| **Analytical** | `clickhouse` | `github.com/ClickHouse/clickhouse-go/v2` | `platform/database/clickhouse.go` |
| **Cache** | `redis` | `github.com/redis/go-redis/v9` | `platform/cache/redis.go` |
| **Messaging** | `kafka` | `github.com/segmentio/kafka-go` | `platform/messaging/kafka.go` |
| | `rabbitmq` | `github.com/rabbitmq/amqp091-go` | `platform/messaging/rabbitmq.go` |
| | `nats` | `github.com/nats-io/nats.go` | `platform/messaging/nats.go` |
| | `pubsub` | `cloud.google.com/go/pubsub` | `platform/messaging/pubsub.go` |
| **Protocols** | `grpc` | `google.golang.org/grpc` | `platform/grpc/server.go` |
| | `mcp` | stdlib only | `platform/mcp/server.go` |
| **Real-time** | `sse` | stdlib only | `internal/api/handlers/sse_handler.go` |
| | `websocket` | stdlib only | `internal/api/handlers/ws_handler.go` |
| **Observability** | `otel` | `go.opentelemetry.io/otel` | `platform/telemetry/otel.go` |
| | `prometheus` | `github.com/prometheus/client_golang` | `platform/metrics/prometheus.go` |

---

## Bancos de Dados

### PostgreSQL

```bash
ginger add postgres
```

**Uso:**

```go
import "yourmodule/platform/database"

cfg := database.Config{
    DSN:     "postgres://<user>:<password>@localhost:5432/foobar?sslmode=disable",
    MaxOpen: 25,
    MaxIdle: 5,
}

db, err := database.Connect(cfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Health check
healthHandler := health.New()
healthHandler.Register(database.NewChecker(db))
```

**DSN Format:**
```
postgres://<user>:<password>@host:port/database?sslmode=disable
```

**Variáveis de Ambiente:**
```bash
DATABASE_DRIVER=postgres
DATABASE_DSN="postgres://<user>:<password>@localhost:5432/foobar?sslmode=disable"
```

### MySQL

```bash
ginger add mysql
```

**DSN Format:**
```
<user>:<password>@tcp(host:3306)/database?parseTime=true
```

**Exemplo:**
```go
cfg := database.MySQLConfig{
    DSN:     "<user>:<password>@tcp(localhost:3306)/foobar?parseTime=true",
    MaxOpen: 25,
    MaxIdle: 5,
}

db, err := database.ConnectMySQL(cfg)
```

### SQLite

```bash
ginger add sqlite
```

**Uso:**
```go
db, err := database.ConnectSQLite("./data.db")
// ou in-memory
db, err := database.ConnectSQLite(":memory:")
```

**Nota:** SQLite usa CGO. Para builds estáticos, considere usar `modernc.org/sqlite` (pure Go).

### SQL Server

```bash
ginger add sqlserver
```

**DSN Format:**
```
sqlserver://<user>:<password>@host:1433?database=foobar
```

---

## Cache

### Redis

```bash
ginger add redis
```

**Uso:**

```go
import "yourmodule/platform/cache"

cfg := cache.Config{
    Addr:     "localhost:6379",
    Password: "",
    DB:       0,
}

client, err := cache.Connect(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Set
ctx := context.Background()
client.Set(ctx, "key", "value", 10*time.Minute)

// Get
val, err := client.Get(ctx, "key").Result()

// Health check
healthHandler.Register(cache.NewChecker(client))
```

**Comandos Comuns:**

```go
// String operations
client.Set(ctx, "key", "value", 0)
client.Get(ctx, "key")
client.Del(ctx, "key")

// Hash operations
client.HSet(ctx, "user:1", "name", "Alice")
client.HGet(ctx, "user:1", "name")
client.HGetAll(ctx, "user:1")

// List operations
client.LPush(ctx, "queue", "item1", "item2")
client.RPop(ctx, "queue")

// Set operations
client.SAdd(ctx, "tags", "go", "redis")
client.SMembers(ctx, "tags")

// Sorted set
client.ZAdd(ctx, "leaderboard", redis.Z{Score: 100, Member: "player1"})
client.ZRange(ctx, "leaderboard", 0, 9)
```

---

## NoSQL

### MongoDB

```bash
ginger add mongodb
```

**Uso:**

```go
import "yourmodule/platform/nosql"

cfg := nosql.MongoConfig{
    URI:      "mongodb://localhost:27017",
    Database: "foobar",
}

client, db, err := nosql.ConnectMongo(cfg)
if err != nil {
    log.Fatal(err)
}
defer client.Disconnect(context.Background())

// Collection
users := db.Collection("users")

// Insert
result, err := users.InsertOne(ctx, bson.M{
    "name":  "Alice",
    "email": "alice@example.com",
})

// Find one
var user User
err = users.FindOne(ctx, bson.M{"email": "alice@example.com"}).Decode(&user)

// Find many
cursor, err := users.Find(ctx, bson.M{"active": true})
defer cursor.Close(ctx)

var users []User
cursor.All(ctx, &users)

// Update
users.UpdateOne(ctx, 
    bson.M{"_id": id},
    bson.M{"$set": bson.M{"name": "Alice Updated"}},
)

// Delete
users.DeleteOne(ctx, bson.M{"_id": id})

// Health check
healthHandler.Register(nosql.NewMongoChecker(client))
```

### Couchbase

```bash
ginger add couchbase
```

**Uso:**

```go
cfg := nosql.CouchbaseConfig{
    ConnectionString: "couchbase://localhost",
    Username:         "Administrator",
    Password:         "password",
    BucketName:       "foobar",
}

cluster, bucket, err := nosql.ConnectCouchbase(cfg)
if err != nil {
    log.Fatal(err)
}
defer cluster.Close(nil)

// Get collection
collection := bucket.DefaultCollection()

// Insert
_, err = collection.Insert("user::1", User{Name: "Alice"}, nil)

// Get
var user User
_, err = collection.Get("user::1", &user, nil)

// Update
_, err = collection.Replace("user::1", user, nil)

// Delete
_, err = collection.Remove("user::1", nil)

// N1QL query
rows, err := cluster.Query("SELECT * FROM `foobar` WHERE type = 'user'", nil)
defer rows.Close()

// Health check
healthHandler.Register(nosql.NewCouchbaseChecker(cluster))
```

### ClickHouse

```bash
ginger add clickhouse
```

**Uso:**

```go
cfg := database.ClickHouseConfig{
    Addr:     []string{"localhost:9000"},
    Database: "default",
    Username: "default",
    Password: "",
    MaxOpen:  10,
    MaxIdle:  5,
}

db, err := database.ConnectClickHouse(cfg)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Insert
_, err = db.Exec(`
    INSERT INTO events (timestamp, user_id, event_type)
    VALUES (?, ?, ?)
`, time.Now(), 123, "click")

// Query
rows, err := db.Query(`
    SELECT event_type, count() as cnt
    FROM events
    WHERE timestamp >= ?
    GROUP BY event_type
`, time.Now().Add(-24*time.Hour))

// Health check
healthHandler.Register(database.NewClickHouseChecker(db))
```

---

## Mensageria

### Kafka

```bash
ginger add kafka
```

**Producer:**

```go
import "yourmodule/platform/messaging"

cfg := messaging.ProducerConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "events",
}

writer := messaging.NewWriter(cfg)
defer writer.Close()

// Publish
err := messaging.Publish(ctx, writer, []byte("key"), []byte(`{"event":"user.created"}`))
```

**Consumer:**

```go
cfg := messaging.ConsumerConfig{
    Brokers: []string{"localhost:9092"},
    Topic:   "events",
    GroupID: "foobar",
}

reader := messaging.NewReader(cfg)
defer reader.Close()

for {
    msg, err := reader.ReadMessage(ctx)
    if err != nil {
        break
    }
    fmt.Printf("Message: %s = %s\n", msg.Key, msg.Value)
}
```

### RabbitMQ

```bash
ginger add rabbitmq
```

**Uso:**

```go
rmq, err := messaging.ConnectRabbitMQ("amqp://guest:guest@localhost:5672/")
if err != nil {
    log.Fatal(err)
}
defer rmq.Close()

// Declare queue
_, err = rmq.Channel.QueueDeclare("tasks", true, false, false, false, nil)

// Publish
err = rmq.Publish(ctx, "", "tasks", []byte(`{"task":"send_email"}`))

// Consume
msgs, err := rmq.Channel.Consume("tasks", "", false, false, false, false, nil)
for msg := range msgs {
    fmt.Printf("Received: %s\n", msg.Body)
    msg.Ack(false)
}

// Health check
healthHandler.Register(messaging.NewRabbitMQChecker(rmq))
```

### NATS

```bash
ginger add nats
```

**Uso:**

```go
nc, err := messaging.ConnectNATS("nats://localhost:4222")
if err != nil {
    log.Fatal(err)
}
defer nc.Close()

// Publish
messaging.Publish(nc, "events", []byte(`{"event":"user.created"}`))

// Subscribe
sub, err := messaging.Subscribe(nc, "events", func(msg *nats.Msg) {
    fmt.Printf("Received: %s\n", msg.Data)
})
defer sub.Unsubscribe()

// Health check
healthHandler.Register(messaging.NewNATSChecker(nc))
```

### Google Pub/Sub

```bash
ginger add pubsub
```

**Uso:**

```go
client, err := messaging.ConnectPubSub(ctx, "foobar")
if err != nil {
    log.Fatal(err)
}
defer client.Close()

// Publish
msgID, err := client.Publish(ctx, "foobar", []byte(`{"event":"user.created"}`))

// Subscribe
err = client.Subscribe(ctx, "foobar", func(ctx context.Context, msg *pubsub.Message) {
    fmt.Printf("Received: %s\n", msg.Data)
    msg.Ack()
})
```

---

## Protocolos

### gRPC

```bash
ginger add grpc
```

**Server:**

```go
import "yourmodule/platform/grpc"

cfg := grpc.ServerConfig{Addr: ":50051"}
srv := grpc.NewServer(cfg)

// Register your service
srv.Register(func(s *grpc.Server) {
    pb.RegisterUserServiceServer(s, &userServiceImpl{})
})

// Start (blocking)
go srv.Serve()

// Graceful shutdown
app.OnStop(func(ctx context.Context) error {
    srv.Stop()
    return nil
})
```

**Client:**

```go
conn, err := grpc.NewClient(ctx, "localhost:50051")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewUserServiceClient(conn)
resp, err := client.GetUser(ctx, &pb.GetUserRequest{Id: 1})
```

**Health Check:**

gRPC server inclui health check automático (compatível com `grpc-health-probe`):

```bash
grpc-health-probe -addr=localhost:50051
```

### MCP (Model Context Protocol)

```bash
ginger add mcp
```

**Server:**

```go
import "yourmodule/platform/mcp"

mcpServer := mcp.NewServer()

// Register tool
mcpServer.Register(mcp.Tool{
    Name:        "get_user",
    Description: "Fetch user by ID",
    InputSchema: json.RawMessage(`{"type":"object","properties":{"id":{"type":"number"}}}`),
    Handler: func(ctx context.Context, input json.RawMessage) (any, error) {
        var req struct{ ID int `json:"id"` }
        json.Unmarshal(input, &req)
        
        user, err := userService.GetByID(ctx, req.ID)
        if err != nil {
            return nil, err
        }
        return user, nil
    },
})

// Mount handler
app.Router.HandleRaw("/mcp", mcpServer.Handler())
```

**Client (Claude Desktop, Cursor, Kiro):**

```json
{
  "mcpServers": {
    "foobar": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

---

## Observabilidade

### OpenTelemetry

```bash
ginger add otel
```

**Setup:**

```go
import "yourmodule/platform/telemetry"

shutdown, err := telemetry.Setup(ctx, "foobar", "1.0.0")
if err != nil {
    log.Fatal(err)
}
app.OnStop(shutdown)

// Create tracer
tracer := telemetry.Tracer("foobar")

// Trace operation
ctx, span := tracer.Start(ctx, "create-user")
defer span.End()

user, err := userService.Create(ctx, input)
if err != nil {
    span.RecordError(err)
    return err
}
```

**Environment Variables:**

```bash
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
OTEL_EXPORTER_OTLP_HEADERS="x-api-key=secret"
```

### Prometheus

```bash
ginger add prometheus
```

**Setup:**

```go
import "yourmodule/platform/metrics"

m := metrics.NewDefaultMetrics("myapi")

// Add middleware
app.Router.Use(metrics.Middleware(m))

// Expose /metrics endpoint
app.Router.HandleRaw("/metrics", metrics.Handler())
```

**Custom Metrics:**

```go
// Counter
requestsTotal := prometheus.NewCounter(prometheus.CounterOpts{
    Name: "myapi_requests_total",
    Help: "Total number of requests",
})
prometheus.MustRegister(requestsTotal)
requestsTotal.Inc()

// Gauge
activeConnections := prometheus.NewGauge(prometheus.GaugeOpts{
    Name: "myapi_active_connections",
    Help: "Number of active connections",
})
prometheus.MustRegister(activeConnections)
activeConnections.Set(42)

// Histogram
requestDuration := prometheus.NewHistogram(prometheus.HistogramOpts{
    Name:    "myapi_request_duration_seconds",
    Help:    "Request duration in seconds",
    Buckets: prometheus.DefBuckets,
})
prometheus.MustRegister(requestDuration)
requestDuration.Observe(0.123)
```

---

## Real-time

### Server-Sent Events

```bash
ginger add sse
```

Gera `internal/api/handlers/sse_handler.go` com exemplo completo.

**Uso:** Ver [pkg/sse](./PACKAGES.md#pkgsse)

### WebSocket

```bash
ginger add websocket
```

Gera `internal/api/handlers/ws_handler.go` com exemplo completo.

**Uso:** Ver [pkg/ws](./PACKAGES.md#pkgws)

---

## Próximos Passos

- [🏗️ Arquitetura](./ARCHITECTURE.md) — Estrutura e padrões
- [📦 Pacotes](./PACKAGES.md) — Documentação de cada pacote
- [🧪 Testes](./TESTING.md) — Estratégias de teste
- [🚀 Deploy](./DEPLOYMENT.md) — Docker, Kubernetes, CI/CD

[← Voltar ao README](../README.md)
