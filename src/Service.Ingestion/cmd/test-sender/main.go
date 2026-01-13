package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/internal/core"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connectionString := strings.TrimSpace(os.Getenv("ConnectionStrings__pocitpevh001"))
	hubName := strings.TrimSpace(os.Getenv("TELEMETRY_METRICS_EVENTHUBNAME"))

	sender, err := eventhub.NewPublisher(eventhub.Config{
		ConnectionString: connectionString,
		EventHubName:     hubName,
	})
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	for range 10 {
		wg.Go(func() {
			if err = sendBatch(ctx, sender); err != nil {
				fmt.Printf("Error in sendBatch: %s\n", err.Error())
			}
		})
	}
	wg.Wait()
}

func sendBatch(ctx context.Context, sender *eventhub.Publisher) error {
	nb := 200
	enableFailCase := false
	for {
		batch := make([][]byte, 0, nb)

		for range nb {
			//goland:noinspection ALL
			msg := generateRandomMetric(enableFailCase)
			fmt.Printf("MSG: %s %s %v %s\n", msg.SensorPath, msg.Name, msg.Value, msg.Unit)
			data, err := json.Marshal(msg)
			if err != nil {
				fmt.Printf("Marshal error: %v\n", err)
				continue
			}
			batch = append(batch, data)
		}

		tCtx, tCancel := context.WithTimeout(ctx, 30*time.Second)
		fmt.Printf("Sending batch of %d messages\n", len(batch))
		err := sender.PublishBatch(tCtx, batch)
		if err != nil {
			fmt.Printf("Failed to send batch: %s\n", err.Error())
		}
		tCancel()
	}
}

type metricSimDef struct {
	unit       string
	gen        func() any
	genInvalid func() any
}

type sensorSimDef struct {
	suffix  string
	metrics map[string]metricSimDef
}

// Helpers for value generation
func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func randInt(min, max int) int {
	return rand.IntN(max-min+1) + min
}

func pick[T any](options []T) T {
	return options[rand.IntN(len(options))]
}

// Data dictionary mirroring constants.go logic
var simulationRules = map[string]sensorSimDef{
	"device": {
		suffix: "",
		metrics: map[string]metricSimDef{
			"state": {
				unit:       "",
				gen:        func() any { return pick(core.DeviceStatus) },
				genInvalid: func() any { return "exploded" },
			},
			"uptime": {
				unit:       "s",
				gen:        func() any { return randFloat(0, 100000) },
				genInvalid: func() any { return -100.0 },
			},
			"battery_level": {
				unit:       "%",
				gen:        func() any { return randFloat(0, 100) },
				genInvalid: func() any { return 150.0 },
			},
			"firmware_version": {
				unit:       "",
				gen:        func() any { return fmt.Sprintf("v%d.%d.%d", randInt(1, 5), randInt(0, 9), randInt(0, 9)) },
				genInvalid: func() any { return 101 }, // Wrong type (int instead of string)
			},
		},
	},
	"cpu": {
		suffix: "/cpu",
		metrics: map[string]metricSimDef{
			"usage": {
				unit:       "%",
				gen:        func() any { return randFloat(0, 100) },
				genInvalid: func() any { return 150.0 },
			},
			"usage_user": {
				unit:       "%",
				gen:        func() any { return randFloat(0, 50) },
				genInvalid: func() any { return -10.0 },
			},
			// Adding a few representative metrics
			"temperature": {
				unit:       "K",
				gen:        func() any { return randFloat(290, 350) },
				genInvalid: func() any { return 5000.0 },
			},
			"fan_speed": {
				unit:       "RPM",
				gen:        func() any { return float64(randInt(1000, 5000)) },
				genInvalid: func() any { return -500.0 },
			},
		},
	},
	"environment": {
		suffix: "/environment",
		metrics: map[string]metricSimDef{
			"temperature": {
				unit:       "K",
				gen:        func() any { return randFloat(270, 310) },
				genInvalid: func() any { return 100.0 }, // Below 200
			},
			"humidity": {
				unit:       "%",
				gen:        func() any { return randFloat(20, 80) },
				genInvalid: func() any { return 110.0 },
			},
			"pressure": {
				unit:       "Pa",
				gen:        func() any { return randFloat(90000, 110000) },
				genInvalid: func() any { return 10.0 },
			},
		},
	},
	"network": {
		suffix: "/network",
		metrics: map[string]metricSimDef{
			"bytes_received": {
				unit:       "B",
				gen:        func() any { return float64(randInt(1000, 1000000)) },
				genInvalid: func() any { return -1.0 },
			},
			"latency": {
				unit:       "s",
				gen:        func() any { return randFloat(0.001, 0.5) },
				genInvalid: func() any { return -0.1 },
			},
		},
	},
	"storage": {
		suffix: "/storage",
		metrics: map[string]metricSimDef{
			"usage_percent": {
				unit:       "%",
				gen:        func() any { return randFloat(10, 90) },
				genInvalid: func() any { return 105.0 },
			},
			"health": {
				unit:       "",
				gen:        func() any { return pick(core.StorageHealths) },
				genInvalid: func() any { return "dying" },
			},
		},
	},
	"battery": {
		suffix: "/battery",
		metrics: map[string]metricSimDef{
			"level": {
				unit:       "%",
				gen:        func() any { return randFloat(0, 100) },
				genInvalid: func() any { return -1.0 },
			},
			"status": {
				unit:       "",
				gen:        func() any { return pick(core.BatteryStatus) },
				genInvalid: func() any { return "leaking" },
			},
		},
	},
}

func generateRandomMetric(enableFailCase bool) core.MetricEvent {
	// 5% chance of generating a failure case
	var isFailCase bool
	if !enableFailCase {
		isFailCase = false
	} else {
		isFailCase = rand.Float64() < 0.05
	}

	// Pick a random sensor type
	sensorKeys := make([]string, 0, len(simulationRules))
	for k := range simulationRules {
		sensorKeys = append(sensorKeys, k)
	}
	sensorType := pick(sensorKeys)
	sensorDef := simulationRules[sensorType]

	// Generate base path: /device/12345/suffix
	deviceID := randInt(10000, 99999)
	sensorPath := fmt.Sprintf("/device/%d%s", deviceID, sensorDef.suffix)

	// Pick a random metric for this sensor
	metricKeys := make([]string, 0, len(sensorDef.metrics))
	for k := range sensorDef.metrics {
		metricKeys = append(metricKeys, k)
	}
	metricName := pick(metricKeys)
	metricDef := sensorDef.metrics[metricName]

	evt := core.MetricEvent{
		SensorPath: sensorPath,
		Name:       metricName,
		Unit:       metricDef.unit,
		Timestamp:  time.Now().Unix(),
	}

	if !isFailCase {
		// Happy Path
		evt.Value = metricDef.gen()
	} else {
		// Failure Path - pick one of 4 failure modes
		failMode := rand.IntN(4)
		switch failMode {
		case 0: // Unknown Sensor Path
			evt.SensorPath = fmt.Sprintf("/unknown/%d", deviceID)
			evt.Value = metricDef.gen()
		case 1: // Wrong Property Name
			evt.Name = "invalid_metric_name"
			evt.Value = metricDef.gen()
		case 2: // Wrong Value Type (Simulated by genInvalid if it changes type, or just swapping type)
			// Try to handle generically: if float, send string. If string, send int.
			val := metricDef.gen()
			switch v := val.(type) {
			case float64:
				evt.Value = fmt.Sprintf("value_is_%f", v)
			case string:
				evt.Value = 12345
			default:
				evt.Value = "invalid_type"
			}
		case 3: // Invalid Value (Out of range / Bad enum)
			evt.Value = metricDef.genInvalid()
		}
	}

	return evt
}
