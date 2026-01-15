package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/observability/logger"
	"service.ingestion/external/storage"
	"service.ingestion/internal/bronze"
	"service.ingestion/internal/core/processor"
)

// App represents the entire application with all its dependencies.
type App struct {
	logger       logger.Logger
	orchestrator *processor.Processor
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (*App, error) {
	lToCtx, bCancel := context.WithTimeout(ctx, 10*time.Second)
	defer bCancel()

	log, err := logger.NewOtelLogger(
		lToCtx,
		attribute.String("service.layer", "bronze-layer"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize open telemetry logger: %w", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	_ = cfg // Placeholder to avoid unused variable error.

	app := &App{logger: log}

	batchSize := 50
	workers := 4

	if err = app.initializeOrchestrator(batchSize, workers); err != nil {
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
func (a *App) initializeOrchestrator(batchSize int, workers int) error { // TODO: pass conf
	p, s, err := a.initializeMessaging(batchSize, workers)
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	v2h := bronze.NewV2Handler(a.logger, p)
	v1h := bronze.NewV1Handler(a.logger, p)
	lh := bronze.NewLegacyHandler(a.logger, p)

	o := processor.NewProcessor(
		processor.Config{
			Workers: workers,
		},
		a.logger,
		s,
		func(i int) processor.Worker {
			return bronze.NewWorker(i, batchSize, a.logger, v2h, v1h, lh)
		},
	)

	a.orchestrator = o
	return nil
}

// initializeMessaging sets up the messaging services.
func (a *App) initializeMessaging(batchSize int, workers int) (*eventhub.Publisher, *eventhub.Subscriber, error) { // TODO: pass conf
	pwcs := strings.TrimSpace(os.Getenv("PARTITION_WORKERS_CONNECTIONSTRING"))
	pwbcn := strings.TrimSpace(os.Getenv("PARTITION_WORKERS_BLOBCONTAINERNAME"))

	cp, err := storage.NewCheckpoint(storage.Config{
		ConnectionString: pwcs,
		ContainerName:    pwbcn,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	echcs := strings.TrimSpace(os.Getenv("ConnectionStrings__pocitpevh001"))
	tre := strings.TrimSpace(os.Getenv("TELEMETRY_RAW_EVENTHUBNAME"))
	tme := strings.TrimSpace(os.Getenv("TELEMETRY_METRICS_EVENTHUBNAME"))

	publisher, err := eventhub.NewPublisher(eventhub.Config{
		ConnectionString: echcs,
		EventHubName:     tme,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub publisher: %w", err)
	}

	subscriber, err := eventhub.NewSubscriber(
		a.logger,
		eventhub.SubscriberConfig{
			Config: eventhub.Config{
				ConnectionString: echcs,
				EventHubName:     tre,
			},
			BatchSize:    batchSize,
			PrefetchSize: int32(batchSize * (workers + workers/2)),
		},
		cp,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	return publisher, subscriber, nil
}
