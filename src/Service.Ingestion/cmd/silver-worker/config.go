package main

import (
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core/processor"
)

const (
	Key = "Value"
)

// Config holds the entire configuration for the application.
type Config struct {
	producer   processor.Config
	checkpoint container.Config
	eventHub   struct {
		metrics eventhub.SubscriberConfig
		data    eventhub.Config
	}
	// TODO: redis
}

// LoadConfig loads configuration from environment variables and azure app configuration.
func LoadConfig() (*Config, error) {
	cfg := &Config{}
	return cfg, nil
}
