package main

import (
	"context"
	"fmt"
	"time"

	"service.ingestion/external/credential"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/bronze"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/observability/probes"
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

	logger        logger.Logger
	cred          credential.AzureCredentials
	orchestrator  *processor.Processor
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
	a.logger.Info("Starting bronze worker...")

	if err := a.probesManager.StartProbes(ctx); err != nil {
		return fmt.Errorf("failed to start health probes: %w", err)
	}

	a.orchestrator.Start(ctx)

	if err := a.probesManager.MarkProbeAs(probes.StartupProbeName, probes.ProbeStatusReady); err != nil {
		a.logger.Error(err, "Unable to mark startup probe as ready: %v", err)
	}

	a.logger.Info("Bronze worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info("Shutting down bronze worker...")

	if err := a.probesManager.StopProbes(); err != nil {
		a.logger.Error(err, "Error stopping health manager: %v", err)
	}

	tCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	if err := a.orchestrator.Stop(tCtx); err != nil {
		a.logger.Error(err, "Error shutting down subscriber: %v", err)
	}

	a.logger.Info("Bronze worker shutdown complete")
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

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *App) initializeOrchestrator() error {
	p, s, err := a.initializeMessaging()
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	o := processor.NewProcessor(a.cfg.Processor, a.logger, s, func(i int) processor.Worker {
		return bronze.NewWorker(i, a.logger, []processor.VersionAdapter{
			processor.NewHandlerAdapter(pkg.V2, a.cfg.EventHub.Subscriber.BatchSize, bronze.NewV2Handler(a.logger, p)),
			processor.NewHandlerAdapter(pkg.V1, a.cfg.EventHub.Subscriber.BatchSize, bronze.NewV1Handler(a.logger, p)),
		}, processor.NewHandlerAdapter(pkg.Legacy, a.cfg.EventHub.Subscriber.BatchSize, bronze.NewLegacyHandler(a.logger, p)))
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
