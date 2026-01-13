package messaging

import (
	"context"
)

// RawEventBatch represents a batch of raw event data, where each event is a byte slice.
type RawEventBatch = [][]byte

// RawBatchHandler defines a function type that processes a batch of raw events.
type RawBatchHandler func(ctx context.Context, batch RawEventBatch)

// BatchHandler defines a generic function type that processes a batch of events of type T.
type BatchHandler[T any] func(ctx context.Context, batch []*T)
