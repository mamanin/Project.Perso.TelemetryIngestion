package main

import (
	"os"
	"strconv"

	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/adx"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core/processor"
)

const (
	BlobContainerConnectionStringKey = "PARTITION_CHECKPOINTS_CONNECTIONSTRING"
	BlobContainerNameKey             = "PARTITION_CHECKPOINTS_BLOBCONTAINERNAME"

	EventHubConnectionStringKey    = "ConnectionStrings__tispocevh001"
	EventHubSubscriberNameKey      = "TELEMETRY_DATA_EVENTHUBNAME"
	EventHubSubscriberBatchSizeKey = "TELEMETRY_DATA_BATCHSIZE"

	ProcessorWorkerCountKey = "PROCESSOR_WORKERCOUNT"

	AdxConnectionStringKey = "TELEMETRIES_URI"
	AdxDatabaseKey         = "TELEMETRIES_DATABASENAME"
)

// AspireConfig holds the entire configuration for the application.
type AspireConfig struct {
	processor          processor.Config
	container          container.AspireConfig
	eventhubSubscriber eventhub.AspireSubscriberConfig
	adx                adx.AspireConfig
}

// LoadAspireConfig loads configuration from environment variables provided in Aspire.
func LoadAspireConfig() (*AspireConfig, error) {
	cfg := &AspireConfig{}

	workers, _ := strconv.Atoi(os.Getenv(ProcessorWorkerCountKey))
	cfg.processor = processor.Config{
		Workers: workers,
	}

	cfg.container = container.AspireConfig{
		ConnectionString: os.Getenv(BlobContainerConnectionStringKey),
		ContainerName:    os.Getenv(BlobContainerNameKey),
	}

	batchSize, _ := strconv.Atoi(os.Getenv(EventHubSubscriberBatchSizeKey))
	cfg.eventhubSubscriber = eventhub.AspireSubscriberConfig{
		AspireConfig: eventhub.AspireConfig{
			ConnectionString: os.Getenv(EventHubConnectionStringKey),
			EventHubName:     os.Getenv(EventHubSubscriberNameKey),
		},
		BatchSize:    batchSize,
		PrefetchSize: int32(batchSize * (workers + workers/2)),
	}

	cfg.adx = adx.AspireConfig{
		ConnectionString: os.Getenv(AdxConnectionStringKey),
		Database:         os.Getenv(AdxDatabaseKey),
		Table:            "metrics",
	}

	return cfg, nil
}
