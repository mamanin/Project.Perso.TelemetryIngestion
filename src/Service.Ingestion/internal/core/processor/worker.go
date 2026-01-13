package processor

import (
	"context"

	"service.ingestion/external/messaging"
)

// Worker defines the interface for a worker that processes metric batches.
type Worker interface {
	// Process starts the worker to process batches from the queue.
	Process(ctx context.Context, queue chan messaging.RawEventBatch)
}

// WorkerBuilder is a function type that creates a new Worker instance.
type WorkerBuilder func(id int) Worker
