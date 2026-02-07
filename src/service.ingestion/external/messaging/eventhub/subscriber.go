package eventhub

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/messaging/azeventhubs/v2"
	"service.ingestion/external/messaging"
	"service.ingestion/external/storage/container"
	"service.ingestion/internal/core/observability/logger"
)

// Subscriber implements the messaging.Subscriber interface for Azure Event Hub.
type Subscriber struct {
	options SubscriberOption

	mu        sync.Mutex
	partition string
	latest    *azeventhubs.ReceivedEventData

	logger    logger.Logger
	client    *azeventhubs.ConsumerClient
	processor *azeventhubs.Processor

	stopChan chan struct{}
	doneChan chan struct{}
}

// SubscriberOption holds configuration options for the Subscriber.
type SubscriberOption struct {
	// batchSize defines the number of messages to receive in each batch.
	batchSize int
	// prefetchSize defines the number of messages to prefetch.
	prefetchSize int32
}

// NewSubscriber creates a new Event Hub subscriber.
func NewSubscriber(_ logger.Logger, _ SubscriberConfig, _ *container.Checkpoint) (*Subscriber, error) {
	panic("not implemented")
}

// NewSubscriberForAspire creates a new Event Hub subscriber for Aspire configuration.
func NewSubscriberForAspire(logger logger.Logger, cfg AspireSubscriberConfig, cp *container.Checkpoint) (*Subscriber, error) {
	client, err := azeventhubs.NewConsumerClientFromConnectionString(cfg.ConnectionString, cfg.EventHubName, azeventhubs.DefaultConsumerGroup, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer client: %w", err)
	}

	processor, err := azeventhubs.NewProcessor(client, cp, &azeventhubs.ProcessorOptions{Prefetch: cfg.PrefetchSize})
	if err != nil {
		return nil, fmt.Errorf("failed to create processor: %w", err)
	}

	return &Subscriber{
		logger:    logger,
		client:    client,
		processor: processor,

		options: SubscriberOption{
			batchSize:    cfg.BatchSize,
			prefetchSize: cfg.PrefetchSize,
		},

		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}, nil
}

// Subscribe starts receiving messages an event hub and sends them to the provided channel.
func (s *Subscriber) Subscribe(ctx context.Context, queue chan messaging.RawEventBatch) {
	var wg sync.WaitGroup

	wg.Go(func() {
		if err := s.processor.Run(ctx); err != nil {
			s.logger.Error(err, "Processor run failed: %v", err)
		}
	})

	wg.Go(func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-s.stopChan:
				return
			default:
				pc := s.processor.NextPartitionClient(ctx)
				if pc == nil {
					s.logger.Debug("No partition client found for subscription")
					time.Sleep(1 * time.Second)
					continue
				}

				s.mu.Lock()
				s.partition = pc.PartitionID()
				s.mu.Unlock()

				s.processPartition(ctx, pc, queue)
			}
		}
	})

	go func() {
		wg.Wait()
		close(s.doneChan)
	}()
}

// processPartition processes events from a specific partition client.
func (s *Subscriber) processPartition(ctx context.Context, pc *azeventhubs.ProcessorPartitionClient, queue chan messaging.RawEventBatch) {
	defer func(ppc *azeventhubs.ProcessorPartitionClient) {
		s.updateCheckpoint(ctx, pc)

		if err := ppc.Close(ctx); err != nil {
			s.logger.Error(err, "Failed to close processor partition client: %v", err)
		}
	}(pc)

	s.receiveEvents(ctx, pc, queue)
}

// receiveEvents receives events from the given partition client and sends them to the queue.
func (s *Subscriber) receiveEvents(ctx context.Context, pc *azeventhubs.ProcessorPartitionClient, queue chan messaging.RawEventBatch) {
	ct := time.NewTicker(3 * time.Minute)
	defer ct.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			s.updateCheckpoint(ctx, pc)
			return
		case <-ct.C:
			go s.updateCheckpoint(ctx, pc)
		default:
			start := time.Now()

			reCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
			events, err := pc.ReceiveEvents(reCtx, s.options.batchSize, nil)
			cancel()

			if err != nil {
				s.logger.Debug("Unable to receive events for partition %s: %v", pc.PartitionID(), err)
				return
			}

			if len(events) == 0 {
				continue
			}

			eventPool := make(messaging.RawEventBatch, len(events))
			for i, event := range events {
				eventPool[i] = event.Body
			}

			select {
			case <-ctx.Done():
				return
			case <-s.stopChan:
				return
			case queue <- eventPool:
				s.mu.Lock()
				s.latest = events[len(events)-1]
				s.mu.Unlock()
			}

			s.logger.Info("R|%.2fμs|%de", float64(time.Since(start).Microseconds()), len(events))
		}
	}
}

// updateCheckpoint updates the checkpoint for the latest processed event.
func (s *Subscriber) updateCheckpoint(ctx context.Context, pc *azeventhubs.ProcessorPartitionClient) {
	s.mu.Lock()
	le := s.latest
	s.mu.Unlock()

	if le == nil {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := pc.UpdateCheckpoint(ctx, le, nil); err != nil {
		s.logger.Error(err, "Failed to final checkpoint for partition %s: %v", pc.PartitionID(), err)
	}
}

// IsHealthy checks if the subscriber is healthy by verifying connectivity to the Event Hub partition.
func (s *Subscriber) IsHealthy(ctx context.Context) bool {
	s.mu.Lock()
	pId := s.partition
	s.mu.Unlock()

	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if _, err := s.client.GetPartitionProperties(ctx, pId, nil); err != nil {
		return false
	}

	return true
}

// Close releases resources associated with the subscriber.
func (s *Subscriber) Close(ctx context.Context) error {
	defer func(dCtx context.Context, c *azeventhubs.ConsumerClient) {
		if err := c.Close(dCtx); err != nil {
			s.logger.Error(err, "Failed to close event hub consumer client: %v", err)
		}
	}(ctx, s.client)

	close(s.stopChan)

	select {
	case <-s.doneChan:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout waiting for receiver to stop")
	}
}
