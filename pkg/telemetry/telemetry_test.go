package telemetry

import (
	"context"
	"testing"
	"time"
)

// TestSetupAndShutdownStdout (GIN-010): primeiro teste do submódulo — valida
// que o ciclo de vida básico (Setup com exporter stdout + Shutdown) funciona.
func TestSetupAndShutdownStdout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	provider, err := Setup(ctx, Config{Exporter: "stdout", ServiceName: "ginger-telemetry-test"})
	if err != nil {
		t.Fatalf("Setup(stdout): %v", err)
	}
	if provider == nil {
		t.Fatal("Setup returned nil provider without error")
	}
	if Tracer("test") == nil {
		t.Fatal("provider.Tracer returned nil")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := provider.Shutdown(shutdownCtx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
