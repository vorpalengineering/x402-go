package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse subcommand
	subcommand := os.Args[1]

	switch subcommand {
	case "browse":
		browseCommand()
	case "check":
		checkCommand()
	case "pay":
		payCommand()
	case "supported":
		supportedCommand()
	case "verify":
		verifyCommand()
	case "settle":
		settleCommand()
	case "payload":
		payloadCommand()
	case "requirements", "req":
		requirementsCommand()
	case "proof":
		proofCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "x402cli - CLI tool for interacting with x402-protected resources")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  x402cli <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  browse      Fetch the /.well-known/x402 discovery document")
	fmt.Fprintln(os.Stderr, "  check       Check if a resource requires payment")
	fmt.Fprintln(os.Stderr, "  pay         Pay for a resource with a payment payload")
	fmt.Fprintln(os.Stderr, "  supported   Query a facilitator for supported schemes/networks")
	fmt.Fprintln(os.Stderr, "  verify      Verify a payment payload against a facilitator")
	fmt.Fprintln(os.Stderr, "  settle      Settle a payment payload via a facilitator")
	fmt.Fprintln(os.Stderr, "  payload     Generate a payment payload with EIP-3009 authorization")
	fmt.Fprintln(os.Stderr, "  req         Generate a payment requirements object")
	fmt.Fprintln(os.Stderr, "  proof       Generate an ownership proof signature for a resource URL")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  x402cli browse -u https://api.example.com")
	fmt.Fprintln(os.Stderr, "  x402cli check -u http://localhost:3000/api/data")
	fmt.Fprintln(os.Stderr, "  x402cli pay -u http://localhost:3000/api/data -p payload.json --req requirements.json")
	fmt.Fprintln(os.Stderr, "  x402cli supported --facilitator http://localhost:8080")
	fmt.Fprintln(os.Stderr, "  x402cli verify -f http://localhost:8080 -p payload.json -r requirements.json")
	fmt.Fprintln(os.Stderr, "  x402cli settle -f http://localhost:8080 -p payload.json -r requirements.json")
	fmt.Fprintln(os.Stderr, "  x402cli payload --to 0x... --value 10000 --private-key 0x...")
	fmt.Fprintln(os.Stderr, "  x402cli req --scheme exact --network eip155:84532 --amount 10000")
	fmt.Fprintln(os.Stderr, "  x402cli proof -u https://api.example.com --private-key 0x...")
}
