package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/bronze"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/processor"
	"service.ingestion/pkg"
)

// AppManager defines the interface for managing the application lifecycle.
type AppManager interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context)
}

// App represents the entire application with all its dependencies.
type App struct {
	cfg *Config

	logger       logger.Logger
	orchestrator *processor.Processor
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (AppManager, error) {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log, err := logger.NewOtelLogger(
		tCtx,
		attribute.String("service.layer", "bronze-layer"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize open telemetry logger: %w", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}

	app := &App{
		cfg:    cfg,
		logger: log,
	}

	if err = app.initializeOrchestrator(); err != nil {
		return nil, fmt.Errorf("failed to initialize processor: %w", err)
	}

	return app, nil
}

// Start starts all application services.
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting bronze worker...")

	a.orchestrator.Start(ctx)

	a.logger.Info("Bronze worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info("Shutting down bronze worker...")

	tCtx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	if err := a.orchestrator.Stop(tCtx); err != nil {
		a.logger.Error(err, "Error shutting down subscriber: %v", err)
	}

	a.logger.Info("Bronze worker shutdown complete")
}

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *App) initializeOrchestrator() error {
	p, s, err := a.initializeMessaging()
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	h2 := bronze.NewV2Handler(a.logger, p)
	h1 := bronze.NewV1Handler(a.logger, p)
	lh := bronze.NewLegacyHandler(a.logger, p)

	o := processor.NewProcessor(
		processor.Config{
			Workers: 0,
		},
		a.logger,
		s,
		func(i int) processor.Worker {
			return bronze.NewWorker(i, a.logger,
				[]processor.VersionAdapter{
					processor.NewHandlerAdapter(pkg.V2, 0, h2),
					processor.NewHandlerAdapter(pkg.V1, 0, h1),
				},
				processor.NewHandlerAdapter(pkg.Legacy, 0, lh),
			)
		},
	)

	a.orchestrator = o
	return nil
}

// initializeMessaging sets up the messaging services.
func (a *App) initializeMessaging() (*eventhub.Publisher, *eventhub.Subscriber, error) {
	cp, err := container.NewCheckpoint(container.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	p, err := eventhub.NewPublisher(eventhub.Config{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub publisher: %w", err)
	}

	s, err := eventhub.NewSubscriber(
		a.logger,
		eventhub.SubscriberConfig{
			Config:       eventhub.Config{},
			BatchSize:    0,
			PrefetchSize: 0,
		},
		cp,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	return p, s, nil
}
