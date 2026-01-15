package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"service.ingestion/external/cache"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/observability/logger"
	"service.ingestion/external/storage"
	"service.ingestion/internal/core/processor"
	"service.ingestion/internal/silver"
)

// App represents the entire application with all its dependencies.
type App struct {
	logger       logger.Logger
	orchestrator *processor.Processor
	redis        *cache.Redis
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (*App, error) {
	tCtx, tCancel := context.WithTimeout(ctx, 10*time.Second)
	defer tCancel()

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
	_ = cfg // Placeholder to avoid unused variable error.

	app := &App{logger: log}

	batchSize := 200
	workers := 4

	app.initializeRedis(workers)

	if err = app.initializeOrchestrator(batchSize, workers); err != nil {
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
func (a *App) initializeOrchestrator(batchSize int, workers int) error { // TODO: pass conf
	p, s, err := a.initializeMessaging(batchSize, workers)
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	h := silver.NewHandler(a.logger, a.redis, p)

	o := processor.NewProcessor(
		processor.Config{
			Workers: workers,
		},
		a.logger,
		s,
		func(i int) processor.Worker {
			return silver.NewWorker(i, batchSize, a.logger, h)
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
	tme := strings.TrimSpace(os.Getenv("TELEMETRY_METRICS_EVENTHUBNAME"))
	tde := strings.TrimSpace(os.Getenv("TELEMETRY_DATA_EVENTHUBNAME"))

	publisher, err := eventhub.NewPublisher(eventhub.Config{
		ConnectionString: echcs,
		EventHubName:     tde,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub publisher: %w", err)
	}

	subscriber, err := eventhub.NewSubscriber(
		a.logger,
		eventhub.SubscriberConfig{
			Config: eventhub.Config{
				ConnectionString: echcs,
				EventHubName:     tme,
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

// initializeRedis sets up the Redis client.
func (a *App) initializeRedis(workers int) { // TODO: pass conf
	hr := strings.TrimSpace(os.Getenv("POCITPRED001_HOST"))
	pr := strings.TrimSpace(os.Getenv("POCITPRED001_PORT"))
	pwdr := strings.TrimSpace(os.Getenv("POCITPRED001_PASSWORD"))

	r := cache.NewRedis(&cache.Config{
		Host:        hr,
		Port:        func() int { p, _ := strconv.Atoi(pr); return p }(),
		Password:    pwdr,
		Connections: workers,
	})

	a.redis = r
}
