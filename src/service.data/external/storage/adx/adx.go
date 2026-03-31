package adx

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-kusto-go/kusto"
	"github.com/Azure/azure-kusto-go/kusto/data/errors"
	"github.com/Azure/azure-kusto-go/kusto/data/table"
	"github.com/Azure/azure-kusto-go/kusto/kql"
	"service.data/internal/core/observability/logger"
)

const (
	devicesStatsQuery = `
		devices_state
		| extend is_off = last_update < ago(1h)
		| summarize
				total_device = count(),
				on_device = countif(not(is_off) and status == "on"),
				off_device = countif(is_off),
				issue_device = countif(not(is_off) and (status == "warning" or status == "error")),
				total_sensors = sum(sensors_count)
	`
	listDevicesStateQuery = `
		devices_state
		| where device_id contains match
		| extend is_off = last_update < ago(1h)
		| where filterMode == ""
				or (filterMode == "on" and not(is_off) and status == "on")
				or (filterMode == "off" and is_off)
				or (filterMode == "issue" and not(is_off) and (status == "warning" or status == "error"))
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

// GetDevicesStats retrieves statistics about devices.
func (c *Client) GetDevicesStats(ctx context.Context) (DevicesStats, error) {
	var stats DevicesStats

	iter, err := c.client.Query(ctx, c.database,
		kql.New(devicesStatsQuery),
	)
	if err != nil {
		return stats, fmt.Errorf("failed to execute query: %w", err)
	}
	defer iter.Stop()

	row, iErr, err := iter.NextRowOrError()
	if iErr != nil || err != nil {
		return stats, fmt.Errorf("query returned an error for a specific row: %w", iErr)
	}

	if err = row.ToStruct(&stats); err != nil {
		return stats, fmt.Errorf("failed to map query result to struct: %w", err)
	}

	return stats, nil
}

// ListDevices retrieves a list of Device, limited by the specified number of records to take.
func (c *Client) ListDevices(ctx context.Context, take int64, match string, status string) ([]DeviceState, error) {
	devices := make([]DeviceState, 0, take)

	iter, err := c.client.Query(ctx, c.database,
		kql.New(listDevicesStateQuery),
		kusto.QueryParameters(kql.NewParameters().AddString("match", match).AddString("filterMode", status)),
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
