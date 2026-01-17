package bronze

import (
	"context"

	"github.com/mailru/easyjson"
	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/processor"
	"service.ingestion/pkg"
)

// V1Handler processes batches of pkg.TelemetryV1Event items.
type V1Handler struct {
	logger    logger.Logger
	publisher messaging.Publisher
}

// NewV1Handler creates a new V1Handler instance.
func NewV1Handler(logger logger.Logger, publisher messaging.Publisher) processor.Handler[pkg.TelemetryV1Event] {
	return &V1Handler{
		logger:    logger,
		publisher: publisher,
	}
}

// Handle processes a batch of pkg.TelemetryV1Event items.
func (h *V1Handler) Handle(ctx context.Context, batch []*pkg.TelemetryV1Event) {
	var events [][]byte
	var me core.MetricEvent

	for _, event := range batch {
		for _, data := range event.Data {
			me = core.MetricEvent{
				SensorPath: core.FormatSensorPath(event.DeviceId, data.Origin),
				Name:       data.Name,
				Value:      data.Value,
				Unit:       data.Unit,
				Timestamp:  event.Timestamp,
			}

			if bytes, err := easyjson.Marshal(me); err == nil {
				events = append(events, bytes)
			}
		}
	}

	if err := h.publisher.PublishBatch(ctx, events); err != nil {
		h.logger.Error(err, "Error publishing metric updates: %v", err)
	}
}
