package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/vorpalengineering/x402-go/types"
)

func requirementsCommand() {
	// Define flags
	reqFlags := flag.NewFlagSet("requirements", flag.ExitOnError)
	var output, scheme, network, amount, asset, payTo, extraName, extraVersion string
	var maxTimeout int
	reqFlags.StringVar(&output, "output", "", "File path to write JSON output")
	reqFlags.StringVar(&output, "o", "", "File path to write JSON output")
	reqFlags.StringVar(&scheme, "scheme", "", "Payment scheme (e.g. exact)")
	reqFlags.StringVar(&network, "network", "", "CAIP-2 network (e.g. eip155:84532)")
	reqFlags.StringVar(&amount, "amount", "", "Amount in smallest unit")
	reqFlags.StringVar(&asset, "asset", "", "Token contract address")
	reqFlags.StringVar(&payTo, "pay-to", "", "Recipient address")
	reqFlags.IntVar(&maxTimeout, "max-timeout", 0, "Max timeout in seconds")
	reqFlags.StringVar(&extraName, "extra-name", "", "EIP-712 domain name (e.g. USD Coin)")
	reqFlags.StringVar(&extraVersion, "extra-version", "", "EIP-712 domain version (e.g. 2)")

	// Parse flags
	reqFlags.Parse(os.Args[2:])

	// Build requirements
	req := types.PaymentRequirements{
		Scheme:            scheme,
		Network:           network,
		Amount:            amount,
		Asset:             asset,
		PayTo:             payTo,
		MaxTimeoutSeconds: maxTimeout,
	}

	// Only include Extra if at least one extra field is provided
	if extraName != "" || extraVersion != "" {
		req.Extra = map[string]any{}
		if extraName != "" {
			req.Extra["name"] = extraName
		}
		if extraVersion != "" {
			req.Extra["version"] = extraVersion
		}
	}

	// Marshal to JSON
	jsonBytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting requirements: %v\n", err)
		os.Exit(1)
	}

	// Output
	if output != "" {
		if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Requirements written to %s\n", output)
	} else {
		fmt.Println(string(jsonBytes))
	}
}
