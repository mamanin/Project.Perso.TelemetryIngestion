package silver

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"service.ingestion/external/cache"
	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/processor"
)

// Handler processes batches of core.MetricEvent items.
type Handler struct {
	logger    logger.Logger
	redis     *cache.Redis
	publisher messaging.Publisher
}

// NewHandler creates a new Handler instance.
func NewHandler(logger logger.Logger, redis *cache.Redis, publisher messaging.Publisher) processor.Handler[core.MetricEvent] {
	return &Handler{
		logger:    logger,
		redis:     redis,
		publisher: publisher,
	}
}

// Handle processes a batch of core.MetricEvent items.
func (h *Handler) Handle(ctx context.Context, batch []*core.MetricEvent) {
	metricProcesses := make([]*cache.ItemProcess[core.MetricEvent], 0, len(batch))
	rp := h.redis.Pipeline()

	for _, event := range batch {
		rules := event.GetMetricRule()
		if rules == nil {
			h.logger.Warn("No metric rules found for sensor path: %s", event.SensorPath)
			continue
		}

		mr, ok := rules.Rules[event.Name]
		if !ok {
			h.logger.Warn("No metric rules found for metric name: %s", event.Name)
			continue
		}

		if !mr.Validate(event) {
			continue
		}

		key := event.GetMetricKey()
		m := &cache.ItemProcess[core.MetricEvent]{
			Key:    key,
			Item:   event,
			GetCmd: rp.Get(ctx, key),
		}
		metricProcesses = append(metricProcesses, m)
	}

	if _, err := rp.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		h.logger.Error(err, "Error executing hit redis pipeline: %v", err)
		return
	}

	metricUpdates := make([][]byte, 0, len(metricProcesses))
	rp = h.redis.Pipeline()

	for _, mp := range metricProcesses {
		_, err := mp.GetCmd.Result()
		if err == nil {
			continue
		}

		bytes, err := json.Marshal(mp.Item)
		if err != nil {
			continue
		}

		metricUpdates = append(metricUpdates, bytes)
		rp.Set(ctx, mp.Key, mp.Item.Value, 1*time.Hour)
	}

	if _, err := rp.Exec(ctx); err != nil {
		h.logger.Error(err, "Error executing set redis pipeline: %v", err)
		return
	}

	if len(metricUpdates) == 0 {
		return
	}

	tCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.publisher.PublishBatch(tCtx, metricUpdates); err != nil {
		h.logger.Error(err, "Error publishing metric updates: %v", err)
	}
}
