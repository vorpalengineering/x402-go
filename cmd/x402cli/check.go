package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/vorpalengineering/x402-go/resource/client"
)

func checkCommand() {
	// Define flags for check command
	checkFlags := flag.NewFlagSet("check", flag.ExitOnError)
	var resource, output, method, data string
	checkFlags.StringVar(&resource, "resource", "", "URL of the resource to check (required)")
	checkFlags.StringVar(&resource, "r", "", "URL of the resource to check (required)")
	checkFlags.StringVar(&output, "output", "", "File path to write JSON output")
	checkFlags.StringVar(&output, "o", "", "File path to write JSON output")
	checkFlags.StringVar(&method, "method", "GET", "HTTP method (GET or POST)")
	checkFlags.StringVar(&method, "m", "GET", "HTTP method (GET or POST)")
	checkFlags.StringVar(&data, "data", "", "Request body data")
	checkFlags.StringVar(&data, "d", "", "Request body data")

	// Parse flags
	checkFlags.Parse(os.Args[2:])

	// Validate method
	if method != "GET" && method != "POST" {
		fmt.Fprintf(os.Stderr, "Error: --method must be GET or POST, got %s\n", method)
		os.Exit(1)
	}

	// Validate required flags
	if resource == "" {
		fmt.Fprintln(os.Stderr, "Error: --resource or -r flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli check --resource <url>")
		fmt.Fprintln(os.Stderr, "  x402cli check -r <url>")
		checkFlags.PrintDefaults()
		os.Exit(1)
	}

	// Create read-only client (no private key needed for checking)
	rc := client.NewResourceClient(nil)

	// Check if payment is required
	resp, paymentRequired, err := rc.Check(method, resource, "", []byte(data))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if paymentRequired == nil {
		fmt.Fprintf(os.Stderr, "Resource returned status %d (not payment-protected)\n", resp.StatusCode)
		return
	}

	// Marshal the PaymentRequired response to pretty JSON
	jsonBytes, err := json.MarshalIndent(paymentRequired, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting response: %v\n", err)
		os.Exit(1)
	}

	// Output
	if output != "" {
		if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Payment requirements written to %s\n", output)
	} else {
		fmt.Println(string(jsonBytes))
	}
}
