package eventhub

// Config holds configuration for the Event Hub.
type Config struct {
	FullyQualifiedNamespace string           `env:"FullyQualifiedNamespace,required"`
	Subscriber              SubscriberConfig `envPrefix:"Subscriber__"`
	Publisher               PublisherConfig  `envPrefix:"Publisher__"`
}

// AspireConfig holds configuration for connecting to an Aspire Event Hub instance.
type AspireConfig struct {
	ConnectionString string
	EventHubName     string
}

// AspireSubscriberConfig holds configuration for the Aspire Event Hub subscriber.
type AspireSubscriberConfig struct {
	AspireConfig

	BatchSize    int
	PrefetchSize int32
}
