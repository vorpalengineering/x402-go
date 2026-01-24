package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/vorpalengineering/x402-go/facilitator"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "facilitator/config.yaml", "Path to config file")
	flag.Parse()

	// Load config
	cfg, err := facilitator.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create context that listens for shutdown signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v", sig)
		cancel()
	}()

	// Create and start facilitator
	f := facilitator.NewFacilitator(cfg)
	defer f.Close()

	if err := f.Run(ctx); err != nil {
		log.Fatalf("Failed to run facilitator: %v", err)
	}
}
