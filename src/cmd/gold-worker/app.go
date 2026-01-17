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
	"service.ingestion/external/storage/adx"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/processor"
	"service.ingestion/internal/gold"
)

// App represents the entire application with all its dependencies.
type App struct {
	logger       logger.Logger
	orchestrator *processor.Processor
	adx          *adx.Client[core.DataMetric]
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (*App, error) {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log, err := logger.NewOtelLogger(
		tCtx,
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

	batchSize := 500
	workers := 2

	if err = app.initializeAdx(); err != nil {
		return nil, fmt.Errorf("failed to initialize ADX client: %w", err)
	}

	if err = app.initializeOrchestrator(batchSize, workers); err != nil {
		return nil, fmt.Errorf("failed to initialize processor: %w", err)
	}

	return app, nil
}

// Start starts all application services.
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting gold worker...")

	a.orchestrator.Start(ctx)

	a.logger.Info("Gold worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info("Shutting down gold worker...")

	tCtx, cancel := context.WithTimeout(ctx, 100*time.Second)
	defer cancel()

	if err := a.orchestrator.Stop(tCtx); err != nil {
		a.logger.Error(err, "Error shutting down subscriber: %v", err)
	}

	if err := a.adx.Close(); err != nil {
		a.logger.Error(err, "Error closing ADX client: %v", err)
	}
	a.logger.Info("Gold worker shutdown complete")
}

func (a *App) initializeAdx() error {
	adxCS := strings.TrimSpace(os.Getenv("TELEMETRIES_URI"))
	adxDB := strings.TrimSpace(os.Getenv("TELEMETRIES_DATABASENAME"))
	adxTable := "metrics"

	c, err := adx.NewClient[core.DataMetric](&adx.Config{
		Endpoint: adxCS,
		Database: adxDB,
		Table:    adxTable,
	}, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize ADX client: %w", err)
	}

	a.adx = c
	return nil
}

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *App) initializeOrchestrator(batchSize int, workers int) error { // TODO: pass conf
	s, err := a.initializeMessaging(batchSize, workers)
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	h := gold.NewHandler(a.logger, a.adx)

	o := processor.NewProcessor(
		processor.Config{
			Workers: workers,
		},
		a.logger,
		s,
		func(i int) processor.Worker {
			return gold.NewWorker(i, batchSize, a.logger, h)
		},
	)

	a.orchestrator = o
	return nil
}

// initializeMessaging sets up the messaging services.
func (a *App) initializeMessaging(batchSize int, workers int) (*eventhub.Subscriber, error) { // TODO: pass conf
	pwcs := strings.TrimSpace(os.Getenv("PARTITION_WORKERS_CONNECTIONSTRING"))
	pwbcn := strings.TrimSpace(os.Getenv("PARTITION_WORKERS_BLOBCONTAINERNAME"))

	cp, err := container.NewCheckpoint(container.Config{
		ConnectionString: pwcs,
		ContainerName:    pwbcn,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	echcs := strings.TrimSpace(os.Getenv("ConnectionStrings__pocitpevh001"))
	tde := strings.TrimSpace(os.Getenv("TELEMETRY_DATA_EVENTHUBNAME"))

	s, err := eventhub.NewSubscriber(
		a.logger,
		eventhub.SubscriberConfig{
			Config: eventhub.Config{
				ConnectionString: echcs,
				EventHubName:     tde,
			},
			BatchSize:    batchSize,
			PrefetchSize: int32(batchSize * (workers + workers/2)),
		},
		cp,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	return s, nil
}
