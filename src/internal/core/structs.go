package core

//>easyjson .\internal\core\structs.go

import (
	"strconv"
	"strings"
	"time"
)

// MetricEvent represents a single metric data point.
//
//easyjson:json
type MetricEvent struct {
	SensorPath string `json:"sp"`
	Name       string `json:"n"`
	Value      any    `json:"v"`
	Unit       string `json:"u,omitempty"`
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
	return "service.ingestion:metric.update:" + strings.ReplaceAll(e.SensorPath[1:], "/", ":") + ":" + e.Name + ":" + strconv.FormatInt(roundTo5Minutes(e.Timestamp), 10)
}

// roundTo5Minutes rounds the given timestamp (in seconds) up to the nearest 5-minute interval.
func roundTo5Minutes(timestamp int64) int64 {
	const fiveMinutes = int64(300) // 5 minutes = 300 seconds
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
	if !mr.converter(event) {
		return false
	}
	return mr.validator(event)
}

// DataMetric represents the structure of a metric data point to be ingested into storage.
//
//easyjson:json
type DataMetric struct {
	Timestamp time.Time `json:"timestamp"`
	DeviceId  string    `json:"device_id"`
	Metric    string    `json:"metric"`
	Value     any       `json:"value"`
	Unit      string    `json:"unit,omitempty"`
}
