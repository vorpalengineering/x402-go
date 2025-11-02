package main

import (
	"flag"
	"log"

	"github.com/vorpalengineering/x402-go/facilitator"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()

	// Load config
	cfg, err := facilitator.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create and start facilitator
	f := facilitator.NewFacilitator(cfg)
	defer f.Close()

	if err := f.Run(); err != nil {
		log.Fatalf("Failed to run facilitator: %v", err)
	}
}
