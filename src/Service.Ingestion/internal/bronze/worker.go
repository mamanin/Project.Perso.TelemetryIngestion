package bronze

import (
	"context"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
	"service.ingestion/internal/core/processor"
	"service.ingestion/pkg"
)

// Worker represents a single worker that processes telemetry batches.
type Worker struct {
	id   int
	pool sync.Pool
	wg   sync.WaitGroup

	logger          logger.Logger
	versionHandlers map[string]processor.HandlerAdapter
	legacyHandler   processor.HandlerAdapter
}

// NewWorker creates a new Worker instance.
func NewWorker(id int, batchSize int, logger logger.Logger, handlers []processor.HandlerAdapter, legacyHandler processor.HandlerAdapter) processor.Worker {
	poolFactory := func() any {
		m := make(map[string]any)
		for _, h := range handlers {
			m[h.Version()] = h.CreateBatch(batchSize)
		}
		m[legacyHandler.Version()] = legacyHandler.CreateBatch(batchSize)
		return m
	}

	hm := make(map[string]processor.HandlerAdapter)
	for _, h := range handlers {
		hm[h.Version()] = h
	}

	return &Worker{
		id: id,
		pool: sync.Pool{
			New: poolFactory,
		},

		logger:          logger,
		versionHandlers: hm,
		legacyHandler:   legacyHandler,
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
	pool := w.pool.Get().(map[string]any)

	defer func() {
		for _, h := range w.versionHandlers {
			h.ClearBatch(pool)
		}
		w.legacyHandler.ClearBatch(pool)
		w.pool.Put(pool)
	}()
	defer w.wg.Wait()

	for _, bytes := range batch {
		var e pkg.VersionDiscriminant

		if err := easyjson.Unmarshal(bytes, &e); err == nil {
			if h, ok := w.versionHandlers[e.Version]; ok {
				h.UnmarshalAndAppend(bytes, pool)
				continue
			}
		}

		w.legacyHandler.UnmarshalAndAppend(bytes, pool)
	}

	for _, h := range w.versionHandlers {
		w.wg.Go(func() {
			h.Handle(ctx, pool)
		})
	}

	w.wg.Go(func() {
		w.legacyHandler.Handle(ctx, pool)
	})

	w.logger.Info("W%d|%.2fμs|%de", w.id, float64(time.Since(start).Microseconds()), len(batch))
}
