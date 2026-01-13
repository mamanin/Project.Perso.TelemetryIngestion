package silver

import (
	"context"
	"errors"
	"time"

	"github.com/goccy/go-json"
	"github.com/redis/go-redis/v9"
	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
	"service.ingestion/internal/core"
)

// Handler processes batches of core.MetricEvent items.
type Handler struct {
	logger    logger.Logger
	redis     *redis.Client
	publisher messaging.Publisher
}

// NewHandler creates a new Handler instance.
func NewHandler(logger logger.Logger, redis *redis.Client, publisher messaging.Publisher) *Handler {
	return &Handler{
		logger:    logger,
		redis:     redis,
		publisher: publisher,
	}
}

// metricProcess holds information for processing a single metric event.
type metricProcess struct {
	key    string
	event  *core.MetricEvent
	getCmd *redis.StringCmd
}

// Handle processes a batch of core.MetricEvent items.
func (h *Handler) Handle(ctx context.Context, batch []*core.MetricEvent) {
	metricProcesses := make([]*metricProcess, 0, len(batch))
	hitPipe := h.redis.Pipeline()

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
		m := &metricProcess{
			key:    key,
			event:  event,
			getCmd: hitPipe.Get(ctx, key),
		}
		metricProcesses = append(metricProcesses, m)
	}

	if _, err := hitPipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		h.logger.Error(err, "Error executing hit redis pipeline: %v", err)
		return
	}

	metricUpdates := make([][]byte, 0, len(metricProcesses))
	setPipe := h.redis.Pipeline()

	for _, mp := range metricProcesses {
		_, err := mp.getCmd.Result()
		if err == nil {
			continue
		}

		bytes, err := json.Marshal(mp.event)
		if err != nil {
			continue
		}

		metricUpdates = append(metricUpdates, bytes)
		setPipe.Set(ctx, mp.key, mp.event.Value, 1*time.Hour)
	}

	if _, err := setPipe.Exec(ctx); err != nil {
		h.logger.Error(err, "Error executing set redis pipeline: %v", err)
		return
	}

	if len(metricUpdates) == 0 {
		return
	}

	if err := h.publisher.PublishBatch(ctx, metricUpdates); err != nil {
		h.logger.Error(err, "Error publishing metric updates: %v", err)
	}
}
