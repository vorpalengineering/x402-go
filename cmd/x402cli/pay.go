package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/vorpalengineering/x402-go/types"
)

func payCommand() {
	// Define flags
	payFlags := flag.NewFlagSet("pay", flag.ExitOnError)
	var resource, method, payloadInput, requirementsInput, output, data string
	payFlags.StringVar(&resource, "resource", "", "URL of the resource to pay for (required)")
	payFlags.StringVar(&resource, "r", "", "URL of the resource to pay for (required)")
	payFlags.StringVar(&method, "method", "GET", "HTTP method (GET or POST)")
	payFlags.StringVar(&method, "m", "GET", "HTTP method (GET or POST)")
	payFlags.StringVar(&payloadInput, "payload", "", "Inner payload as JSON or file path (required)")
	payFlags.StringVar(&payloadInput, "p", "", "Inner payload as JSON or file path (required)")
	payFlags.StringVar(&requirementsInput, "requirements", "", "PaymentRequirements as JSON or file path (required)")
	payFlags.StringVar(&requirementsInput, "req", "", "PaymentRequirements as JSON or file path (required)")
	payFlags.StringVar(&output, "output", "", "File path to write response body")
	payFlags.StringVar(&output, "o", "", "File path to write response body")
	payFlags.StringVar(&data, "data", "", "Request body as JSON string or file path")
	payFlags.StringVar(&data, "d", "", "Request body as JSON string or file path")

	// Parse flags
	payFlags.Parse(os.Args[2:])

	// Validate method
	if method != "GET" && method != "POST" {
		fmt.Fprintf(os.Stderr, "Error: --method must be GET or POST, got %s\n", method)
		os.Exit(1)
	}

	// Validate required flags
	if resource == "" || payloadInput == "" || requirementsInput == "" {
		fmt.Fprintln(os.Stderr, "Error: --resource, --payload, and --requirements flags are all required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli pay -r <url> -p <payload-json|file> --req <requirements-json|file>")
		payFlags.PrintDefaults()
		os.Exit(1)
	}

	// Parse inner payload
	payloadData := readJSONOrFile(payloadInput)
	var innerPayload map[string]any
	if err := json.Unmarshal(payloadData, &innerPayload); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing payload JSON: %v\n", err)
		os.Exit(1)
	}

	// Parse requirements
	requirementsData := readJSONOrFile(requirementsInput)
	var requirements types.PaymentRequirements
	if err := json.Unmarshal(requirementsData, &requirements); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing requirements JSON: %v\n", err)
		os.Exit(1)
	}

	// Construct full PaymentPayload
	fullPayload := types.PaymentPayload{
		X402Version: 2,
		Accepted:    requirements,
		Payload:     innerPayload,
	}

	// Marshal to JSON, then base64 encode
	payloadJSON, err := json.Marshal(fullPayload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding payment payload: %v\n", err)
		os.Exit(1)
	}
	paymentHeader := base64.StdEncoding.EncodeToString(payloadJSON)

	// Make HTTP request with PAYMENT-SIGNATURE header
	var reqBody io.Reader
	if data != "" {
		dataBytes := readJSONOrFile(data)
		reqBody = bytes.NewReader(dataBytes)
	}
	req, err := http.NewRequest(method, resource, reqBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}
	req.Header.Set("PAYMENT-SIGNATURE", paymentHeader)
	if data != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Handle response based on status
	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		// Success — decode PAYMENT-RESPONSE header if present
		if prHeader := resp.Header.Get("PAYMENT-RESPONSE"); prHeader != "" {
			decoded, err := base64.StdEncoding.DecodeString(prHeader)
			if err == nil {
				var settleResp types.SettleResponse
				if json.Unmarshal(decoded, &settleResp) == nil {
					prettySettle, _ := json.MarshalIndent(settleResp, "", "  ")
					fmt.Fprintf(os.Stderr, "Settlement: %s\n", string(prettySettle))
				}
			}
		}

		// Output response body
		if output != "" {
			if err := os.WriteFile(output, body, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Response written to %s\n", output)
		} else {
			fmt.Print(string(body))
		}

	case resp.StatusCode == http.StatusPaymentRequired:
		// Payment failed — print the PaymentRequired response
		fmt.Fprintf(os.Stderr, "Payment required (402)\n")
		// Try to pretty-print if it's JSON
		if strings.Contains(resp.Header.Get("Content-Type"), "json") {
			var prettyJSON json.RawMessage
			if json.Unmarshal(body, &prettyJSON) == nil {
				indented, _ := json.MarshalIndent(prettyJSON, "", "  ")
				fmt.Println(string(indented))
				return
			}
		}
		fmt.Print(string(body))

	default:
		fmt.Fprintf(os.Stderr, "Unexpected status: %d %s\n", resp.StatusCode, resp.Status)
		fmt.Print(string(body))
	}
}
