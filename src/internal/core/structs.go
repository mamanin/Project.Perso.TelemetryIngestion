package core

//>easyjson .\internal\core\structs.go

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DataMetric represents the structure of a metric data point to be ingested into storage.
//
//easyjson:json
type DataMetric struct {
	Timestamp time.Time `json:"timestamp"`
	DeviceId  string    `json:"device_id"`
	Sensor    string    `json:"sensor"`
	Metric    string    `json:"metric"`
	Value     any       `json:"value"`
	Unit      string    `json:"unit,omitempty"`
}

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

// Origin extracts the device id and sensor name from the SensorPath.
//
// See core.DevicePattern and core.SensorPattern for deconstruction of the SensorPath.
func (e *MetricEvent) Origin() (string, string, bool) {
	sps := strings.Split(e.SensorPath[1:], "/")
	switch len(sps) {
	case 2:
		return sps[1], sps[0], true // core.DevicePattern metric with no sensor
	case 3:
		return sps[1], fmt.Sprintf("%s.%s", sps[0], sps[2]), true // core.SensorPattern metric
	default:
		return "", "", false
	}
}

// Rule returns the MetricRuleProvider based on the sensor path.
func (e *MetricEvent) Rule() *MetricRuleProvider {
	for _, entry := range metricRuleMap {
		if entry.pattern.MatchString(e.SensorPath) {
			return entry.rules
		}
	}
	return nil
}

// CacheKey generates a unique cache key for the MetricEvent.
//
// See core.DevicePattern and core.SensorPattern for deconstruction of the SensorPath.
func (e *MetricEvent) CacheKey() string {
	return "service.ingestion:silver:metric.update" +
		":" + strings.ReplaceAll(e.SensorPath[1:], "/", ":") +
		":" + e.Name +
		":" + strconv.FormatInt(roundTo2Minutes(e.Timestamp), 10)
}

// roundTo2Minutes rounds the given timestamp (in seconds) up to the nearest 2-minute interval.
func roundTo2Minutes(timestamp int64) int64 {
	const twoMinutes = int64(120) // 2 minutes = 120 seconds
	remainder := timestamp % twoMinutes

	if remainder == 0 {
		return timestamp
	}
	return timestamp + (twoMinutes - remainder)
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

// ConvertAndValidate applies the validator function to the given MetricEvent.
func (mr MetricRules) ConvertAndValidate(event *MetricEvent) bool {
	if !mr.converter(event) {
		return false
	}
	return mr.validator(event)
}
