package core

import "slices"

// DefaultValidator is a no-op validator that always returns true.
func DefaultValidator(_ *MetricEvent) bool {
	return true
}

// RangeValidator checks if the MetricEvent value is within the specified min and max range (inclusive).
func RangeValidator(min, max float64) MetricValidator {
	return func(event *MetricEvent) bool {
		value, ok := event.Value.(float64)
		return ok && value >= min && value <= max
	}
}

// MinValidator checks if the MetricEvent value is greater than or equal to min.
func MinValidator(min float64) MetricValidator {
	return func(event *MetricEvent) bool {
		value, ok := event.Value.(float64)
		return ok && value >= min
	}
}

// EnumValidator creates a validator that checks if the value is one of the allowed strings.
func EnumValidator(allowed []string) MetricValidator {
	return func(event *MetricEvent) bool {
		value, ok := event.Value.(string)
		return ok && slices.Contains(allowed, value)
	}
}
