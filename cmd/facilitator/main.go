package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/facilitator"
)

func main() {
	// Create Gin router
	router := gin.Default()

	// Register routes
	facilitator.RegisterRoutes(router)

	// Start facilitator
	log.Println("Starting x402 facilitator service on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
