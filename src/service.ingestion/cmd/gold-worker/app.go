package main

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/adx"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/observability/probes"
	"service.ingestion/internal/core/processor"
	"service.ingestion/internal/gold"
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
	orchestrator  *processor.Processor
	adx           *adx.Client[core.DataMetric]
	subscriber    *eventhub.Subscriber
	probesManager *probes.Manager
}

// NewApp creates and initializes a new App instance with all dependencies.
func NewApp(ctx context.Context) (AppManager, error) {
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

	app := &App{
		cfg:    cfg,
		logger: log,
	}

	if err = app.initializeAdx(); err != nil {
		return nil, fmt.Errorf("failed to initialize ADX client: %w", err)
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
	a.logger.Info("Starting gold worker...")

	if err := a.probesManager.StartProbes(ctx); err != nil {
		return fmt.Errorf("failed to start health probes: %w", err)
	}

	a.orchestrator.Start(ctx)

	if err := a.probesManager.MarkProbeAs(probes.StartupProbeName, probes.ProbeStatusReady); err != nil {
		a.logger.Error(err, "Unable to mark startup probe as ready: %v", err)
	}

	a.logger.Info("Gold worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *App) Stop(ctx context.Context) {
	a.logger.Info("Shutting down gold worker...")

	if err := a.probesManager.StopProbes(); err != nil {
		a.logger.Error(err, "Error stopping health manager: %v", err)
	}

	tCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
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
	c, err := adx.NewClient[core.DataMetric](adx.Config{}, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize ADX client: %w", err)
	}

	a.adx = c
	return nil
}

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *App) initializeOrchestrator() error {
	s, err := a.initializeMessaging()
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	h := gold.NewHandler(a.logger, a.adx)

	o := processor.NewProcessor(
		processor.Config{
			Workers: 0,
		},
		a.logger,
		s,
		func(i int) processor.Worker {
			return gold.NewWorker(i, 0, a.logger, h)
		},
	)

	a.orchestrator = o
	return nil
}

// initializeMessaging sets up the messaging services.
func (a *App) initializeMessaging() (*eventhub.Subscriber, error) {
	cp, err := container.NewCheckpoint(container.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	s, err := eventhub.NewSubscriber(a.logger, eventhub.SubscriberConfig{}, cp)
	if err != nil {
		return nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	a.subscriber = s
	return s, nil
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
