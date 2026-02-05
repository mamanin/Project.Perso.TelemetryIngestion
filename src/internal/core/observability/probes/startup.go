package probes

import (
	"context"

	"service.ingestion/internal/core/observability/logger"
)

const (
	StartupProbeName = "Startup"
)

// StartupProbe represents a probe that checks if the application has started successfully.
type StartupProbe struct {
	*Probe
}

// NewStartupProbe creates a new startup probe.
func NewStartupProbe(logger logger.Logger) HealthProbe {
	return &StartupProbe{
		Probe: NewProbe(StartupProbeName, 8080, logger),
	}
}

// Start begins the startup probe.
func (sp *StartupProbe) Start(ctx context.Context) error {
	return sp.Probe.Start(ctx)
}

// SetReady marks the startup probe as ready.
func (sp *StartupProbe) SetReady() {
	sp.setStatus(ProbeStatusReady)
}

// SetFailed marks the startup probe as failed.
func (sp *StartupProbe) SetFailed() {
	sp.setStatus(ProbeStatusFailed)
}
