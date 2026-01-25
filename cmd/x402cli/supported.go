package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	facilitatorclient "github.com/vorpalengineering/x402-go/facilitator/client"
)

func supportedCommand() {
	// Define flags for supported command
	supportedFlags := flag.NewFlagSet("supported", flag.ExitOnError)
	var facilitatorURL string
	supportedFlags.StringVar(&facilitatorURL, "facilitator", "", "URL of the facilitator service (required)")
	supportedFlags.StringVar(&facilitatorURL, "f", "", "URL of the facilitator service (required)")

	// Parse flags
	supportedFlags.Parse(os.Args[2:])

	// Validate required flags
	if facilitatorURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --facilitator or -f flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli supported --facilitator <url>")
		fmt.Fprintln(os.Stderr, "  x402cli supported -f <url>")
		supportedFlags.PrintDefaults()
		os.Exit(1)
	}

	// Create facilitator client and call /supported
	fc := facilitatorclient.NewFacilitatorClient(facilitatorURL)
	resp, err := fc.Supported()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Pretty-print the response as JSON
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}
