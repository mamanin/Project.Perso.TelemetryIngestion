package probes

import (
	"context"
	"time"

	"service.ingestion/internal/core/observability/logger"
)

const (
	ReadinessProbeName = "Readiness"
)

// ReadinessProbe represents a probe that continuously checks if the application is ready to serve traffic.
type ReadinessProbe struct {
	*Probe

	checkers      []Checker
	checkInterval time.Duration
}

// NewReadinessProbe creates a new readiness probe.
func NewReadinessProbe(c []Checker, logger logger.Logger) HealthProbe {
	return &ReadinessProbe{
		Probe:         NewProbe(ReadinessProbeName, 8082, logger),
		checkers:      c,
		checkInterval: 5 * time.Second, // Initial check interval
	}
}

// Start begins the readiness probe and starts the health checking routine.
func (rp *ReadinessProbe) Start(ctx context.Context) error {
	if err := rp.Probe.Start(ctx); err != nil {
		return err
	}

	go rp.healthCheckRoutine(ctx)
	return nil
}

// SetReady marks the readiness probe as ready and sets the check interval to 1 minute.
func (rp *ReadinessProbe) SetReady() {
	rp.checkInterval = 30 * time.Second
	rp.setStatus(ProbeStatusReady)
}

// SetFailed marks the readiness probe as failed.
func (rp *ReadinessProbe) SetFailed() {
	rp.setStatus(ProbeStatusFailed)
}

// healthCheckRoutine continuously checks the health of the application.
func (rp *ReadinessProbe) healthCheckRoutine(ctx context.Context) {
	ticker := time.NewTicker(rp.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rp.stopChan:
			return
		case <-ticker.C:
			rp.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck performs a single health check.
func (rp *ReadinessProbe) performHealthCheck(ctx context.Context) {
	for _, checker := range rp.checkers {
		if !checker.IsHealthy(ctx) {
			rp.SetFailed()
			return
		}
	}

	rp.SetReady()
}
