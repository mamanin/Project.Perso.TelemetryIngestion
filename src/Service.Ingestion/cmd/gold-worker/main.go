package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// main is the entry point for the gold worker.
func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := NewApp(ctx)
	if err != nil {
		fmt.Printf("Failed to initialize gold worker: %s", err.Error())
		os.Exit(1)
	}

	if err = app.Start(ctx); err != nil {
		fmt.Printf("Failed to start gold worker: %s", err.Error())
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	app.Stop(ctx)
}
