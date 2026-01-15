package bronze

import (
	"context"
	"time"

	"github.com/goccy/go-json"
	"service.ingestion/external/messaging"
	"service.ingestion/external/observability/logger"
	"service.ingestion/internal/core"
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

	for _, event := range batch {
		me := core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, ""),
			Name:       "state",
			Value:      event.State,
			Timestamp:  event.Timestamp,
		}
		if bytes, err := json.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, ""),
			Name:       "uptime",
			Value:      event.Uptime,
			Timestamp:  event.Timestamp,
		}
		if bytes, err := json.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.CpuSensor),
			Name:       "usage",
			Value:      event.CpuUsage,
			Timestamp:  event.Timestamp,
		}
		if bytes, err := json.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.CpuSensor),
			Name:       "memory",
			Value:      event.CpuMemory,
			Timestamp:  event.Timestamp,
		}
		if bytes, err := json.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.CpuSensor),
			Name:       "temperature",
			Value:      event.CpuTemperature,
			Timestamp:  event.Timestamp,
		}
		if bytes, err := json.Marshal(me); err == nil {
			events = append(events, bytes)
		}

		me = core.MetricEvent{
			SensorPath: core.FormatSensorPath(event.DeviceId, core.NetworkSensor),
			Name:       "latency",
			Value:      event.NetworkLatency,
			Timestamp:  event.Timestamp,
		}
		if bytes, err := json.Marshal(me); err == nil {
			events = append(events, bytes)
		}
	}

	tCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.publisher.PublishBatch(tCtx, events); err != nil {
		h.logger.Error(err, "Error publishing metric updates: %v", err)
	}
}
