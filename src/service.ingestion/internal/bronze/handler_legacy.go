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

// LegacyHandler processes batches of pkg.TelemetryLegacyEvent items.
type LegacyHandler struct {
	logger    logger.Logger
	publisher messaging.Publisher
}

// NewLegacyHandler creates a new LegacyHandler instance.
func NewLegacyHandler(logger logger.Logger, publisher messaging.Publisher) processor.Handler[pkg.TelemetryLegacyEvent] {
	return &LegacyHandler{
		logger:    logger,
		publisher: publisher,
	}
}

// Handle processes a batch of pkg.TelemetryLegacyEvent items.
func (h *LegacyHandler) Handle(ctx context.Context, batch []*pkg.TelemetryLegacyEvent) {
	var events [][]byte
	var me core.MetricEvent

	for _, event := range batch {
		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.Device),
			Name:       "state",
			Value:      event.State,
			Timestamp:  event.Timestamp,
		}
		if bytes, err := easyjson.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.Device),
			Name:       "uptime",
			Value:      event.Uptime,
			Timestamp:  event.Timestamp,
			Unit:       "min",
		}
		if bytes, err := easyjson.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.CpuSensor),
			Name:       "usage",
			Value:      event.CpuUsage,
			Timestamp:  event.Timestamp,
			Unit:       "%",
		}
		if bytes, err := easyjson.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.CpuSensor),
			Name:       "memory",
			Value:      event.CpuMemory,
			Timestamp:  event.Timestamp,
			Unit:       "%",
		}
		if bytes, err := easyjson.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.CpuSensor),
			Name:       "temperature",
			Value:      event.CpuTemperature,
			Timestamp:  event.Timestamp,
			Unit:       "°C",
		}
		if bytes, err := easyjson.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.NetworkSensor),
			Name:       "latency",
			Value:      event.NetworkLatency,
			Timestamp:  event.Timestamp,
			Unit:       "µs",
		}
		if bytes, err := easyjson.Marshal(me); err == nil {
			events = append(events, bytes)
		}
	}

	if err := h.publisher.PublishBatch(ctx, events); err != nil {
		h.logger.Error(err, "Error publishing metric updates: %v", err)
	}
}
