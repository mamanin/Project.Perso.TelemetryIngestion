package dtos

import (
	"fmt"
	"math"
	"time"

	"service.data/external/storage/adx"
)

// DeviceStateDto dto representing a device.
type DeviceStateDto struct {
	Id           string
	Status       string
	Sensors      []string
	Heartbeat    string
	HeartbeatISO string
}

// NewDeviceStateFromAdxDeviceState creates a new DeviceStateDto instance from an adx.DeviceState.
func NewDeviceStateFromAdxDeviceState(d adx.DeviceState) DeviceStateDto {
	return DeviceStateDto{
		Id:           d.DeviceId,
		Status:       d.Status,
		Sensors:      d.Sensors,
		Heartbeat:    formatRelativeTime(d.Heartbeat),
		HeartbeatISO: d.Heartbeat.UTC().Format("2006-01-02 15:04 UTC"),
	}
}

// formatRelativeTime formats a time.Time as a human-readable relative duration (e.g. "2m ago").
func formatRelativeTime(t time.Time) string {
	duration := time.Since(t)

	if duration < 0 {
		return "just now"
	}

	seconds := int(math.Floor(duration.Seconds()))
	minutes := int(math.Floor(duration.Minutes()))
	hours := int(math.Floor(duration.Hours()))
	days := hours / 24

	switch {
	case seconds < 60:
		return "just now"
	case minutes < 60:
		return fmt.Sprintf("%dm ago", minutes)
	case hours < 24:
		return fmt.Sprintf("%dh ago", hours)
	case days < 30:
		return fmt.Sprintf("%dd ago", days)
	default:
		months := days / 30
		return fmt.Sprintf("%dmo ago", months)
	}
}
