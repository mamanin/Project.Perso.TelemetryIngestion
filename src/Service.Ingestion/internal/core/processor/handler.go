package processor

import (
	"context"
)

type Handler[T any] interface {
	Handle(ctx context.Context, batch []*T)
}
