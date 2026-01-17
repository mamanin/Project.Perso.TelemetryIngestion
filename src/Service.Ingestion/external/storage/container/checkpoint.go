package container

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2/checkpoints"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
)

// Config holds the configuration for the Blob Storage checkpoint store.
type Config struct {
	ConnectionString string
	ContainerName    string
}

// Checkpoint represents a Blob Storage checkpoint store.
type Checkpoint = checkpoints.BlobStore

// NewCheckpoint creates a new Blob Storage checkpoint store.
func NewCheckpoint(cfg Config) (*Checkpoint, error) {
	client, err := container.NewClientFromConnectionString(cfg.ConnectionString, cfg.ContainerName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob container client: %w", err)
	}

	cs, err := checkpoints.NewBlobStore(client, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create checkpoint store: %w", err)
	}

	return cs, nil
}
