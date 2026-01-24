package main

import (
	"fmt"
	"os"
	"strings"
)

// readJSONOrFile returns JSON bytes from either an inline JSON string or a file path.
func readJSONOrFile(input string) []byte {
	if strings.HasPrefix(strings.TrimSpace(input), "{") {
		return []byte(input)
	}
	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", input, err)
		os.Exit(1)
	}
	return data
}
