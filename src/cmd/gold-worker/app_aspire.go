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
	"service.ingestion/internal/core/processor"
	"service.ingestion/internal/gold"
)

// AspireApp represents the entire application with all its dependencies.
type AspireApp struct {
	cfg *AspireConfig

	logger       logger.Logger
	orchestrator *processor.Processor
	adx          *adx.Client[core.DataMetric]
}

// NewAspireApp creates and initializes a new AspireApp instance with all dependencies.
func NewAspireApp(ctx context.Context) (AppManager, error) {
	tCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log, err := logger.NewOtelLogger(
		tCtx,
		attribute.String("service.layer", "gold-layer"),
	)
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

	if err = app.initializeAdx(); err != nil {
		return nil, fmt.Errorf("failed to initialize ADX client: %w", err)
	}

	if err = app.initializeOrchestrator(); err != nil {
		return nil, fmt.Errorf("failed to initialize processor: %w", err)
	}

	return app, nil
}

// Start starts all application services.
func (a *AspireApp) Start(ctx context.Context) error {
	a.logger.Info("Starting gold worker...")

	a.orchestrator.Start(ctx)

	a.logger.Info("Gold worker started successfully")
	return nil
}

// Stop gracefully shuts down all application services.
func (a *AspireApp) Stop(ctx context.Context) {
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

func (a *AspireApp) initializeAdx() error {
	c, err := adx.NewClientForAspire[core.DataMetric](a.cfg.adx, a.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize ADX client: %w", err)
	}

	a.adx = c
	return nil
}

// initializeOrchestrator sets up the orchestrator with its dependencies.
func (a *AspireApp) initializeOrchestrator() error {
	s, err := a.initializeMessaging()
	if err != nil {
		return fmt.Errorf("failed to initialize event hub subscriber: %w", err)
	}

	h := gold.NewHandler(a.logger, a.adx)

	o := processor.NewProcessor(a.cfg.processor, a.logger, s,
		func(i int) processor.Worker {
			return gold.NewWorker(i, 0, a.logger, h)
		},
	)

	a.orchestrator = o
	return nil
}

// initializeMessaging sets up the messaging services.
func (a *AspireApp) initializeMessaging() (*eventhub.Subscriber, error) {
	cp, err := container.NewCheckpointForAspire(a.cfg.container)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	s, err := eventhub.NewSubscriberForAspire(a.logger, a.cfg.eventhubSubscriber, cp)
	if err != nil {
		return nil, fmt.Errorf("failed to create event hub subscriber: %w", err)
	}

	return s, nil
}
