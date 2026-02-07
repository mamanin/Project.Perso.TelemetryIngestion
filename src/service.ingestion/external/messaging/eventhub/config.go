package eventhub

// Config holds configuration for the Event Hub.
type Config struct {
	// TODO
}

// SubscriberConfig holds configuration for the Event Hub subscriber.
type SubscriberConfig struct {
	Config

	// BatchSize defines the number of messages to receive in each batch.
	BatchSize int
	// PrefetchSize defines the number of messages to prefetch.
	PrefetchSize int32
}

// AspireConfig holds configuration for connecting to an Aspire Event Hub instance.
type AspireConfig struct {
	ConnectionString string
	EventHubName     string
}

// AspireSubscriberConfig holds configuration for the Aspire Event Hub subscriber.
type AspireSubscriberConfig struct {
	AspireConfig

	// BatchSize defines the number of messages to receive in each batch.
	BatchSize int
	// PrefetchSize defines the number of messages to prefetch.
	PrefetchSize int32
}
