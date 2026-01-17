package processor

import (
	"context"

	"github.com/mailru/easyjson"
)

// HandlerAdapter defines the interface for version-specific item handling.
type HandlerAdapter interface {
	// Version returns the version string this handler supports.
	Version() string

	// UnmarshalAndAppend parses the data and, if successful, appends the item to the batch using the pool.
	UnmarshalAndAppend(data []byte, pool map[string]any)

	// Handle processes the collected batch of items.
	Handle(ctx context.Context, pool map[string]any)

	// CreateBatch creates a new empty batch slice for this handler.
	CreateBatch(capacity int) any

	// ClearBatch resets the batch slice in the pool to zero length for reuse.
	ClearBatch(pool map[string]any)
}

// GenericHandler implements HandlerAdapter for a specific item type T.
type GenericHandler[T any] struct {
	version string
	handler Handler[T]
}

// NewHandlerAdapter creates a new GenericHandler for the given version and processor.
func NewHandlerAdapter[T any](version string, handler Handler[T]) *GenericHandler[T] {
	var item T
	if _, ok := any(&item).(easyjson.Unmarshaler); !ok {
		panic("Type T must implement easyjson.Unmarshaler : (i *T) UnmarshalEasyJSON(l *jlexer.Lexer)")
	}

	return &GenericHandler[T]{
		version: version,
		handler: handler,
	}
}

// Version returns the version string this handler supports.
func (h *GenericHandler[T]) Version() string {
	return h.version
}

// UnmarshalAndAppend parses the data and, if successful, appends the item to the batch using the pool.
func (h *GenericHandler[T]) UnmarshalAndAppend(data []byte, pool map[string]any) {
	var item T
	if err := easyjson.Unmarshal(data, any(&item).(easyjson.Unmarshaler)); err == nil {
		if slice, ok := pool[h.version].([]*T); ok {
			pool[h.version] = append(slice, &item)
		}
	}
}

// Handle processes the collected batch of items.
func (h *GenericHandler[T]) Handle(ctx context.Context, pool map[string]any) {
	if slice, ok := pool[h.version].([]*T); ok && len(slice) > 0 {
		h.handler.Handle(ctx, slice)
	}
}

// CreateBatch creates a new empty batch slice for this handler.
func (h *GenericHandler[T]) CreateBatch(capacity int) any {
	return make([]*T, 0, capacity)
}

// ClearBatch resets the batch slice in the pool to zero length for reuse.
func (h *GenericHandler[T]) ClearBatch(pool map[string]any) {
	if slice, ok := pool[h.version].([]*T); ok {
		pool[h.version] = slice[:0]
	}
}
