package adx

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/ingest"
	"github.com/mailru/easyjson"
	"service.ingestion/external/credential"
	"service.ingestion/internal/core/observability/logger"
)

// Client handles ingestion into Azure Data Explorer.
type Client[T easyjson.Marshaler] struct {
	client   *kusto.Client
	ingestor ingest.Ingestor
	logger   logger.Logger
}

// Config holds configuration for the Azure Data Explorer client.
type Config struct {
	Endpoint string `env:"Endpoint,required"`
	Database string `env:"Database,required"`
	Table    string `env:"Table,required"`
}

// AspireConfig holds configuration for the Aspire Azure Data Explorer client.
type AspireConfig struct {
	ConnectionString string
	Database         string
	Table            string
}

// NewClient creates a new adx.Client based on the provided configuration.
func NewClient[T easyjson.Marshaler](cfg Config, logger logger.Logger, cred credential.AzureCredentials) (*Client[T], error) {
	client, err := kusto.New(kusto.NewConnectionStringBuilder(cfg.Endpoint).WithTokenCredential(cred))
	if err != nil {
		return nil, fmt.Errorf("failed to create kusto client: %w", err)
	}

	ingestor, err := ingest.New(client, cfg.Database, cfg.Table)
	if err != nil {
		if err = client.Close(); err != nil {
			return nil, fmt.Errorf("failed to close kusto client after ingestor creation failure: %w", err)
		}
		return nil, fmt.Errorf("failed to create ingestor: %w", err)
	}

	return &Client[T]{
		client:   client,
		ingestor: ingestor,
		logger:   logger,
	}, nil
}

// NewClientForAspire creates a new adx.Client for Aspire based on the provided configuration.
func NewClientForAspire[T easyjson.Marshaler](cfg AspireConfig, logger logger.Logger) (*Client[T], error) {
	client, err := kusto.New(kusto.NewConnectionStringBuilder(cfg.ConnectionString))
	if err != nil {
		return nil, fmt.Errorf("failed to create kusto client: %w", err)
	}

	ingestor, err := ingest.NewStreaming(client, cfg.Database, cfg.Table)
	if err != nil {
		if err = client.Close(); err != nil {
			return nil, fmt.Errorf("failed to close kusto client after ingestor creation failure: %w", err)
		}
		return nil, fmt.Errorf("failed to create ingestor: %w", err)
	}

	return &Client[T]{
		client:   client,
		ingestor: ingestor,
		logger:   logger,
	}, nil
}

// IngestBatch ingests a batch of data into Azure Data Explorer.
func (c *Client[T]) IngestBatch(ctx context.Context, data []T) error {
	r, w := io.Pipe()

	go func(d []T) {
		defer func(w *io.PipeWriter) {
			if err := w.Close(); err != nil {
				c.logger.Error(err, "failed to close pipe writer: %v", err)
			}
		}(w)

		for _, item := range d {
			if _, err := easyjson.MarshalToWriter(item, w); err != nil {
				c.logger.Error(err, "failed to marshal item to JSON: %v", err)
				continue
			}
			if _, err := w.Write([]byte("\n")); err != nil {
				c.logger.Error(err, "failed to write newline: %v", err)
				return
			}
		}
	}(data)

	if _, err := c.ingestor.FromReader(ctx, r, ingest.FileFormat(ingest.JSON)); err != nil {
		return fmt.Errorf("failed to ingest from reader: %w", err)
	}
	return nil
}

// Close closes the underlying Kusto client.
func (c *Client[T]) Close() error {
	return c.client.Close()
}
