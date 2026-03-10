package probes

import (
	"context"
	"fmt"

	"service.ingestion/internal/core/observability/logger"
)

// Manager manages all health probes in a centralized way.
type Manager struct {
	logger logger.Logger
	probes map[string]HealthProbe
}

// NewManager initializes all health probes.
func NewManager(c []Checker, logger logger.Logger) (*Manager, error) {
	logger.Info("Initializing health checks...")

	s := NewStartupProbe(logger)
	l := NewLivenessProbe(logger)
	r := NewReadinessProbe(c, logger)

	return &Manager{
		probes: map[string]HealthProbe{
			s.Name(): s,
			l.Name(): l,
			r.Name(): r,
		},
		logger: logger,
	}, nil
}

// StartProbes starts all health probes.
func (hm *Manager) StartProbes(ctx context.Context) error {
	var errors []error
	for _, p := range hm.probes {
		hm.logger.Info("Starting health probe %s", p.Name())
		if err := p.Start(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to start probe %s: %w", p.Name(), err))
		}
	}

	if len(errors) > 0 {
		err := hm.stopAllProbes()
		if err != nil {
			return fmt.Errorf("failed to start and stop %d probe(s): %v", len(errors), errors)
		}
		return fmt.Errorf("failed to start %d probe(s): %v", len(errors), errors)
	}

	hm.logger.Info("All health probes started successfully")
	return nil
}

// StopProbes gracefully stops all health probes.
func (hm *Manager) StopProbes() error {
	return hm.stopAllProbes()
}

// stopAllProbes is an internal method to stop all probes (caller must hold lock).
func (hm *Manager) stopAllProbes() error {
	var errors []error
	for _, p := range hm.probes {
		if err := p.Stop(); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop probe %s: %w", p.Name(), err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to stop %d probe(s): %v", len(errors), errors)
	}

	hm.logger.Info("All health probes stopped successfully")
	return nil
}

// MarkProbeAs marks a specific probe as ready or failed based on the provided status.
func (hm *Manager) MarkProbeAs(name string, s ProbeStatus) error {
	probe, ok := hm.probes[name]
	if !ok {
		return fmt.Errorf("probe %s not found", name)
	}

	switch s {
	case ProbeStatusNotReady:
		return nil
	case ProbeStatusReady:
		probe.SetReady()
	case ProbeStatusFailed:
		probe.SetFailed()
	}
	return nil
}
