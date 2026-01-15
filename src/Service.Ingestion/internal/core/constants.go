package core

import (
	"fmt"
	"regexp"
)

const (
	CpuSensor         = "cpu"
	EnvironmentSensor = "environment"
	NetworkSensor     = "network"
	StorageSensor     = "storage"
	BatterySensor     = "battery"

	devicePath    = "/device/%s"
	DevicePattern = "^\\/device\\/[0-9]{4,8}$"

	sensorPath    = "/device/%s/%s"
	SensorPattern = "^\\/device\\/[0-9]{4,8}\\/%s$"
)

var metricRuleMap = []struct {
	pattern *regexp.Regexp
	rules   *MetricRuleProvider
}{
	{regexp.MustCompile(DevicePattern), &DeviceRules},
	{regexp.MustCompile(fmt.Sprintf(SensorPattern, CpuSensor)), &CpuRules},
	{regexp.MustCompile(fmt.Sprintf(SensorPattern, EnvironmentSensor)), &EnvironmentRules},
	{regexp.MustCompile(fmt.Sprintf(SensorPattern, NetworkSensor)), &NetworkRules},
	{regexp.MustCompile(fmt.Sprintf(SensorPattern, StorageSensor)), &StorageRules},
	{regexp.MustCompile(fmt.Sprintf(SensorPattern, BatterySensor)), &BatteryRules},
}

// FormatSensorPath constructs the sensor path based on device ID and sensor name.
func FormatSensorPath(deviceID, sensor string) string {
	if sensor == "" {
		return fmt.Sprintf(devicePath, deviceID)
	}

	return fmt.Sprintf(sensorPath, deviceID, sensor)
}

var (
	// Device variable

	DeviceStatus = []string{"on", "off", "booting", "maintenance", "error"}
)

// DeviceRules defines the metric rules for device sensors.
var DeviceRules = MetricRuleProvider{
	Source: "device",
	Rules: MetricProvider{
		"state": MetricRules{
			converter: DefaultConverter,
			validator: EnumValidator(DeviceStatus),
		},
		"uptime": MetricRules{
			converter: SecondsConverter,
			validator: MinValidator(0),
		},
		"battery_level": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"firmware_version": MetricRules{
			converter: DefaultConverter,
			validator: func(event *MetricEvent) bool {
				_, ok := event.Value.(string)
				return ok
			},
		},
	},
}

// CpuRules defines the metric rules for CPU sensors.
var CpuRules = MetricRuleProvider{
	Source: CpuSensor,
	Rules: MetricProvider{
		"usage": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"usage_user": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"usage_system": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"usage_idle": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"memory": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"temperature": MetricRules{
			converter: KelvinConverter,
			validator: RangeValidator(0, 400),
		},
		"voltage": MetricRules{
			converter: VoltsConverter,
			validator: RangeValidator(0.5, 3),
		},
		"fan_speed": MetricRules{
			converter: DefaultConverter,
			validator: RangeValidator(0, 10000),
		},
	},
}

// EnvironmentRules defines the metric rules for environmental sensors.
var EnvironmentRules = MetricRuleProvider{
	Source: EnvironmentSensor,
	Rules: MetricProvider{
		"temperature": MetricRules{
			converter: KelvinConverter,
			validator: RangeValidator(200, 350),
		},
		"humidity": MetricRules{
			converter: DefaultConverter,
			validator: RangeValidator(0, 100),
		},
		"pressure": MetricRules{
			converter: PascalsConverter,
			validator: RangeValidator(80000, 120000),
		},
		"wind_speed": MetricRules{
			converter: MetersPerSecondConverter,
			validator: RangeValidator(0, 150),
		},
		"wind_direction": MetricRules{
			converter: DefaultConverter,
			validator: RangeValidator(0, 360),
		},
		"rainfall": MetricRules{
			converter: MetersConverter,
			validator: MinValidator(0),
		},
		"lux": MetricRules{
			converter: DefaultConverter,
			validator: MinValidator(0),
		},
		"co2": MetricRules{
			converter: DefaultConverter,
			validator: RangeValidator(0, 5000),
		},
	},
}

// NetworkRules defines the metric rules for network sensors.
var NetworkRules = MetricRuleProvider{
	Source: NetworkSensor,
	Rules: MetricProvider{
		"bytes_received": MetricRules{
			converter: BytesConverter,
			validator: MinValidator(0),
		},
		"bytes_sent": MetricRules{
			converter: BytesConverter,
			validator: MinValidator(0),
		},
		"packets_dropped": MetricRules{
			converter: DefaultConverter,
			validator: MinValidator(0),
		},
		"latency": MetricRules{
			converter: SecondsConverter,
			validator: MinValidator(0),
		},
		"bandwidth_usage": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
	},
}

var (
	// Storage variable

	StorageHealths = []string{"good", "warning", "critical"}
)

// StorageRules defines the metric rules for storage sensors.
var StorageRules = MetricRuleProvider{
	Source: StorageSensor,
	Rules: MetricProvider{
		"total_size": MetricRules{
			converter: BytesConverter,
			validator: MinValidator(0),
		},
		"used_size": MetricRules{
			converter: BytesConverter,
			validator: MinValidator(0),
		},
		"free_size": MetricRules{
			converter: BytesConverter,
			validator: MinValidator(0),
		},
		"usage_percent": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"read_throughput": MetricRules{
			converter: BytesPerSecondConverter,
			validator: MinValidator(0),
		},
		"write_throughput": MetricRules{
			converter: BytesPerSecondConverter,
			validator: MinValidator(0),
		},
		"health": MetricRules{
			converter: DefaultConverter,
			validator: EnumValidator(StorageHealths),
		},
	},
}

var (
	// Battery variable

	BatteryStatus = []string{"charging", "discharging", "full", "not_charging", "unknown"}
)

// BatteryRules defines the metric rules for battery sensors.
var BatteryRules = MetricRuleProvider{
	Source: BatterySensor,
	Rules: MetricProvider{
		"level": MetricRules{
			converter: PercentConverter,
			validator: RangeValidator(0, 100),
		},
		"voltage": MetricRules{
			converter: VoltsConverter,
			validator: RangeValidator(0, 100),
		},
		"current": MetricRules{
			converter: AmperesConverter,
			validator: RangeValidator(-1000, 1000),
		},
		"power": MetricRules{
			converter: WattsConverter,
			validator: MinValidator(0),
		},
		"temperature": MetricRules{
			converter: KelvinConverter,
			validator: RangeValidator(200, 400),
		},
		"status": MetricRules{
			converter: DefaultConverter,
			validator: EnumValidator(BatteryStatus),
		},
		"cycle_count": MetricRules{
			converter: DefaultConverter,
			validator: MinValidator(0),
		},
		"time_to_empty": MetricRules{
			converter: SecondsConverter,
			validator: MinValidator(0),
		},
	},
}
