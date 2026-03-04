package adx

import "time"

// DevicesStats represents the state of a device.
type DevicesStats struct {
	Total        int64 `kusto:"total_device"`
	Online       int64 `kusto:"on_device"`
	Offline      int64 `kusto:"off_device"`
	Issues       int64 `kusto:"issue_device"`
	TotalSensors int64 `kusto:"total_sensors"`
}

// DeviceState represents the state of a device.
type DeviceState struct {
	DeviceId  string    `kusto:"device_id"`
	Status    string    `kusto:"status"`
	Sensors   []string  `kusto:"sensors"`
	Heartbeat time.Time `kusto:"last_update"`
}
