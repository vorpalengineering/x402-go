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
	var url string
	supportedFlags.StringVar(&url, "url", "", "URL of the facilitator service (required)")
	supportedFlags.StringVar(&url, "u", "", "URL of the facilitator service (required)")

	// Parse flags
	supportedFlags.Parse(os.Args[2:])

	// Validate required flags
	if url == "" {
		fmt.Fprintln(os.Stderr, "Error: --url or -u flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli supported --url <url>")
		fmt.Fprintln(os.Stderr, "  x402cli supported -u <url>")
		supportedFlags.PrintDefaults()
		os.Exit(1)
	}

	// Create facilitator client and call /supported
	fc := facilitatorclient.NewFacilitatorClient(url)
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
