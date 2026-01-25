package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func proofCommand() {
	if len(os.Args) < 3 {
		printProofUsage()
		os.Exit(1)
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "gen":
		proofGenCommand()
	case "verify":
		proofVerifyCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown proof subcommand: %s\n\n", subcommand)
		printProofUsage()
		os.Exit(1)
	}
}

func printProofUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  x402cli proof <subcommand> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Subcommands:")
	fmt.Fprintln(os.Stderr, "  gen       Generate an ownership proof signature")
	fmt.Fprintln(os.Stderr, "  verify    Verify an ownership proof signature")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  x402cli proof gen -u https://api.example.com --private-key 0x...")
	fmt.Fprintln(os.Stderr, "  x402cli proof verify -u https://api.example.com --proof 0x... --address 0x...")
}

func proofGenCommand() {
	genFlags := flag.NewFlagSet("proof gen", flag.ExitOnError)
	var url, privateKeyHex string
	genFlags.StringVar(&url, "url", "", "URL of the resource to sign (required)")
	genFlags.StringVar(&url, "u", "", "URL of the resource to sign (required)")
	genFlags.StringVar(&privateKeyHex, "private-key", "", "Hex-encoded private key for signing (required)")

	genFlags.Parse(os.Args[3:])

	if url == "" || privateKeyHex == "" {
		fmt.Fprintln(os.Stderr, "Error: --url and --private-key flags are required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli proof gen -u <url> --private-key <hex>")
		genFlags.PrintDefaults()
		os.Exit(1)
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing private key: %v\n", err)
		os.Exit(1)
	}

	// EIP-191 personal message hash: keccak256("\x19Ethereum Signed Message:\n" + len(msg) + msg)
	msg := url
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

func proofVerifyCommand() {
	verifyFlags := flag.NewFlagSet("proof verify", flag.ExitOnError)
	var url, proofHex, addressHex string
	verifyFlags.StringVar(&url, "url", "", "URL that was signed (required)")
	verifyFlags.StringVar(&url, "u", "", "URL that was signed (required)")
	verifyFlags.StringVar(&proofHex, "proof", "", "Proof signature in hex (required)")
	verifyFlags.StringVar(&proofHex, "p", "", "Proof signature in hex (required)")
	verifyFlags.StringVar(&addressHex, "address", "", "Expected signer address (required)")
	verifyFlags.StringVar(&addressHex, "a", "", "Expected signer address (required)")

	verifyFlags.Parse(os.Args[3:])

	if url == "" || proofHex == "" || addressHex == "" {
		fmt.Fprintln(os.Stderr, "Error: --url, --proof, and --address flags are required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli proof verify -u <url> -p <proof-hex> -a <address>")
		verifyFlags.PrintDefaults()
		os.Exit(1)
	}

	// Decode proof signature
	proofHex = strings.TrimPrefix(proofHex, "0x")
	sig, err := hex.DecodeString(proofHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding proof: %v\n", err)
		os.Exit(1)
	}

	if len(sig) != 65 {
		fmt.Fprintf(os.Stderr, "Error: invalid signature length (expected 65 bytes, got %d)\n", len(sig))
		os.Exit(1)
	}

	// Adjust v back from Ethereum format (subtract 27)
	if sig[64] >= 27 {
		sig[64] -= 27
	}

	// Compute message hash (EIP-191)
	msg := url
	prefix := "\x19Ethereum Signed Message:\n" + strconv.Itoa(len(msg))
	hash := crypto.Keccak256Hash([]byte(prefix + msg))

	// Recover public key from signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error recovering signer: %v\n", err)
		os.Exit(1)
	}

	// Get recovered address
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	expectedAddr := common.HexToAddress(addressHex)

	// Compare
	if recoveredAddr == expectedAddr {
		fmt.Printf("Valid: signature was created by %s\n", recoveredAddr.Hex())
	} else {
		fmt.Fprintf(os.Stderr, "Invalid: signature was created by %s, expected %s\n",
			recoveredAddr.Hex(), expectedAddr.Hex())
		os.Exit(1)
	}
}
