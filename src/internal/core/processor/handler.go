package processor

import (
	"context"
)

// Handler defines the interface for processing batches of events of type T.
type Handler[T any] interface {
	// Handle processes a batch of events of type T.
	Handle(ctx context.Context, batch []*T)
}
