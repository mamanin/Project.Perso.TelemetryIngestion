package probes

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"service.ingestion/internal/core/observability/logger"
)

// ProbeStatus represents the current status of a probe.
type ProbeStatus int

const (
	_                   ProbeStatus = iota
	ProbeStatusNotReady             // ProbeStatusNotReady indicates the probe is not ready to serve traffic.
	ProbeStatusReady                // ProbeStatusReady indicates the probe is ready and healthy.
	ProbeStatusFailed               // ProbeStatusFailed indicates the probe has failed.
)

// String returns a string representation of the probe status.
func (s ProbeStatus) String() string {
	switch s {
	case ProbeStatusNotReady:
		return "NOT_READY"
	case ProbeStatusReady:
		return "READY"
	case ProbeStatusFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// HealthProbe defines the common interface for all health probes.
type HealthProbe interface {
	Name() string
	Status() ProbeStatus
	Start(ctx context.Context) error
	Stop() error
	SetReady()
	SetFailed()
}

// Checker defines the interface for checking application health.
type Checker interface {
	IsHealthy(ctx context.Context) bool
}

// Probe represents a generic health check probe.
type Probe struct {
	HealthProbe

	name     string
	port     int
	status   ProbeStatus
	stopChan chan struct{}
	mu       sync.RWMutex
	listener net.Listener
	logger   logger.Logger
}

// NewProbe creates a new probe instance.
func NewProbe(name string, port int, logger logger.Logger) *Probe {
	return &Probe{
		name:     name,
		port:     port,
		status:   ProbeStatusNotReady,
		stopChan: make(chan struct{}, 1),
		logger:   logger,
	}
}

// Name returns the probe name.
func (p *Probe) Name() string {
	return p.name
}

// Status returns the current probe status thread-safely.
func (p *Probe) Status() ProbeStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return p.status
}

// SetStatus updates the probe status thread-safely.
func (p *Probe) setStatus(s ProbeStatus) {
	p.mu.Lock()
	defer p.mu.Unlock()

	oldStatus := p.status

	if oldStatus == s {
		return
	}

	p.status = s
	p.logger.Info("Probe '%s' status changed from '%s' to '%s'", p.name, oldStatus.String(), s.String())
}

// Start begins listening on the probe's TCP port and handles incoming connections.
func (p *Probe) Start(ctx context.Context) error {
	addr := fmt.Sprintf(":%d", p.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start %s probe on port %d: %w", p.name, p.port, err)
	}

	p.listener = listener

	p.logger.Info("Started probe: '%s' on port : '%d'", p.name, p.port)
	go p.acceptConnections(ctx)

	return nil
}

// Stop gracefully stops the probe.
func (p *Probe) Stop() error {
	close(p.stopChan)
	return p.listener.Close()
}

// acceptConnections handles incoming probe connections.
func (p *Probe) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		default:
			if p.Status() != ProbeStatusReady {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if tcpListener, ok := p.listener.(*net.TCPListener); ok {
				err := tcpListener.SetDeadline(time.Now().Add(500 * time.Millisecond))
				if err != nil {
					p.logger.Warn("Failed to set deadline on tcp listener")
				}
			}

			conn, err := p.listener.Accept()
			if err != nil {
				var netErr net.Error
				if errors.As(err, &netErr) && netErr.Timeout() {
					continue // No incoming connection, continue
				}

				select {
				case <-ctx.Done():
					return
				case <-p.stopChan:
					return
				default:
					p.logger.Error(err, "'%s' probe failed to accept connection: %v", p.name, err)
					continue
				}
			}

			go p.handleConnection(conn)
		}
	}
}

// handleConnection processes a single probe connection.
func (p *Probe) handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		if err := conn.Close(); err != nil {
			p.logger.Error(err, "'%s' probe failed to close connection: %v", p.name, err)
		}
	}(conn)

	// Simply accept and close the connection to indicate success
}
