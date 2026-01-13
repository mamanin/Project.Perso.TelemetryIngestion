package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"service.ingestion/external/observability/logger"
)

// App represents the entire application with all its dependencies.
type App struct {
	logger logger.Logger
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (*App, error) {
	lToCtx, bCancel := context.WithTimeout(ctx, 10*time.Second)
	defer bCancel()

	log, err := logger.NewOtelLogger(
		lToCtx,
		attribute.String("service.layer", "gold-layer"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize open telemetry logger: %w", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	_ = cfg // Placeholder to avoid unused variable error

	app := &App{logger: log}

	return app, nil
}

// Start starts all application services.
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting gold worker...")

	a.logger.Info("Gold worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info("Shutting down gold worker...")

	a.logger.Info("Gold worker shutdown complete")
}
