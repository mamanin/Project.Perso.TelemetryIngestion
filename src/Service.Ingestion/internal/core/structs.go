package core

import (
	"strconv"
	"strings"
)

// MetricEvent represents a single metric data point.
type MetricEvent struct {
	SensorPath string `json:"sp"`
	Name       string `json:"n"`
	Value      any    `json:"v"`
	Unit       string `json:"u"`
	Timestamp  int64  `json:"_ts"`
}

// GetMetricRule returns the MetricRuleProvider based on the sensor path.
func (e *MetricEvent) GetMetricRule() *MetricRuleProvider {
	for _, entry := range metricRuleMap {
		if entry.pattern.MatchString(e.SensorPath) {
			return entry.rules
		}
	}
	return nil
}

// GetMetricKey generates a unique key for the MetricEvent by combining the sensor path and metric name.
func (e *MetricEvent) GetMetricKey() string {
	return strings.ReplaceAll(e.SensorPath[1:], "/", ":") + ":" + e.Name + ":" + strconv.FormatInt(roundTo2Minutes(e.Timestamp), 10)
}

// roundTo2Minutes rounds the given timestamp (in seconds) up to the nearest 2-minute interval.
func roundTo2Minutes(timestamp int64) int64 {
	const fiveMinutes = int64(120) // 2 minutes = 120 seconds
	remainder := timestamp % fiveMinutes

	if remainder == 0 {
		return timestamp
	}
	return timestamp + (fiveMinutes - remainder)
}

// MetricRuleProvider contains a list of MetricRules associated with a specific source.
type MetricRuleProvider struct {
	Source string
	Rules  MetricProvider
}

// MetricProvider maps metric names to their corresponding MetricRules.
type MetricProvider = map[string]MetricRules

// MetricRules represents a rule for processing MetricEvent.
type MetricRules struct {
	converter MetricConverter
	validator MetricValidator
}

// MetricConverter defines a function type for converting the value of a MetricConverter depending on the unit.
type MetricConverter = func(event *MetricEvent) bool

// MetricValidator defines a function type for validating the value of a MetricEvent.
type MetricValidator = func(event *MetricEvent) bool

// Validate applies the validator function to the given MetricEvent.
func (mr MetricRules) Validate(event *MetricEvent) bool {
	mr.converter(event)
	return mr.validator(event)
}
