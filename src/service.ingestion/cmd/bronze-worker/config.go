package main

import (
	"github.com/caarlos0/env/v11"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core/processor"
)

// Config holds the entire configuration for the application.
type Config struct {
	Processor processor.Config `envPrefix:"Processor__"`
	Container container.Config `envPrefix:"Container__"`
	EventHub  eventhub.Config  `envPrefix:"EventHub__"`
}

// LoadConfig loads configuration from environment variables and azure app configuration.
func LoadConfig() (*Config, error) {
	cfg := Config{}

	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
