package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/vorpalengineering/x402-go/types"
)

func browseCommand() {
	browseFlags := flag.NewFlagSet("browse", flag.ExitOnError)
	var baseURL, output string
	browseFlags.StringVar(&baseURL, "url", "", "Base URL of the server (required)")
	browseFlags.StringVar(&baseURL, "u", "", "Base URL of the server (required)")
	browseFlags.StringVar(&output, "output", "", "File path to write JSON output")
	browseFlags.StringVar(&output, "o", "", "File path to write JSON output")

	browseFlags.Parse(os.Args[2:])

	if baseURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --url or -u flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli browse --url <base-url>")
		fmt.Fprintln(os.Stderr, "  x402cli browse -u <base-url>")
		browseFlags.PrintDefaults()
		os.Exit(1)
	}

	// Build discovery URL
	discoveryURL := strings.TrimSuffix(baseURL, "/") + "/.well-known/x402"

	// Fetch discovery document
	resp, err := http.Get(discoveryURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching discovery endpoint: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Discovery endpoint returned status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		os.Exit(1)
	}

	// Parse response
	var discovery types.DiscoveryResponse
	if err := json.Unmarshal(body, &discovery); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing discovery response: %v\n", err)
		os.Exit(1)
	}

	// Marshal to pretty JSON
	jsonBytes, err := json.MarshalIndent(discovery, "", "  ")
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
		fmt.Fprintf(os.Stderr, "Discovery response written to %s\n", output)
	} else {
		fmt.Println(string(jsonBytes))
	}
}
