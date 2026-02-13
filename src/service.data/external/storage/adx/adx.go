package adx

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/kql"
	"service.data/internal/core/observability/logger"
)

const (
	listDevicesStateQuery = `
		devices_state
		| where device_id contains match
		| sort by last_update desc, device_id
		| project device_id, status, sensors, last_update
	`
)

// Client handles interactions with Azure Data Explorer.
type Client struct {
	database string
	client   *kusto.Client
	logger   logger.Logger
}

// Config holds configuration for the Azure Data Explorer client.
type Config struct {
	// TODO
}

// AspireConfig holds configuration for the Aspire Azure Data Explorer client.
type AspireConfig struct {
	Endpoint string
	Database string
}

// NewClient creates a new adx.Client based on the provided configuration.
func NewClient(_ Config, _ logger.Logger) (*Client, error) {
	// kusto_conn_string.WithDefaultAzureCredential()
	// ingestor, err = ingest.New(client, cfg.Database, cfg.Table)
	panic("not implemented")
}

// NewClientForAspire creates a new adx.Client for Aspire based on the provided configuration.
func NewClientForAspire(cfg AspireConfig, logger logger.Logger) (*Client, error) {
	client, err := kusto.New(kusto.NewConnectionStringBuilder(cfg.Endpoint))
	if err != nil {
		return nil, fmt.Errorf("failed to create kusto client: %w", err)
	}

	return &Client{
		database: cfg.Database,
		client:   client,
		logger:   logger,
	}, nil
}

// DeviceState represents the state of a device.
type DeviceState struct {
	// DeviceId is the unique identifier for the device.
	DeviceId string `kusto:"device_id"`
	// Status is the current status of the device.
	Status string `kusto:"status"`
	// Sensors is a set of unique sensor names associated with the device.
	Sensors []string `kusto:"sensors"`
	// Heartbeat is the most recent timestamp of data received for the device.
	Heartbeat time.Time `kusto:"last_update"`
}

// ListDevices retrieves a list of Device, limited by the specified number of records to take.
func (c *Client) ListDevices(ctx context.Context, take int64, match string) ([]DeviceState, error) {
	devices := make([]DeviceState, 0, take)

	iter, err := c.client.Query(ctx, c.database,
		kql.New(listDevicesStateQuery),
		kusto.QueryParameters(kql.NewParameters().AddString("match", match)),
		kusto.QueryTakeMaxRecords(take),
		kusto.QueryDataScope(kusto.DSHotCache),
	)
	if err != nil {
		return devices, fmt.Errorf("failed to execute query: %w", err)
	}
	defer iter.Stop()

	var row *table.Row
	var iErr *errors.Error
	var d DeviceState
	for {
		row, iErr, err = iter.NextRowOrError()
		if iErr != nil {
			c.logger.Error(iErr, "Query returned an error for a specific row: %v", iErr)
			continue
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return devices, fmt.Errorf("failed to read query results: %w", err)
		}

		if err = row.ToStruct(&d); err != nil {
			c.logger.Error(err, "Failed to map query result to struct: %v", err)
			continue
		}

		devices = append(devices, d)
	}

	return devices, nil
}

// Close closes the underlying Kusto client.
func (c *Client) Close() error {
	return c.client.Close()
}
