package eventhub

import (
	"context"
	"errors"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"service.ingestion/external/credential"
)

// PublisherConfig holds configuration for the Event Hub publisher.
type PublisherConfig struct {
	Name string `env:"Name,required"`
}

// Publisher implements the messaging.Publisher interface for Azure Event Hub.
type Publisher struct {
	client *azeventhubs.ProducerClient
}

// NewPublisher creates a new Event Hub publisher.
func NewPublisher(cfg Config, cred credential.AzureCredentials) (*Publisher, error) {
	client, err := azeventhubs.NewProducerClient(cfg.FullyQualifiedNamespace, cfg.Publisher.Name, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create event hub producer client: %w", err)
	}

	return &Publisher{
		client: client,
	}, nil
}

// NewPublisherForAspire creates a new Event Hub publisher for Aspire configuration.
func NewPublisherForAspire(cfg AspireConfig) (*Publisher, error) {
	client, err := azeventhubs.NewProducerClientFromConnectionString(cfg.ConnectionString, cfg.EventHubName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create event hub producer client: %w", err)
	}

	return &Publisher{
		client: client,
	}, nil
}

// PublishBatch sends multiple messages to the event hub.
func (p *Publisher) PublishBatch(ctx context.Context, messages [][]byte) error {
	if len(messages) == 0 {
		return nil
	}

	batch, err := p.client.NewEventDataBatch(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to create event batch: %w", err)
	}

	for i, msg := range messages {
		eventData := &azeventhubs.EventData{
			Body: msg,
		}

		if err = batch.AddEventData(eventData, nil); err == nil {
			continue
		}

		if !errors.Is(err, azeventhubs.ErrEventDataTooLarge) {
			return fmt.Errorf("failed to add event %d to batch: %w", i, err)
		}

		if err = p.client.SendEventDataBatch(ctx, batch, nil); err != nil {
			return fmt.Errorf("failed to send partial batch: %w", err)
		}

		batch, err = p.client.NewEventDataBatch(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to create new batch: %w", err)
		}
	}

	if batch.NumEvents() == 0 {
		return nil
	}

	err = p.client.SendEventDataBatch(ctx, batch, nil)
	if err != nil {
		return fmt.Errorf("failed to send final batch: %w", err)
	}

	return nil
}

// Close releases resources associated with the publisher.
func (p *Publisher) Close(ctx context.Context) error {
	if err := p.client.Close(ctx); err != nil {
		return fmt.Errorf("failed to close producer client: %w", err)
	}

	return nil
}
