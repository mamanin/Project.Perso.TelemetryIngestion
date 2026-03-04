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
	nb := 50
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

func generateTelemetries() pkg.TelemetryLegacyEvent {
	return pkg.TelemetryLegacyEvent{
		DeviceId:       fmt.Sprintf("%d", rand.Intn(9000)+1000),
		Timestamp:      time.Now().Unix(),
		CpuMemory:      rand.Float64() * 100,
		CpuUsage:       rand.Float64() * 100,
		Uptime:         rand.Float64() * 3600,
		State:          []string{"on", "warning", "error"}[rand.Intn(3)],
		CpuTemperature: 20 + rand.Float64()*100,
		NetworkLatency: rand.Float64(),
	}
}
