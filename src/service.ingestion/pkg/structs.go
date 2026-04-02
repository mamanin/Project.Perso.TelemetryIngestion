package pkg

//>easyjson .\pkg\structs.go

const (
	V2     = "v2.0"
	V1     = "v1.0"
	Legacy = "legacy"
)

// VersionDiscriminant represent contains the discriminator for telemetries
//
//easyjson:json
type VersionDiscriminant struct {
	Version string `json:"version"`
}

// TelemetryV2Event represents the version 2.0 telemetry.
//
//easyjson:json
type TelemetryV2Event struct {
	VersionDiscriminant

	DeviceId  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"`
	Metrics   []struct {
		Name     string `json:"name"`
		Origin   string `json:"origin"`
		Unit     string `json:"unit,omitempty"`
		Measures []struct {
			Timestamp int64 `json:"timestamp"`
			Value     any   `json:"value"`
		} `json:"measures"`
	} `json:"metrics"`
}

// TelemetryV1Event represents the version 1.0 telemetry.
//
//easyjson:json
type TelemetryV1Event struct {
	VersionDiscriminant

	DeviceId  string `json:"device_id"`
	Timestamp int64  `json:"timestamp"`
	Data      []struct {
		Name   string `json:"name"`
		Origin string `json:"origin"`
		Unit   string `json:"unit,omitempty"`
		Value  any    `json:"value"`
	} `json:"data"`
}

// TelemetryLegacyEvent represents the legacy versions telemetry.
//
//easyjson:json
type TelemetryLegacyEvent struct {
	DeviceId       string  `json:"device_id"`
	Timestamp      int64   `json:"timestamp"`
	State          string  `json:"state"`
	Uptime         float64 `json:"uptime"`
	CpuUsage       float64 `json:"cpu_usage"`
	CpuMemory      float64 `json:"cpu_memory"`
	CpuTemperature float64 `json:"cpu_temperature"`
	NetworkLatency float64 `json:"network_latency"`
}
