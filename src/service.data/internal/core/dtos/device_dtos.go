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

// NewFromAdxDevice creates a new DeviceStateDto instance from an adx.DeviceState.
func NewFromAdxDevice(d adx.DeviceState) DeviceStateDto {
	return DeviceStateDto{
		Id:        d.DeviceId,
		Status:    d.Status,
		Sensors:   d.Sensors,
		Heartbeat: d.Heartbeat.Format(time.RFC3339),
	}
}
