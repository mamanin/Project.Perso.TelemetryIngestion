package dtos

import (
	"time"

	"service.data/external/storage/adx"
)

// DeviceStateDto dto representing a device.
type DeviceStateDto struct {
	Id        string
	Status    string
	Sensors   []string
	Heartbeat string
}

// NewDeviceStateFromAdxDeviceState creates a new DeviceStateDto instance from an adx.DeviceState.
func NewDeviceStateFromAdxDeviceState(d adx.DeviceState) DeviceStateDto {
	return DeviceStateDto{
		Id:        d.DeviceId,
		Status:    d.Status,
		Sensors:   d.Sensors,
		Heartbeat: d.Heartbeat.Format(time.RFC3339),
	}
}
