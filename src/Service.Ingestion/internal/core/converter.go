package core

import (
	"github.com/martinlindhe/unit"
)

// DefaultConverter is a no-op converter that always returns true.
func DefaultConverter(_ *MetricEvent) bool {
	return true
}

// PercentConverter ensures that the unit is percentage (%).
func PercentConverter(event *MetricEvent) bool {
	_, ok := event.Value.(float64)
	if !ok {
		return false
	}

	if event.Unit != "%" {
		return false
	}

	return true
}

// KelvinConverter converts temperature values to Kelvin (K).
func KelvinConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var cv unit.Temperature
	switch event.Unit {
	case "K":
		return true
	case "°C":
		cv = unit.FromCelsius(value)
	case "°F":
		cv = unit.FromFahrenheit(value)
	default:
		return false
	}

	event.Unit = "K"
	event.Value = cv.Kelvin()
	return true
}

// PascalsConverter converts pressure values to Pascals (Pa).
func PascalsConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var p unit.Pressure
	switch event.Unit {
	case "Pa":
		return true
	case "hPa":
		p = unit.Pressure(value) * unit.Hectopascal
	case "kPa":
		p = unit.Pressure(value) * unit.Kilopascal
	case "bar":
		p = unit.Pressure(value) * unit.Bar
	default:
		return false
	}

	event.Unit = "Pa"
	event.Value = p.Pascals()
	return true
}

// MetersPerSecondConverter converts speed values to Meters Per Second (m/s).
func MetersPerSecondConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var s unit.Speed
	switch event.Unit {
	case "m/s":
		return true
	case "km/h":
		s = unit.Speed(value) * unit.KilometersPerHour
	case "mph":
		s = unit.Speed(value) * unit.MilesPerHour
	case "knots":
		// 1 knot = 0.514444 m/s
		s = unit.Speed(value*0.514444) * unit.MetersPerSecond
	default:
		return false
	}

	event.Unit = "m/s"
	event.Value = s.MetersPerSecond()
	return true
}

// BytesConverter converts data size values to Bytes (B).
func BytesConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var d unit.Datasize
	switch event.Unit {
	case "B":
		return true
	case "KB":
		d = unit.Datasize(value) * unit.Kilobyte
	case "MB":
		d = unit.Datasize(value) * unit.Megabyte
	case "GB":
		d = unit.Datasize(value) * unit.Gigabyte
	case "TB":
		d = unit.Datasize(value) * unit.Terabyte
	case "KiB":
		d = unit.Datasize(value) * unit.Kibibyte
	case "MiB":
		d = unit.Datasize(value) * unit.Mebibyte
	case "GiB":
		d = unit.Datasize(value) * unit.Gibibyte
	default:
		return false
	}

	event.Unit = "B"
	event.Value = d.Bytes()
	return true
}

// VoltsConverter converts voltage values to Volts (V).
func VoltsConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var v unit.Voltage
	switch event.Unit {
	case "V":
		return true
	case "mV":
		v = unit.Voltage(value) * unit.Millivolt
	case "kV":
		v = unit.Voltage(value) * unit.Kilovolt
	default:
		return false
	}

	event.Unit = "V"
	event.Value = v.Volts()
	return true
}

// MetersConverter converts length values to Meters (m).
func MetersConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var l unit.Length
	switch event.Unit {
	case "m":
		return true
	case "cm":
		l = unit.Length(value) * unit.Centimeter
	case "mm":
		l = unit.Length(value) * unit.Millimeter
	case "km":
		l = unit.Length(value) * unit.Kilometer
	case "ft":
		l = unit.Length(value) * unit.Foot
	case "mi":
		l = unit.Length(value) * unit.Mile
	default:
		return false
	}

	event.Unit = "m"
	event.Value = l.Meters()
	return true
}

// AmperesConverter converts electric current values to Amperes (A).
func AmperesConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var c unit.ElectricCurrent
	switch event.Unit {
	case "A":
		return true
	case "mA":
		c = unit.ElectricCurrent(value) * unit.Milliampere
	default:
		return false
	}

	event.Unit = "A"
	event.Value = c.Amperes()
	return true
}

// WattsConverter converts power values to Watts (W).
func WattsConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var p unit.Power
	switch event.Unit {
	case "W":
		return true
	case "kW":
		p = unit.Power(value) * unit.Kilowatt
	case "mW":
		p = unit.Power(value) * unit.Milliwatt
	default:
		return false
	}

	event.Unit = "W"
	event.Value = p.Watts()
	return true
}

// BytesPerSecondConverter converts data rate values to Bytes Per Second (B/s).
func BytesPerSecondConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var r unit.Datarate
	switch event.Unit {
	case "B/s":
		return true
	case "MB/s":
		r = unit.Datarate(value) * unit.MegabytePerSecond
	case "Mbps":
		// Megabits per second
		r = unit.Datarate(value) * unit.MegabitPerSecond
	case "Gbps":
		r = unit.Datarate(value) * unit.GigabitPerSecond
	case "KB/s":
		r = unit.Datarate(value) * unit.KilobytePerSecond
	default:
		return false
	}

	event.Unit = "B/s"
	event.Value = r.BytesPerSecond()
	return true
}

// SecondsConverter converts time values to Seconds (s).
func SecondsConverter(event *MetricEvent) bool {
	value, ok := event.Value.(float64)
	if !ok {
		return false
	}

	var t unit.Duration
	switch event.Unit {
	case "s":
		return true
	case "ms":
		t = unit.Duration(value) * unit.Millisecond
	case "us", "µs":
		t = unit.Duration(value) * unit.Microsecond
	case "ns":
		t = unit.Duration(value) * unit.Nanosecond
	case "min":
		t = unit.Duration(value) * unit.Minute
	case "h":
		t = unit.Duration(value) * unit.Hour
	case "d":
		t = unit.Duration(value) * unit.Day
	default:
		return false
	}

	event.Unit = "s"
	event.Value = t.Seconds()
	return true
}
