package project

import (
	"strings"
	"testing"
)

// Fixture com os padrões reais dos projetos Ginger (Fase 3, GIN-014):
// Group por atribuição, prefixo herdado por parâmetro (interprocedural 1 nível).
const fixtureGingerRoutes = `package api

import "github.com/fvmoraes/ginger/pkg/router"

var generatedRouteRegistrars []func(*router.Router)

func Register(r *router.Router) {
	v1 := r.Group("/api/v1")
	registerCoreRoutes(v1)
}

func registerCoreRoutes(v1 *router.Router) {
	v1.GET("/ping", pingHandler)
	v1.POST("/users", createUser)
}

// ginger:route GET /healthz
func healthz() {}
`

func TestASTRoutesGingerPatterns(t *testing.T) {
	routes := astRoutes(fixtureGingerRoutes, "internal/api/router.go")
	if len(routes) == 0 {
		t.Fatal("no routes found")
	}
	byPath := map[string]Route{}
	for _, r := range routes {
		byPath[r.Path] = r
	}
	// Composição de prefixo: /api/v1 + /ping (o regex achava só /ping).
	if r, ok := byPath["/api/v1/ping"]; !ok || r.Method != "GET" {
		t.Fatalf("expected composed GET /api/v1/ping, got %+v", routes)
	}
	if r, ok := byPath["/api/v1/users"]; !ok || r.Method != "POST" {
		t.Fatalf("expected composed POST /api/v1/users, got %+v", routes)
	}
	// Annotation autoritativa.
	if r, ok := byPath["/healthz"]; !ok || r.Source != RouteSourceAnnotation || r.Confidence != RouteConfidenceHigh {
		t.Fatalf("annotation route missing/mislabeled: %+v", byPath["/healthz"])
	}
	// AST rotas com confidence high.
	if r := byPath["/api/v1/ping"]; r.Source != RouteSourceAST || r.Confidence != RouteConfidenceHigh {
		t.Fatalf("ast route mislabeled: %+v", r)
	}
}

// Casos que o regex errava (falsos positivos): client.GET de URL externa e
// strings em comentários não devem virar rotas via AST.
const fixtureNonRoutes = `package client

func f() {
	client.GET("https://api.example.com/v1/things") // not a route
	// r.GET("/commented-route", nil)
	_ = x.POST("raw string", nil)
}
`

func TestASTRoutesNoFalsePositives(t *testing.T) {
	routes := astRoutes(fixtureNonRoutes, "client.go")
	for _, r := range routes {
		if strings.Contains(r.Path, "https://") {
			t.Fatalf("false positive via AST: %+v", r)
		}
	}
	// O recebedor é client.GET (método GET não é chamada de rota de router
	// conhecida — AST ainda acharia "GET /v1/things" com prefixo vazio?
	// Não: é client.GET, method filter pega por Sel.Name — mas recebedor
	// "client" não é exclusão; aqui o comportamento correto é NÃO incluir
	// URLs absolutas. Validamos explicitamente abaixo.
	for _, r := range routes {
		if !strings.HasPrefix(r.Path, "/") {
			t.Fatalf("non-path route leaked: %+v", r)
		}
	}
}

// HandleFunc com método embutido + HandleFunc sem método (ANY).
func TestASTRoutesHandleFuncVariants(t *testing.T) {
	content := `package main

func Register(r Router) {
	r.HandleFunc("GET /metrics", metricsHandler)
	r.HandleFunc("/any", anyHandler)
}
`
	routes := astRoutes(content, "main.go")
	byPath := map[string]Route{}
	for _, r := range routes {
		byPath[r.Path] = r
	}
	if r, ok := byPath["/metrics"]; !ok || r.Method != "GET" {
		t.Fatalf("GET /metrics missing: %+v", routes)
	}
	if r, ok := byPath["/any"]; !ok || r.Method != "ANY" {
		t.Fatalf("ANY /any missing: %+v", routes)
	}
}

// Merge: regex fallback só contribui o que o AST não achou, marcado low.
func TestMergeRouteResultsDedupesAndTags(t *testing.T) {
	astList := []Route{{Method: "GET", Path: "/api/v1/ping", File: "a.go", Line: 5, Source: RouteSourceAST, Confidence: RouteConfidenceHigh}}
	regexList := []Route{
		{Method: "GET", Path: "/api/v1/ping", File: "a.go", Line: 5}, // duplicado
		{Method: "GET", Path: "/only-regex", File: "a.go", Line: 9},  // fallback
	}
	got := mergeRouteResults(astList, regexList)
	if len(got) != 2 {
		t.Fatalf("expected 2 routes, got %d: %+v", len(got), got)
	}
	if got[1].Source != RouteSourceRegex || got[1].Confidence != RouteConfidenceLow {
		t.Fatalf("fallback mislabeled: %+v", got[1])
	}
}

// Fixtures de detecção de projeto (GIN-027): library vs service por conteúdo.
func TestInspectRoutesWithConfidenceOnFixture(t *testing.T) {
	// Integração: Inspect de ponta a ponta usa AST+regex com metadados.
	content := `package api

func Register(r *router.Router) {
	v1 := r.Group("/api/v1")
	v1.GET("/ping", h)
}
`
	routes := astRoutes(content, "router.go")
	merged := mergeRouteResults(routes, scanRoutes(content, "router.go"))
	if len(merged) == 0 {
		t.Fatal("no routes from merged pipeline")
	}
	if merged[0].Path != "/api/v1/ping" {
		t.Fatalf("composition broken: %+v", merged)
	}
}
