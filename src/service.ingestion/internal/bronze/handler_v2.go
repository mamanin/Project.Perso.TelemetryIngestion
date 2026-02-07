package bronze

import (
	"context"

	"github.com/mailru/easyjson"
	"service.ingestion/external/messaging"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/processor"
	"service.ingestion/pkg"
)

// V2Handler processes batches of pkg.TelemetryV2Event items.
type V2Handler struct {
	logger    logger.Logger
	publisher messaging.Publisher
}

// NewV2Handler creates a new V2Handler instance.
func NewV2Handler(logger logger.Logger, publisher messaging.Publisher) processor.Handler[pkg.TelemetryV2Event] {
	return &V2Handler{
		logger:    logger,
		publisher: publisher,
	}
}

// Handle processes a batch of pkg.TelemetryV2Event items.
func (h *V2Handler) Handle(ctx context.Context, batch []*pkg.TelemetryV2Event) {
	var events [][]byte
	var me core.MetricEvent

	for _, event := range batch {
		for _, metric := range event.Metrics {
			for _, measure := range metric.Measures {
				me = core.MetricEvent{
					SensorPath: core.FormatSensorPath(event.DeviceId, metric.Origin),
					Name:       metric.Name,
					Value:      measure.Value,
					Unit:       metric.Unit,
					Timestamp:  measure.Timestamp,
				}

				if bytes, err := easyjson.Marshal(me); err == nil {
					events = append(events, bytes)
				}
			}
		}
	}

	if err := h.publisher.PublishBatch(ctx, events); err != nil {
		h.logger.Error(err, "Error publishing metric updates: %v", err)
	}
}
