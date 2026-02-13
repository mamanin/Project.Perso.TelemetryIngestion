package main

import (
	"os"

	"service.data/external/storage/adx"
)

const (
	AdxEndpointKey = "TELEMETRIES_URI"
	AdxDatabaseKey = "TELEMETRIES_DATABASENAME"
)

// AspireConfig holds the entire configuration for the application.
type AspireConfig struct {
	adx adx.AspireConfig
}

// LoadAspireConfig loads configuration from environment variables provided in Aspire.
func LoadAspireConfig() (*AspireConfig, error) {
	cfg := &AspireConfig{}

	cfg.adx = adx.AspireConfig{
		Endpoint: os.Getenv(AdxEndpointKey),
		Database: os.Getenv(AdxDatabaseKey),
	}

	return cfg, nil
}
