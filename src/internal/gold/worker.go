package gold

import (
	"context"
	"sync"
	"time"

	"github.com/mailru/easyjson"
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
	handler processor.Handler[core.MetricEvent]
}

// NewWorker creates a new Worker instance.
func NewWorker(id int, batchSize int, logger logger.Logger, h processor.Handler[core.MetricEvent]) processor.Worker {
	return &Worker{
		id: id,
		pool: sync.Pool{
			New: func() any {
				return make([]*core.MetricEvent, 0, batchSize)
			},
		},

		logger:  logger,
		handler: h,
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
	start := time.Now()

	eventPool := w.pool.Get().([]*core.MetricEvent)
	defer w.pool.Put(eventPool[:0])

	for _, bytes := range batch {
		var e core.MetricEvent
		if err := easyjson.Unmarshal(bytes, &e); err != nil {
			continue
		}

		eventPool = append(eventPool, &e)
	}

	w.handler.Handle(ctx, eventPool)

	w.logger.Info("W%d|%.2fμs|%de", w.id, float64(time.Since(start).Microseconds()), len(batch))
}
