package silver

import (
	"context"

	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/processor"
)

// Processor manages the processing of MetricEvent messages using a pool of workers.
type Processor struct {
	logger     logger.Logger
	subscriber messaging.Subscriber
	pool       *processor.Pool
}

// Config holds the configuration for the Processor.
type Config struct {
	// Workers defines the number of concurrent workers to process messages.
	Workers int
	// BatchSize defines the number of messages each worker processes in a batch.
	BatchSize int
}

// NewProcessor creates a new Processor.
func NewProcessor(cfg Config, logger logger.Logger, subscriber messaging.Subscriber, handler messaging.BatchHandler[core.MetricEvent]) *Processor {
	pool := processor.NewPool(
		cfg.Workers,
		func(i int) processor.Worker {
			return NewWorker(i, logger, handler, cfg.BatchSize)
		},
	)

	return &Processor{
		logger:     logger,
		subscriber: subscriber,
		pool:       pool,
	}
}

// Start begins processing messages from the subscriber using the worker pool.
func (o *Processor) Start(ctx context.Context) {
	o.pool.Start(ctx)
	o.subscriber.Subscribe(ctx, o.pool.Queue())
}

// Stop stops the processor and cleans up resources.
func (o *Processor) Stop(ctx context.Context) error {
	if err := o.subscriber.Close(ctx); err != nil {
		o.logger.Error(err, "Error shutting down subscriber: %v", err)
	}
	o.pool.Stop()

	return nil
}
