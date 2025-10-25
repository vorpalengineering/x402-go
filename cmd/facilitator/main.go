package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
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

	// Create Gin router
	router := gin.Default()

	// Register routes
	facilitator.RegisterRoutes(router, cfg)

	// Start facilitator
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting x402 Facilitator service on %s", addr)
	log.Printf("Supported Schemes: %v", cfg.Supported)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
