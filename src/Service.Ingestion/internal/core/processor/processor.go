package processor

import "context"

// Processor defines the interface for a generic processor with start and stop capabilities.
type Processor interface {
	// Start initiates the processor's operations.
	Start(ctx context.Context)

	// Stop terminates the processor's operations.
	Stop(ctx context.Context) error
}
