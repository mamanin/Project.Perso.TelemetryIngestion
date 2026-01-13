package silver

import (
	"context"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/processor"
)

// Worker represents a single worker that processes metric batches.
type Worker struct {
	id   int
	pool sync.Pool

	logger  logger.Logger
	handler messaging.BatchHandler[core.MetricEvent]
}

// NewWorker creates a new Worker instance.
func NewWorker(id int, logger logger.Logger, handler messaging.BatchHandler[core.MetricEvent], batchSize int) processor.Worker {
	return &Worker{
		id: id,
		pool: sync.Pool{
			New: func() any {
				return make([]*core.MetricEvent, 0, batchSize)
			},
		},

		logger:  logger,
		handler: handler,
	}
}

// Process starts the worker to process batches from the queue.
func (w *Worker) Process(ctx context.Context, queue chan messaging.RawEventBatch) {
	for batch := range queue {
		w.processBatch(ctx, batch)
	}
}

// processBatch processes a single batch of raw events.
func (w *Worker) processBatch(ctx context.Context, batch messaging.RawEventBatch) {
	eventPool := w.pool.Get().([]*core.MetricEvent)
	defer w.pool.Put(eventPool[:0])

	start := time.Now()

	for _, bytes := range batch {
		var e core.MetricEvent
		if err := json.Unmarshal(bytes, &e); err != nil {
			continue
		}

		eventPool = append(eventPool, &e)
	}

	w.handler(ctx, eventPool)

	w.logger.Info("W%d: %.2fμs: %d", w.id, float64(time.Since(start).Microseconds())/float64(len(eventPool)), len(eventPool))
}
