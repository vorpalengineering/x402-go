package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"
)

func proofCommand() {
	// Define flags
	proofFlags := flag.NewFlagSet("proof", flag.ExitOnError)
	var resource, privateKeyHex string
	proofFlags.StringVar(&resource, "resource", "", "URL of the resource to sign (required)")
	proofFlags.StringVar(&resource, "r", "", "URL of the resource to sign (required)")
	proofFlags.StringVar(&privateKeyHex, "private-key", "", "Hex-encoded private key for signing (required)")

	// Parse flags
	proofFlags.Parse(os.Args[2:])

	// Validate required flags
	if resource == "" || privateKeyHex == "" {
		fmt.Fprintln(os.Stderr, "Error: --resource and --private-key flags are required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli proof -r <url> --private-key <hex>")
		proofFlags.PrintDefaults()
		os.Exit(1)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing private key: %v\n", err)
		os.Exit(1)
	}

	// EIP-191 personal message hash: keccak256("\x19Ethereum Signed Message:\n" + len(msg) + msg)
	msg := resource
	prefix := "\x19Ethereum Signed Message:\n" + strconv.Itoa(len(msg))
	hash := crypto.Keccak256Hash([]byte(prefix + msg))

	// Sign
	sig, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing: %v\n", err)
		os.Exit(1)
	}

	// Adjust v for Ethereum (add 27)
	sig[64] += 27

	fmt.Println("0x" + hex.EncodeToString(sig))
}
