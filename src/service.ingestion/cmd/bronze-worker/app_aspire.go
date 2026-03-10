package main

import (
	"context"
	"fmt"
	"time"

	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/bronze"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/observability/probes"
	"service.ingestion/internal/core/processor"
	"service.ingestion/pkg"
)

// AspireApp represents the entire application with all its dependencies.
type AspireApp struct {
	cfg *AspireConfig

	logger        logger.Logger
	orchestrator  *processor.Processor
	subscriber    *eventhub.Subscriber
	probesManager *probes.Manager
}

// NewAspireApp creates and initializes a new AspireApp instance with all dependencies.
func NewAspireApp(ctx context.Context) (AppManager, error) {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log, err := logger.NewOtelLogger(tCtx)
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

	if err = app.initializeOrchestrator(); err != nil {
		return nil, fmt.Errorf("failed to initialize processor: %w", err)
	}

	if err = app.initializeProbesManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize health manager: %w", err)
	}

	return app, nil
}

// Start starts all application services.
func (a *AspireApp) Start(ctx context.Context) error {
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
func (a *AspireApp) Stop(ctx context.Context) {
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

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *AspireApp) initializeOrchestrator() error {
	p, s, err := a.initializeMessaging()
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	o := processor.NewProcessor(a.cfg.processor, a.logger, s, func(i int) processor.Worker {
		return bronze.NewWorker(i, a.logger, []processor.VersionAdapter{
			processor.NewHandlerAdapter(pkg.V2, a.cfg.eventhubSubscriber.BatchSize, bronze.NewV2Handler(a.logger, p)),
			processor.NewHandlerAdapter(pkg.V1, a.cfg.eventhubSubscriber.BatchSize, bronze.NewV1Handler(a.logger, p)),
		}, processor.NewHandlerAdapter(pkg.Legacy, a.cfg.eventhubSubscriber.BatchSize, bronze.NewLegacyHandler(a.logger, p)))
	})

	a.orchestrator = o
	return nil
}

// initializeMessaging sets up the messaging services.
func (a *AspireApp) initializeMessaging() (*eventhub.Publisher, *eventhub.Subscriber, error) {
	cp, err := container.NewCheckpointForAspire(a.cfg.container)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	p, err := eventhub.NewPublisherForAspire(a.cfg.eventhubPublisher)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub publisher: %w", err)
	}

	s, err := eventhub.NewSubscriberForAspire(a.logger, a.cfg.eventhubSubscriber, cp)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	a.subscriber = s
	return p, s, nil
}

// initializeProbesManager sets up the health check manager.
func (a *AspireApp) initializeProbesManager() error {
	var err error

	pm, err := probes.NewManager([]probes.Checker{a.subscriber}, a.logger)
	if err != nil {
		return fmt.Errorf("failed to create health manager: %w", err)
	}

	a.probesManager = pm
	return nil
}
