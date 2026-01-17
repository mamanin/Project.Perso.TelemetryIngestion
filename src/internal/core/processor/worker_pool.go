package processor

import (
	"context"
	"sync"

	"service.ingestion/external/messaging"
)

// Pool manages a pool of workers to process raw events concurrently.
type Pool struct {
	wg    sync.WaitGroup
	queue chan messaging.RawEventBatch

	workers       int
	workerBuilder WorkerBuilder
}

// NewPool creates a new Pool with the specified number of workers.
func NewPool(workers int, wb WorkerBuilder) *Pool {
	return &Pool{
		queue: make(chan messaging.RawEventBatch, workers*2),

		workers:       workers,
		workerBuilder: wb,
	}
}

// Start initializes the worker pool and begins processing events.
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		worker := p.workerBuilder(i)

		p.wg.Go(func() {
			worker.Process(ctx, p.queue)
		})
	}
}

// Stop gracefully shuts down the worker pool.
func (p *Pool) Stop() {
	p.wg.Wait()
	close(p.queue)
}

// Queue returns the channel used for queuing raw events.
func (p *Pool) Queue() chan messaging.RawEventBatch {
	return p.queue
}
