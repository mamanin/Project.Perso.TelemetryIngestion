package bronze

import (
	"context"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	"service.ingestion/external/messaging"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/processor"
	"service.ingestion/pkg"
)

// Worker represents a single worker that processes telemetry batches.
type Worker struct {
	id int
	wg sync.WaitGroup

	logger        logger.Logger
	adapters      map[string]processor.VersionAdapter
	legacyAdapter processor.VersionAdapter
}

// NewWorker creates a new Worker instance.
func NewWorker(id int, logger logger.Logger, handlers []processor.VersionAdapter, legacyHandler processor.VersionAdapter) processor.Worker {
	hm := make(map[string]processor.VersionAdapter)
	for _, h := range handlers {
		hm[h.Version()] = h
	}

	return &Worker{
		id: id,
		wg: sync.WaitGroup{},

		logger:        logger,
		adapters:      hm,
		legacyAdapter: legacyHandler,
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

	for _, bytes := range batch {
		var e pkg.VersionDiscriminant

		if err := easyjson.Unmarshal(bytes, &e); err == nil {
			if h, ok := w.adapters[e.Version]; ok {
				h.RegisterItem(bytes)
				continue
			}
		}

		w.legacyAdapter.RegisterItem(bytes)
	}

	for _, h := range w.adapters {
		w.wg.Go(func() {
			h.Handle(ctx)
		})
	}

	w.wg.Go(func() {
		w.legacyAdapter.Handle(ctx)
	})

	w.wg.Wait()
	w.logger.Info("W%d|%.2fμs|%de", w.id, float64(time.Since(start).Microseconds()), len(batch))
}
