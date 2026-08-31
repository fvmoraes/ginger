package integrations

import "sort"

// DockerService is one compose service contributed by an integration.
// GIN-026/Fase 4: central source of truth for service names, images and
// published host ports — the merge switch and the port-conflict detector
// both derive from it, so they cannot drift apart.
type DockerService struct {
	Service   string
	Image     string
	HostPorts []string // host-side published ports only (e.g. "5432")
}

// DockerServicesByIntegration maps integration name → its compose services.
var DockerServicesByIntegration = map[string][]DockerService{
	"postgres":   {{Service: "postgres", Image: "postgres:16-alpine", HostPorts: []string{"5432"}}},
	"mysql":      {{Service: "mysql", Image: "mysql:8", HostPorts: []string{"3306"}}},
	"redis":      {{Service: "redis", Image: "redis:7-alpine", HostPorts: []string{"6379"}}},
	"rabbitmq":   {{Service: "rabbitmq", Image: "rabbitmq:3-management-alpine", HostPorts: []string{"5672", "15672"}}},
	"kafka":      {{Service: "kafka", Image: "bitnami/kafka:3.7", HostPorts: []string{"9092"}}},
	"nats":       {{Service: "nats", Image: "nats:2-alpine", HostPorts: []string{"4222", "8222"}}},
	"mongodb":    {{Service: "mongodb", Image: "mongo:7", HostPorts: []string{"27017"}}},
	"clickhouse": {{Service: "clickhouse", Image: "clickhouse/clickhouse-server:24.3", HostPorts: []string{"8123", "9000"}}},
	"couchbase":  {{Service: "couchbase", Image: "couchbase:community-7.6.2", HostPorts: []string{"8091", "11210"}}},
	"prometheus": {{Service: "prometheus", Image: "prom/prometheus:latest", HostPorts: []string{"9090"}}},
	"otel":       {{Service: "otel-collector", Image: "otel/opentelemetry-collector:0.102.1", HostPorts: []string{"4317", "4318"}}},
}

// hostPort extracts the published (host-side) port from a Compose port
// mapping: "5432:5432" → 5432; "127.0.0.1:5432:5432" → 5432; "5432" → 5432.
func hostPort(mapping string) string {
	first := mapping
	if idx := indexByte(first, ':'); idx >= 0 {
		first = first[:idx]
	}
	// strip protocol suffix and IP prefix
	if idx := lastSlashIndex(first); idx >= 0 {
		first = first[idx+1:]
	}
	return first
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func lastSlashIndex(s string) int {
	idx := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '/' {
			idx = i
		}
	}
	return idx
}

// portConflict describes a published-port collision (GIN-026).
type portConflict struct {
	hostPort string
	owner    string
}

// detectPortConflicts (GIN-026) returns collisions between the host ports
// the integration's services want to publish and ports already published by
// other services in the compose.
func detectPortConflicts(published map[string]string, services []DockerService) []portConflict {
	var conflicts []portConflict
	for _, svc := range services {
		for _, hp := range svc.HostPorts {
			if owner, ok := published[hp]; ok && owner != svc.Service {
				conflicts = append(conflicts, portConflict{hostPort: hp, owner: owner})
			}
		}
	}
	sort.Slice(conflicts, func(i, j int) bool { return conflicts[i].hostPort < conflicts[j].hostPort })
	return conflicts
}
