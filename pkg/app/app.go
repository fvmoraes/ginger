// Package app provides the application bootstrap, dependency injection container,
// HTTP server lifecycle, and graceful shutdown.
//
// The App type is the central wiring point. It owns the HTTP server, router,
// health handler, and logger. Shutdown hooks registered via OnStop are called
// in order during graceful shutdown, giving each component a chance to flush
// and release resources before the process exits.
//
// Graceful shutdown follows the pattern described in The Go Programming Language
// (Donovan & Kernighan): listen for SIGINT/SIGTERM, then call server.Shutdown
// with a context deadline so in-flight requests can complete.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fvmoraes/ginger/pkg/config"
	"github.com/fvmoraes/ginger/pkg/health"
	"github.com/fvmoraes/ginger/pkg/logger"
	"github.com/fvmoraes/ginger/pkg/middleware"
	"github.com/fvmoraes/ginger/pkg/router"
)

// App is the central application container.
type App struct {
	Config *config.Config
	Logger *logger.Logger
	Router *router.Router
	Health *health.Handler
	server *http.Server
	onStop []func(context.Context) error
}

// New creates an App with sensible defaults.
func New(cfg *config.Config) *App {
	log := logger.New(cfg.Log.Level, cfg.Log.Format)
	r := router.New()
	h := health.New()

	a := &App{
		Config: cfg,
		Logger: log,
		Router: r,
		Health: h,
	}

	// Default middlewares
	r.Use(
		middleware.Recover(log),
		middleware.RequestID(),
		middleware.Logger(log),
	)

	// Built-in health endpoint (bypasses prefix/middleware chain)
	r.HandleRaw("GET /health", h)

	return a
}

// OnStop registers a shutdown hook called during graceful shutdown.
func (a *App) OnStop(fn func(context.Context) error) {
	a.onStop = append(a.onStop, fn)
}

// Run starts the HTTP server and blocks until a signal is received.
func (a *App) Run() error {
	addr := fmt.Sprintf("%s:%d", a.Config.HTTP.Host, a.Config.HTTP.Port)
	a.server = &http.Server{
		Addr:         addr,
		Handler:      a.Router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		a.Logger.Info("app_started", "addr", addr, "env", a.Config.App.Env)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	select {
	case err, ok := <-errCh:
		if ok && err != nil {
			return err
		}
	case sig := <-quit:
		a.Logger.Info("shutdown_signal_received", "signal", sig.String())
	}

	return a.shutdown()
}

func (a *App) shutdown() error {
	timeout := time.Duration(a.Config.HTTP.ShutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	a.Logger.Info("app_stopping", "timeout", timeout.String())

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("app: server shutdown: %w", err)
	}

	for i := len(a.onStop) - 1; i >= 0; i-- {
		if err := a.onStop[i](ctx); err != nil {
			a.Logger.Error("shutdown_hook_error", "error", err)
		}
	}

	a.Logger.Info("app_stopped")
	return nil
}
