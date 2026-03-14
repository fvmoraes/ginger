// Package app provides the application bootstrap, dependency injection container,
// HTTP server lifecycle, and graceful shutdown.
package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ginger-framework/ginger/pkg/config"
	"github.com/ginger-framework/ginger/pkg/health"
	"github.com/ginger-framework/ginger/pkg/logger"
	"github.com/ginger-framework/ginger/pkg/middleware"
	"github.com/ginger-framework/ginger/pkg/router"
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
		a.Logger.Info("server starting", "addr", addr, "env", a.Config.App.Env)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-quit:
		a.Logger.Info("shutdown signal received", "signal", sig.String())
	}

	return a.shutdown()
}

func (a *App) shutdown() error {
	timeout := time.Duration(a.Config.HTTP.ShutdownTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	a.Logger.Info("shutting down server", "timeout", timeout.String())

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("app: server shutdown: %w", err)
	}

	for _, fn := range a.onStop {
		if err := fn(ctx); err != nil {
			a.Logger.Error("shutdown hook error", "error", err)
		}
	}

	a.Logger.Info("server stopped")
	return nil
}
