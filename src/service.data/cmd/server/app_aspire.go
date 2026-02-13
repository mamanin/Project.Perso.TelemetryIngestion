package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	static "service.data"
	"service.data/external/storage/adx"
	"service.data/internal/core/handlers"
	"service.data/internal/core/observability/logger"
)

// AspireApp represents the entire application with all its dependencies.
type AspireApp struct {
	cfg *AspireConfig

	logger logger.Logger
	router *chi.Mux
	adx    *adx.Client
}

// NewAspireApp creates and initializes a new AppAspire instance with all dependencies.
func NewAspireApp(ctx context.Context) (AppManager, error) {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log, err := logger.NewOtelLogger(
		tCtx,
		attribute.String("service.layer", "web-server"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize open telemetry logger: %w", err)
	}

	cfg, err := LoadAspireConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	app := &AspireApp{
		cfg:    cfg,
		logger: log,
	}

	if err = app.initializeAdx(); err != nil {
		return nil, fmt.Errorf("failed to initialize ADX client: %w", err)
	}

	app.initializeRouter()

	return app, nil
}

// Start starts all application services.
func (a *AspireApp) Start() error {
	a.logger.Info("Starting web server...")

	addr := ":8080" // TODO: config
	a.logger.Info("Server starting on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, a.router); err != nil {
		return fmt.Errorf("failed to start web server: %w", err)
	}

	a.logger.Info("Web server started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *AspireApp) Stop() {
	a.logger.Info("Shutting down web server...")

	if err := a.adx.Close(); err != nil {
		a.logger.Error(err, "Error closing ADX client: %v", err)
	}
	a.logger.Info("Web server shutdown complete")
}

// initializeRouter sets up the HTTP router with all routes and middleware.
func (a *AspireApp) initializeRouter() {
	r := chi.NewRouter()
	r.Use(middleware.Logger) // TODO: use custom one
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Handle("/static/*", http.FileServer(http.FS(static.Files)))

	h := handlers.New(a.logger, a.adx)

	r.Get("/devices", h.DevicesPageHandler)
	r.Get("/devices/search", h.SearchDevicesHandler)

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/devices", http.StatusTemporaryRedirect)
	})

	a.router = r
}

// initializeAdx initializes the ADX client based on the application configuration.
func (a *AspireApp) initializeAdx() error {
	c, err := adx.NewClientForAspire(a.cfg.adx, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize ADX client: %w", err)
	}

	a.adx = c
	return nil
}
