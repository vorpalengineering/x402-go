package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	facilitatorclient "github.com/vorpalengineering/x402-go/facilitator/client"
	"github.com/vorpalengineering/x402-go/resource/client"
	"github.com/vorpalengineering/x402-go/types"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse subcommand
	subcommand := os.Args[1]

	switch subcommand {
	case "check":
		checkCommand()
	case "supported":
		supportedCommand()
	case "verify":
		verifyCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func checkCommand() {
	// Define flags for check command
	checkFlags := flag.NewFlagSet("check", flag.ExitOnError)
	var resource string
	checkFlags.StringVar(&resource, "resource", "", "URL of the resource to check (required)")
	checkFlags.StringVar(&resource, "r", "", "URL of the resource to check (required)")

	// Parse flags
	checkFlags.Parse(os.Args[2:])

	// Validate required flags
	if resource == "" {
		fmt.Fprintln(os.Stderr, "Error: --resource or -r flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli check --resource <url>")
		fmt.Fprintln(os.Stderr, "  x402cli check --r <url>")
		checkFlags.PrintDefaults()
		os.Exit(1)
	}

	// Create read-only client (no private key needed for checking)
	c := client.NewClient(nil)

	// Check if payment is required
	resp, requirements, err := c.CheckForPaymentRequired("GET", resource, "", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Print results
	fmt.Printf("Resource: %s\n", resource)
	fmt.Printf("Status: %d %s\n\n", resp.StatusCode, resp.Status)

	if len(requirements) > 0 {
		fmt.Println("Payment Required (402)")
		fmt.Println("\nAccepts:")
		for i, req := range requirements {
			if i > 0 {
				fmt.Println("\n---")
			}
			printRequirement(&req)
		}
	} else if resp.StatusCode == 200 {
		fmt.Println("✓ Resource is accessible without payment")
	} else {
		fmt.Printf("Resource returned status %d (not payment-protected)\n", resp.StatusCode)
	}
}

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
	c := facilitatorclient.NewClient(facilitatorURL)
	resp, err := c.Supported()
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

func verifyCommand() {
	// Define flags for verify command
	verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
	var facilitatorURL, file, data string
	verifyFlags.StringVar(&facilitatorURL, "facilitator", "", "URL of the facilitator service (required)")
	verifyFlags.StringVar(&facilitatorURL, "f", "", "URL of the facilitator service (required)")
	verifyFlags.StringVar(&file, "file", "", "Path to JSON file containing VerifyRequest")
	verifyFlags.StringVar(&data, "data", "", "Inline JSON string containing VerifyRequest")
	verifyFlags.StringVar(&data, "d", "", "Inline JSON string containing VerifyRequest")

	// Parse flags
	verifyFlags.Parse(os.Args[2:])

	// Validate required flags
	if facilitatorURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --facilitator or -f flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli verify --facilitator <url> --file <path>")
		fmt.Fprintln(os.Stderr, "  x402cli verify -f <url> -d '<json>'")
		verifyFlags.PrintDefaults()
		os.Exit(1)
	}

	if file == "" && data == "" {
		fmt.Fprintln(os.Stderr, "Error: either --file or --data/-d is required")
		verifyFlags.PrintDefaults()
		os.Exit(1)
	}
	if file != "" && data != "" {
		fmt.Fprintln(os.Stderr, "Error: --file and --data/-d are mutually exclusive")
		os.Exit(1)
	}

	// Read request JSON
	var jsonData []byte
	if file != "" {
		var err error
		jsonData, err = os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
			os.Exit(1)
		}
	} else {
		jsonData = []byte(data)
	}

	// Unmarshal into VerifyRequest
	var req types.VerifyRequest
	if err := json.Unmarshal(jsonData, &req); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
		os.Exit(1)
	}

	// Call facilitator /verify
	c := facilitatorclient.NewClient(facilitatorURL)
	resp, err := c.Verify(&req)
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

func printRequirement(req *types.PaymentRequirements) {
	// Pretty-print the payment requirement as JSON
	jsonBytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting requirement: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "x402cli - CLI tool for interacting with x402-protected resources")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  x402cli <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  check       Check if a resource requires payment")
	fmt.Fprintln(os.Stderr, "  supported   Query a facilitator for supported schemes/networks")
	fmt.Fprintln(os.Stderr, "  verify      Verify a payment payload against a facilitator")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  x402cli check --resource http://localhost:3000/api/data")
	fmt.Fprintln(os.Stderr, "  x402cli supported --facilitator http://localhost:8080")
	fmt.Fprintln(os.Stderr, "  x402cli verify -f http://localhost:8080 --file request.json")
}
