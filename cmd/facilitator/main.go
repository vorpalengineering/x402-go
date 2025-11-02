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

	// Set Gin mode based on log level
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Register routes
	facilitator.RegisterRoutes(router, cfg)

	// Initialize RPC connections
	log.Println("Initializing RPC connections...")
	if err := facilitator.InitializeRPCClients(); err != nil {
		log.Fatalf("Failed to initialize RPC clients: %v", err)
	}
	log.Println("RPC connections established")

	// Cleanup on exit
	defer facilitator.CloseAllRPCClients()

	// Start facilitator
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting x402 Facilitator service on %s", addr)
	log.Printf("Supported Schemes: %v", cfg.Supported)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
