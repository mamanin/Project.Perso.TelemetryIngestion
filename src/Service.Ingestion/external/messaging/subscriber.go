package messaging

import (
	"context"
)

// Subscriber defines the generic interface for subscribing to messages from messaging systems.
type Subscriber interface {
	// Subscribe starts receiving messages and sends them to the provided channel.
	Subscribe(ctx context.Context, msgChan chan RawEventBatch)

	// Close releases resources associated with the subscriber.
	Close(ctx context.Context) error
}
