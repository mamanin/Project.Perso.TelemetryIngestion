package processor

import (
	"context"

	"github.com/mailru/easyjson"
)

// VersionAdapter defines the interface for version-specific item handling.
type VersionAdapter interface {
	// Version returns the version string this handler supports.
	Version() string

	// RegisterItem parses and appends the item to the batch.
	RegisterItem(data []byte)

	// Handle processes the collected batch of items.
	Handle(ctx context.Context)
}

// HandlerAdapter implements VersionAdapter for a specific item type T.
type HandlerAdapter[T any] struct {
	version string

	items   []*T
	handler Handler[T]
}

// NewHandlerAdapter creates a new HandlerAdapter for the given version and processor.
func NewHandlerAdapter[T any](version string, batchSize int, handler Handler[T]) VersionAdapter {
	var item T
	if _, ok := any(&item).(easyjson.Unmarshaler); !ok {
		panic("Type T must implement easyjson.Unmarshaler : (i *T) UnmarshalEasyJSON(l *jlexer.Lexer)")
	}

	return &HandlerAdapter[T]{
		version: version,

		items:   make([]*T, 0, batchSize),
		handler: handler,
	}
}

// Version returns the version string this handler supports.
func (h *HandlerAdapter[T]) Version() string {
	return h.version
}

// RegisterItem parses and appends the item to the batch.
func (h *HandlerAdapter[T]) RegisterItem(data []byte) {
	var item T
	if err := easyjson.Unmarshal(data, any(&item).(easyjson.Unmarshaler)); err == nil {
		h.items = append(h.items, &item)
	}
}

// Handle processes the collected batch of items.
func (h *HandlerAdapter[T]) Handle(ctx context.Context) {
	if len(h.items) > 0 {
		h.handler.Handle(ctx, h.items)
		h.items = h.items[:0]
	}
}
