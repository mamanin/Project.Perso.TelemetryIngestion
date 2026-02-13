package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var app AppManager
	var err error
	switch os.Getenv("deployment-environment") {
	case "Aspire":
		app, err = NewAspireApp(ctx)
	default:
		app, err = NewApp(ctx)
	}
	if err != nil {
		fmt.Printf("Failed to initialize bronze worker: %s", err.Error())
		os.Exit(1)
	}

	if err = app.Start(); err != nil {
		fmt.Printf("Failed to start bronze worker: %s", err.Error())
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	app.Stop()
}
