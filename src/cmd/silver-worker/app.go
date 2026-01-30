package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"service.ingestion/external/cache/redis"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/processor"
	"service.ingestion/internal/silver"
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
	redis        *redis.Client
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (AppManager, error) {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log, err := logger.NewOtelLogger(
		tCtx,
		attribute.String("service.layer", "silver"),
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

	app.initializeRedis()

	if err = app.initializeOrchestrator(); err != nil {
		return nil, fmt.Errorf("failed to initialize processor: %w", err)
	}

	return app, nil
}

// Start starts all application services.
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting silver worker...")

	a.orchestrator.Start(ctx)

	a.logger.Info("Silver worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info("Shutting down silver worker...")

	tCtx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	if err := a.orchestrator.Stop(tCtx); err != nil {
		a.logger.Error(err, "Error shutting down subscriber: %v", err)
	}

	if err := a.redis.Close(); err != nil {
		a.logger.Error(err, "Error closing redis client: %v", err)
	}

	a.logger.Info("Silver worker shutdown complete")
}

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *App) initializeOrchestrator() error {
	p, s, err := a.initializeMessaging()
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	h := silver.NewHandler(a.logger, a.redis, p)

	o := processor.NewProcessor(
		processor.Config{
			Workers: 0,
		},
		a.logger,
		s,
		func(i int) processor.Worker {
			return silver.NewWorker(i, 0, a.logger, h)
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

	s, err := eventhub.NewSubscriber(a.logger, eventhub.SubscriberConfig{}, cp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	return p, s, nil
}

// initializeRedis sets up the redis cache.
func (a *App) initializeRedis() {
	r := redis.NewRedis(redis.Config{})

	a.redis = r
}
