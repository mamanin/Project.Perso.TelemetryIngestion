package eventhub

// Config holds configuration for the Event Hub.
type Config struct {
	ConnectionString string
	EventHubName     string
}

// SubscriberConfig holds configuration for the Event Hub subscriber.
type SubscriberConfig struct {
	Config

	// BatchSize defines the number of messages to receive in each batch.
	BatchSize int
	// PrefetchSize defines the number of messages to prefetch.
	PrefetchSize int32
}
