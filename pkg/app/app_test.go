package app

import (
	"context"
	"net/http"
	"testing"

	"github.com/fvmoraes/ginger/pkg/config"
)

func TestShutdownRunsOnStopHooksInLIFOOrder(t *testing.T) {
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("config.Load returned error: %v", err)
	}
	cfg.Log.Level = "error"
	cfg.HTTP.ShutdownTimeout = 1

	app := New(cfg)
	app.server = &http.Server{}

	var calls []int
	app.OnStop(func(context.Context) error {
		calls = append(calls, 1)
		return nil
	})
	app.OnStop(func(context.Context) error {
		calls = append(calls, 2)
		return nil
	})
	app.OnStop(func(context.Context) error {
		calls = append(calls, 3)
		return nil
	})

	if err := app.shutdown(); err != nil {
		t.Fatalf("shutdown returned error: %v", err)
	}

	want := []int{3, 2, 1}
	if len(calls) != len(want) {
		t.Fatalf("expected %d hook calls, got %d", len(want), len(calls))
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Fatalf("expected LIFO order %v, got %v", want, calls)
		}
	}
}
