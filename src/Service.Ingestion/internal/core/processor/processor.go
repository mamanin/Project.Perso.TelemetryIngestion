package processor

import (
	"context"

	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
)

// Processor manages the processing of MetricEvent messages using a pool of workers.
type Processor struct {
	logger     logger.Logger
	subscriber messaging.Subscriber
	pool       *Pool
}

// Config holds the configuration for the Processor.
type Config struct {
	// Workers defines the number of concurrent workers to process messages.
	Workers int
}

// NewProcessor creates a new Processor.
func NewProcessor(cfg Config, logger logger.Logger, subscriber messaging.Subscriber, wb WorkerBuilder) *Processor {
	pool := NewPool(
		cfg.Workers,
		wb,
	)

	return &Processor{
		logger:     logger,
		subscriber: subscriber,
		pool:       pool,
	}
}

// Start begins processing messages from the subscriber using the worker pool.
func (p *Processor) Start(ctx context.Context) {
	p.pool.Start(ctx)
	p.subscriber.Subscribe(ctx, p.pool.Queue())
}

// Stop stops the processor and cleans up resources.
func (p *Processor) Stop(ctx context.Context) error {
	if err := p.subscriber.Close(ctx); err != nil {
		p.logger.Error(err, "Error shutting down subscriber: %v", err)
	}
	p.pool.Stop()

	return nil
}
