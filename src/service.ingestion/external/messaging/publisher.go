package messaging

import "context"

// Publisher defines the interface for publishing messages to messaging systems.
type Publisher interface {
	// PublishBatch sends multiple messages to the messaging system.
	PublishBatch(ctx context.Context, messages [][]byte) error

	// Close releases resources associated with the publisher.
	Close(ctx context.Context) error
}
