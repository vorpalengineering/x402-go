package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/vorpalengineering/x402-go/resource/client"
	"github.com/vorpalengineering/x402-go/types"
)

func requirementsCommand() {
	// Define flags
	reqFlags := flag.NewFlagSet("requirements", flag.ExitOnError)
	var output, url, method, data, scheme, network, amount, asset, payTo, extraName, extraVersion string
	var maxTimeout, index int
	reqFlags.StringVar(&output, "output", "", "File path to write JSON output")
	reqFlags.StringVar(&output, "o", "", "File path to write JSON output")
	reqFlags.StringVar(&url, "url", "", "URL of resource to fetch requirements from")
	reqFlags.StringVar(&url, "u", "", "URL of resource to fetch requirements from")
	reqFlags.StringVar(&method, "method", "GET", "HTTP method to use when fetching requirements")
	reqFlags.StringVar(&method, "m", "GET", "HTTP method to use when fetching requirements")
	reqFlags.StringVar(&data, "data", "", "Request body data")
	reqFlags.StringVar(&data, "d", "", "Request body data")
	reqFlags.IntVar(&index, "index", 0, "Index into accepts array (default: 0)")
	reqFlags.IntVar(&index, "i", 0, "Index into accepts array (default: 0)")
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
	var req types.PaymentRequirements

	if url != "" {
		// Fetch requirements from resource server
		rc := client.NewResourceClient(nil)
		resp, paymentRequired, err := rc.Check(method, url, "", []byte(data))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		if paymentRequired == nil {
			fmt.Fprintf(os.Stderr, "Error: resource returned status %d (not payment-protected)\n", resp.StatusCode)
			os.Exit(1)
		}

		if index >= len(paymentRequired.Accepts) {
			fmt.Fprintf(os.Stderr, "Error: index %d out of bounds (accepts array has %d entries)\n", index, len(paymentRequired.Accepts))
			os.Exit(1)
		}

		req = paymentRequired.Accepts[index]
	}

	// Apply individual flag overrides
	if scheme != "" {
		req.Scheme = scheme
	}
	if network != "" {
		req.Network = network
	}
	if amount != "" {
		req.Amount = amount
	}
	if asset != "" {
		req.Asset = asset
	}
	if payTo != "" {
		req.PayTo = payTo
	}
	if maxTimeout != 0 {
		req.MaxTimeoutSeconds = maxTimeout
	}
	if extraName != "" || extraVersion != "" {
		if req.Extra == nil {
			req.Extra = map[string]any{}
		}
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
