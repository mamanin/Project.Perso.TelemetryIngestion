package main

import (
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage"
	"service.ingestion/internal/silver"
)

const (
	Key = "Value"
)

// Config holds the entire configuration for the application.
type Config struct {
	producer   silver.Config
	checkpoint storage.Config
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
