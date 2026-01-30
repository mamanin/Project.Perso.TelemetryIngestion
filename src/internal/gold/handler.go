package gold

import (
	"context"
	"time"

	"service.ingestion/external/storage/adx"
	"service.ingestion/internal/core"
	"service.ingestion/internal/core/observability/logger"
	"service.ingestion/internal/core/processor"
)

// Handler processes batches of core.MetricEvent items.
type Handler struct {
	logger logger.Logger
	adx    *adx.Client[core.DataMetric]
}

// NewHandler creates a new Handler instance.
func NewHandler(logger logger.Logger, adx *adx.Client[core.DataMetric]) processor.Handler[core.MetricEvent] {
	return &Handler{
		logger: logger,
		adx:    adx,
	}
}

// Handle processes a batch of core.MetricEvent items.
func (h *Handler) Handle(ctx context.Context, batch []*core.MetricEvent) {
	metrics := make([]core.DataMetric, 0, len(batch))

	for _, event := range batch {
		id, s, ok := event.Origin()
		if !ok {
			h.logger.Warn("Invalid sensor path: %s", event.SensorPath)
			continue
		}

		metrics = append(metrics, core.DataMetric{
			Timestamp: time.Unix(event.Timestamp, 0).UTC(),
			DeviceId:  id,
			Sensor:    s,
			Metric:    event.Name,
			Value:     event.Value,
			Unit:      event.Unit,
		})
	}

	if len(metrics) == 0 {
		return
	}

	if err := h.adx.IngestBatch(ctx, metrics); err != nil {
		h.logger.Error(err, "Failed to ingest metrics to ADX: %v", err)
	}
}
