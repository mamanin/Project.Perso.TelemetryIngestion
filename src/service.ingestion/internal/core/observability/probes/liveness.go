package probes

import (
	"context"
	"runtime/metrics"
	"time"

	"service.ingestion/internal/core/observability/logger"
)

const (
	LivenessProbeName = "Liveness"
)

// LivenessProbe represents a probe that checks if the application has started successfully.
type LivenessProbe struct {
	*Probe

	metrics       map[string]uint64
	checkInterval time.Duration
}

// NewLivenessProbe creates a new startup probe.
func NewLivenessProbe(logger logger.Logger) HealthProbe {
	return &LivenessProbe{
		Probe: NewProbe(LivenessProbeName, 8081, logger),
		metrics: map[string]uint64{
			"/memory/classes/heap/objects:bytes": 1, // TODO: configure with app settings
			// TODO: add additional metrics
		},
		checkInterval: 10 * time.Second, // Initial check interval
	}
}

// Start begins the startup probe.
func (lp *LivenessProbe) Start(ctx context.Context) error {
	if err := lp.Probe.Start(ctx); err != nil {
		return err
	}

	go lp.healthCheckRoutine(ctx)
	return nil
}

// SetReady marks the startup probe as ready.
func (lp *LivenessProbe) SetReady() {
	lp.setStatus(ProbeStatusReady)
}

// SetFailed marks the startup probe as failed.
func (lp *LivenessProbe) SetFailed() {
	lp.setStatus(ProbeStatusFailed)
}

// healthCheckRoutine continuously checks the health of the application.
func (lp *LivenessProbe) healthCheckRoutine(ctx context.Context) {
	ticker := time.NewTicker(lp.checkInterval)
	defer ticker.Stop()

	s := make([]metrics.Sample, 0, len(lp.metrics))
	for name := range lp.metrics {
		s = append(s, metrics.Sample{
			Name: name,
		})
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-lp.stopChan:
			return
		case <-ticker.C:
			lp.performHealthCheck(s)
		}
	}
}

func (lp *LivenessProbe) performHealthCheck(s []metrics.Sample) {
	metrics.Read(s)

	for _, sample := range s {
		if sample.Value.Uint64() > lp.metrics[sample.Name] {
			// TODO: activate when thresholds are configured
			//lp.logger.Warn("Health check failed: metric %s value %d exceeds threshold %d", sample.Name, sample.Value.Uint64(), lp.metrics[sample.Name])
			//lp.SetFailed()
			return
		}
	}

	lp.SetReady()
}
