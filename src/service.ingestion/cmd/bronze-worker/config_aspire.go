package main

import (
	"os"
	"strconv"

	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core/processor"
)

const (
	BlobContainerConnectionStringKey = "PARTITION_CHECKPOINTS_CONNECTIONSTRING"
	BlobContainerNameKey             = "PARTITION_CHECKPOINTS_BLOBCONTAINERNAME"

	EventHubConnectionStringKey    = "ConnectionStrings__tispocevh001"
	EventHubSubscriberNameKey      = "TELEMETRY_RAW_EVENTHUBNAME"
	EventHubPublisherNameKey       = "TELEMETRY_METRICS_EVENTHUBNAME"
	EventHubSubscriberBatchSizeKey = "TELEMETRY_RAW_BATCHSIZE"

	ProcessorWorkerCountKey = "PROCESSOR_WORKERCOUNT"
)

// AspireConfig holds the entire configuration for the application.
type AspireConfig struct {
	processor          processor.Config
	container          container.AspireConfig
	eventhubSubscriber eventhub.AspireSubscriberConfig
	eventhubPublisher  eventhub.AspireConfig
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

	cfg.eventhubPublisher = eventhub.AspireConfig{
		ConnectionString: os.Getenv(EventHubConnectionStringKey),
		EventHubName:     os.Getenv(EventHubPublisherNameKey),
	}

	return cfg, nil
}
