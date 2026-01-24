package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	facilitatorclient "github.com/vorpalengineering/x402-go/facilitator/client"
	"github.com/vorpalengineering/x402-go/types"
)

func settleCommand() {
	// Define flags for settle command
	settleFlags := flag.NewFlagSet("settle", flag.ExitOnError)
	var facilitatorURL, payloadInput, requirementInput string
	settleFlags.StringVar(&facilitatorURL, "facilitator", "", "URL of the facilitator service (required)")
	settleFlags.StringVar(&facilitatorURL, "f", "", "URL of the facilitator service (required)")
	settleFlags.StringVar(&payloadInput, "payload", "", "Payload object as JSON string or file path (required)")
	settleFlags.StringVar(&payloadInput, "p", "", "Payload object as JSON string or file path (required)")
	settleFlags.StringVar(&requirementInput, "requirement", "", "PaymentRequirements as JSON string or file path (required)")
	settleFlags.StringVar(&requirementInput, "r", "", "PaymentRequirements as JSON string or file path (required)")

	// Parse flags
	settleFlags.Parse(os.Args[2:])

	// Validate required flags
	if facilitatorURL == "" || payloadInput == "" || requirementInput == "" {
		fmt.Fprintln(os.Stderr, "Error: --facilitator, --payload, and --requirement flags are all required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli settle -f <url> -p <json|file> -r <json|file>")
		settleFlags.PrintDefaults()
		os.Exit(1)
	}

	// Parse payload (JSON string or file path)
	payloadData := readJSONOrFile(payloadInput)
	var payloadMap map[string]any
	if err := json.Unmarshal(payloadData, &payloadMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing payload JSON: %v\n", err)
		os.Exit(1)
	}

	// Parse requirements (JSON string or file path)
	requirementData := readJSONOrFile(requirementInput)
	var requirements types.PaymentRequirements
	if err := json.Unmarshal(requirementData, &requirements); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing requirement JSON: %v\n", err)
		os.Exit(1)
	}

	// Construct SettleRequest
	req := types.SettleRequest{
		PaymentPayload: types.PaymentPayload{
			X402Version: 2,
			Accepted:    requirements,
			Payload:     payloadMap,
		},
		PaymentRequirements: requirements,
	}

	// Call facilitator /settle
	c := facilitatorclient.NewClient(facilitatorURL)
	resp, err := c.Settle(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Pretty-print the response
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}
