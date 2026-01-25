package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/vorpalengineering/x402-go/resource/client"
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

	// Create client and fetch discovery document
	rc := client.NewResourceClient(nil)
	discovery, err := rc.Browse(baseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Marshal to pretty JSON
	jsonBytes, err := json.MarshalIndent(*discovery, "", "  ")
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
