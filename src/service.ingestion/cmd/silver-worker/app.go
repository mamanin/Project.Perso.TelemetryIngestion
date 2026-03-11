package main

import (
	"context"
	"fmt"
	"time"

	"service.ingestion/external/cache/redis"
	"service.ingestion/external/credential"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/observability/probes"
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

	logger        logger.Logger
	cred          credential.AzureCredentials
	orchestrator  *processor.Processor
	redis         *redis.Client
	subscriber    *eventhub.Subscriber
	probesManager *probes.Manager
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (AppManager, error) {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log, err := logger.NewOtelLogger(tCtx)
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

	if err = app.initializeCredentials(); err != nil {
		return nil, fmt.Errorf("failed to initialize credentials: %w", err)
	}

	if err = app.initializeRedis(); err != nil {
		return nil, fmt.Errorf("failed to initialize redis client: %w", err)
	}

	if err = app.initializeOrchestrator(); err != nil {
		return nil, fmt.Errorf("failed to initialize processor: %w", err)
	}

	if err = app.initializeProbesManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize health manager: %w", err)
	}

	return app, nil
}

// Start starts all application services.
func (a *App) Start(ctx context.Context) error {
	a.logger.Info("Starting silver worker...")

	if err := a.probesManager.StartProbes(ctx); err != nil {
		return fmt.Errorf("failed to start health probes: %w", err)
	}

	a.orchestrator.Start(ctx)

	if err := a.probesManager.MarkProbeAs(probes.StartupProbeName, probes.ProbeStatusReady); err != nil {
		a.logger.Error(err, "Unable to mark startup probe as ready: %v", err)
	}

	a.logger.Info("Silver worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info("Shutting down silver worker...")

	if err := a.probesManager.StopProbes(); err != nil {
		a.logger.Error(err, "Error stopping health manager: %v", err)
	}

	tCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	if err := a.orchestrator.Stop(tCtx); err != nil {
		a.logger.Error(err, "Error shutting down subscriber: %v", err)
	}

	if err := a.redis.Close(); err != nil {
		a.logger.Error(err, "Error closing redis client: %v", err)
	}

	a.logger.Info("Silver worker shutdown complete")
}

// initializeCredentials retrieves Azure credentials for the application.
func (a *App) initializeCredentials() error {
	cred, err := credential.NewAzureDefault()
	if err != nil {
		return fmt.Errorf("failed to get azure credentials: %w", err)
	}

	a.cred = cred
	return nil
}

// initializeRedis sets up the redis cache.
func (a *App) initializeRedis() error {
	r, err := redis.NewRedis(a.cfg.Redis)

	if err != nil {
		return fmt.Errorf("failed to initialize redis client: %w", err)
	}

	a.redis = r
	return nil
}

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *App) initializeOrchestrator() error {
	p, s, err := a.initializeMessaging()
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	o := processor.NewProcessor(a.cfg.Processor, a.logger, s, func(i int) processor.Worker {
		return silver.NewWorker(i, a.cfg.EventHub.Subscriber.BatchSize, a.logger, silver.NewHandler(a.logger, a.redis, p))
	})

	a.orchestrator = o
	return nil
}

// initializeMessaging sets up the messaging services.
func (a *App) initializeMessaging() (*eventhub.Publisher, *eventhub.Subscriber, error) {
	cp, err := container.NewCheckpoint(a.cfg.Container, a.cred)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	p, err := eventhub.NewPublisher(a.cfg.EventHub, a.cred)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub publisher: %w", err)
	}

	s, err := eventhub.NewSubscriber(a.logger, a.cfg.EventHub, a.cred, cp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	a.subscriber = s
	return p, s, nil
}

// initializeProbesManager sets up the health check manager.
func (a *App) initializeProbesManager() error {
	var err error

	pm, err := probes.NewManager([]probes.Checker{a.subscriber}, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create health manager: %w", err)
	}

	a.probesManager = pm
	return nil
}
