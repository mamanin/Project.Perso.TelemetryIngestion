package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mailru/easyjson"
	"service.ingestion/external/messaging/eventhub"
	"service.ingestion/internal/core"
	"service.ingestion/pkg"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	connectionString := strings.TrimSpace(os.Getenv("ConnectionStrings__tispocevh001"))
	hubName := strings.TrimSpace(os.Getenv("TELEMETRY_RAW_EVENTHUBNAME"))

	sender, err := eventhub.NewPublisherForAspire(eventhub.AspireConfig{
		ConnectionString: connectionString,
		EventHubName:     hubName,
	})
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	for range 3 {
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
	for {
		batch := make([][]byte, 0, nb)

		for range nb {
			msg := generateTelemetries()
			data, err := easyjson.Marshal(msg)
			if err != nil {
				fmt.Printf("Marshal error: %v\n", err)
				continue
			}
			batch = append(batch, data)
		}

		fmt.Printf("Sending batch of %d messages\n", len(batch))
		err := sender.PublishBatch(ctx, batch)
		if err != nil {
			fmt.Printf("Failed to send batch: %s\n", err.Error())
		}

		time.Sleep(2 * time.Second)
	}
}

func generateTelemetries() pkg.TelemetryV2Event {
	event := pkg.TelemetryV2Event{
		VersionDiscriminant: pkg.VersionDiscriminant{Version: pkg.V2},
		DeviceId:            fmt.Sprintf("%d", rand.Intn(13000)+7000),
		Timestamp:           time.Now().Unix(),
	}

	type metricGen struct {
		origin, name, unit string
		value              func() any
	}

	gens := []metricGen{
		// Device
		{"device", "state", "", func() any { return []string{"on", "warning", "error"}[rand.Intn(3)] }},
		{"device", "uptime", "s", func() any { return rand.Float64() * 100000 }},
		{"device", "battery_level", "%", func() any { return rand.Float64() * 100 }},
		{"device", "firmware_version", "", func() any { return "1.2.3" }},

		// CPU
		{core.CpuSensor, "usage", "%", func() any { return rand.Float64() * 100 }},
		{core.CpuSensor, "usage_user", "%", func() any { return rand.Float64() * 100 }},
		{core.CpuSensor, "usage_system", "%", func() any { return rand.Float64() * 100 }},
		{core.CpuSensor, "usage_idle", "%", func() any { return rand.Float64() * 100 }},
		{core.CpuSensor, "temperature", "K", func() any { return 300 + rand.Float64()*50 }},
		{core.CpuSensor, "fan_speed", "RPM", func() any { return rand.Float64() * 5000 }},
		{core.CpuSensor, "memory", "%", func() any { return rand.Float64() * 100 }},
		{core.CpuSensor, "voltage", "V", func() any { return 0.5 + rand.Float64()*2.5 }},

		// Environment
		{core.EnvironmentSensor, "temperature", "K", func() any { return 280 + rand.Float64()*40 }},
		{core.EnvironmentSensor, "humidity", "%", func() any { return rand.Float64() * 100 }},
		{core.EnvironmentSensor, "pressure", "Pa", func() any { return 95000 + rand.Float64()*10000 }},
		{core.EnvironmentSensor, "lux", "lx", func() any { return rand.Float64() * 1000 }},

		// Network
		{core.NetworkSensor, "bytes_received", "B", func() any { return rand.Float64() * 1000000 }},
		{core.NetworkSensor, "bytes_sent", "B", func() any { return rand.Float64() * 1000000 }},
		{core.NetworkSensor, "latency", "s", func() any { return rand.Float64() }},
		{core.NetworkSensor, "packets_dropped", "", func() any { return rand.Float64() * 10 }},

		// Storage
		{core.StorageSensor, "usage_percent", "%", func() any { return rand.Float64() * 100 }},
		{core.StorageSensor, "health", "", func() any { return []string{"good", "warning", "critical"}[rand.Intn(3)] }},
		{core.StorageSensor, "total_size", "B", func() any { return 1000000000000.0 }},

		// Battery
		{core.BatterySensor, "level", "%", func() any { return rand.Float64() * 100 }},
		{core.BatterySensor, "status", "", func() any {
			return []string{"charging", "discharging", "full", "not_charging", "unknown"}[rand.Intn(5)]
		}},
		{core.BatterySensor, "voltage", "V", func() any { return rand.Float64() * 12 }},
	}

	count := min(rand.Intn(7)+6, len(gens))
	indices := rand.Perm(len(gens))

	for i := range count {
		g := gens[indices[i]]

		numMeasures := rand.Intn(5) + 1
		measures := make([]struct {
			Timestamp int64 `json:"timestamp"`
			Value     any   `json:"value"`
		}, 0, numMeasures)

		for j := range numMeasures {
			measures = append(measures, struct {
				Timestamp int64 `json:"timestamp"`
				Value     any   `json:"value"`
			}{
				Timestamp: event.Timestamp - int64(numMeasures-j),
				Value:     g.value(),
			})
		}

		event.Metrics = append(event.Metrics, struct {
			Name     string `json:"name"`
			Origin   string `json:"origin"`
			Unit     string `json:"unit,omitempty"`
			Measures []struct {
				Timestamp int64 `json:"timestamp"`
				Value     any   `json:"value"`
			}
		}{
			Name:     g.name,
			Origin:   g.origin,
			Unit:     g.unit,
			Measures: measures,
		})
	}

	return event
}
