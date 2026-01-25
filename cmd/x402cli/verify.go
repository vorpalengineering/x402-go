package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	facilitatorclient "github.com/vorpalengineering/x402-go/facilitator/client"
	"github.com/vorpalengineering/x402-go/types"
)

func verifyCommand() {
	// Define flags for verify command
	verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
	var facilitatorURL, payloadInput, requirementInput string
	verifyFlags.StringVar(&facilitatorURL, "facilitator", "", "URL of the facilitator service (required)")
	verifyFlags.StringVar(&facilitatorURL, "f", "", "URL of the facilitator service (required)")
	verifyFlags.StringVar(&payloadInput, "payload", "", "Payload object as JSON string or file path (required)")
	verifyFlags.StringVar(&payloadInput, "p", "", "Payload object as JSON string or file path (required)")
	verifyFlags.StringVar(&requirementInput, "requirement", "", "PaymentRequirements as JSON string or file path (required)")
	verifyFlags.StringVar(&requirementInput, "r", "", "PaymentRequirements as JSON string or file path (required)")

	// Parse flags
	verifyFlags.Parse(os.Args[2:])

	// Validate required flags
	if facilitatorURL == "" || payloadInput == "" || requirementInput == "" {
		fmt.Fprintln(os.Stderr, "Error: --facilitator, --payload, and --requirement flags are all required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli verify -f <url> -p <json|file> -r <json|file>")
		verifyFlags.PrintDefaults()
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

	// Construct VerifyRequest
	req := types.VerifyRequest{
		PaymentPayload: types.PaymentPayload{
			X402Version: 2,
			Accepted:    requirements,
			Payload:     payloadMap,
		},
		PaymentRequirements: requirements,
	}

	// Call facilitator /verify
	fc := facilitatorclient.NewFacilitatorClient(facilitatorURL)
	resp, err := fc.Verify(&req)
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
