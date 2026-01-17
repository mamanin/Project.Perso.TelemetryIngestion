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

	logger        logger.Logger
	legacyHandler processor.Handler[pkg.TelemetryLegacyEvent]
	v1Handler     processor.Handler[pkg.TelemetryV1Event]
	v2Handler     processor.Handler[pkg.TelemetryV2Event]
}

// NewWorker creates a new Worker instance.
func NewWorker(id int, batchSize int, logger logger.Logger, v2h processor.Handler[pkg.TelemetryV2Event], v1h processor.Handler[pkg.TelemetryV1Event], lh processor.Handler[pkg.TelemetryLegacyEvent]) processor.Worker {
	return &Worker{
		id: id,
		pool: sync.Pool{
			New: func() any {
				return map[string]any{
					pkg.V2:     make([]*pkg.TelemetryV2Event, 0, batchSize),
					pkg.V1:     make([]*pkg.TelemetryV1Event, 0, batchSize),
					pkg.Legacy: make([]*pkg.TelemetryLegacyEvent, 0, batchSize),
				}
			},
		},

		logger:        logger,
		v2Handler:     v2h,
		v1Handler:     v1h,
		legacyHandler: lh,
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

	// TODO: try refactoring
	eventPool := w.pool.Get().(map[string]any)
	v2Pool := eventPool[pkg.V2].([]*pkg.TelemetryV2Event)
	v1Pool := eventPool[pkg.V1].([]*pkg.TelemetryV1Event)
	legacyPool := eventPool[pkg.Legacy].([]*pkg.TelemetryLegacyEvent)

	defer func() {
		eventPool[pkg.V2] = v2Pool[:0]
		eventPool[pkg.V1] = v1Pool[:0]
		eventPool[pkg.Legacy] = legacyPool[:0]
		w.pool.Put(eventPool)
	}()
	defer w.wg.Wait()

	for _, bytes := range batch {
		var e pkg.VersionDiscriminant

		if err := easyjson.Unmarshal(bytes, &e); err == nil {
			switch e.Version {
			case pkg.V2:
				var ev2 pkg.TelemetryV2Event
				if err = easyjson.Unmarshal(bytes, &ev2); err == nil {
					v2Pool = append(v2Pool, &ev2)
				}
			case pkg.V1:
				var ev1 pkg.TelemetryV1Event
				if err = easyjson.Unmarshal(bytes, &ev1); err == nil {
					v1Pool = append(v1Pool, &ev1)
				}
			}
		}

		var le pkg.TelemetryLegacyEvent
		if err := easyjson.Unmarshal(bytes, &le); err != nil {
			continue
		}
		legacyPool = append(legacyPool, &le)
	}

	w.wg.Go(func() {
		w.v2Handler.Handle(ctx, v2Pool)
	})

	w.wg.Go(func() {
		w.v1Handler.Handle(ctx, v1Pool)
	})

	w.wg.Go(func() {
		w.legacyHandler.Handle(ctx, legacyPool)
	})

	w.logger.Info("W%d|%.2fμs|%de", w.id, float64(time.Since(start).Microseconds()), len(batch))
}
