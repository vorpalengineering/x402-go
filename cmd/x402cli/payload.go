package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vorpalengineering/x402-go/types"
	"github.com/vorpalengineering/x402-go/utils"
)

func payloadCommand() {
	// Define flags
	payloadFlags := flag.NewFlagSet("payload", flag.ExitOnError)
	var output, privateKeyHex, from, to, value, nonce string
	var validAfter, validBefore, validDuration int64
	var asset, domainName, domainVersion string
	var chainID int64
	payloadFlags.StringVar(&output, "output", "", "File path to write JSON output")
	payloadFlags.StringVar(&output, "o", "", "File path to write JSON output")
	payloadFlags.StringVar(&privateKeyHex, "private-key", "", "Hex-encoded private key for signing")
	payloadFlags.StringVar(&from, "from", "", "Payer address (derived from private key if omitted)")
	payloadFlags.StringVar(&to, "to", "", "Recipient address (required)")
	payloadFlags.StringVar(&value, "value", "", "Amount in smallest unit (required)")
	payloadFlags.Int64Var(&validAfter, "valid-after", 0, "Unix timestamp for validity start (default: now)")
	payloadFlags.Int64Var(&validBefore, "valid-before", 0, "Unix timestamp for validity end (default: now + 10min)")
	payloadFlags.Int64Var(&validDuration, "valid-duration", 0, "Validity duration in seconds (alternative to --valid-before)")
	payloadFlags.StringVar(&nonce, "nonce", "", "Hex-encoded bytes32 nonce (default: random)")
	payloadFlags.StringVar(&asset, "asset", "", "Token contract address (required with --private-key)")
	payloadFlags.StringVar(&domainName, "name", "", "EIP-712 domain name (required with --private-key)")
	payloadFlags.StringVar(&domainVersion, "version", "", "EIP-712 domain version (required with --private-key)")
	payloadFlags.Int64Var(&chainID, "chain-id", 0, "Chain ID (required with --private-key)")
	var requirementsInput string
	payloadFlags.StringVar(&requirementsInput, "requirements", "", "PaymentRequirements as JSON or file path")
	payloadFlags.StringVar(&requirementsInput, "req", "", "PaymentRequirements as JSON or file path")
	payloadFlags.StringVar(&requirementsInput, "r", "", "PaymentRequirements as JSON or file path")

	// Parse flags
	payloadFlags.Parse(os.Args[2:])

	// Apply requirements as defaults (individual flags override)
	if requirementsInput != "" {
		reqData := readJSONOrFile(requirementsInput)
		var req types.PaymentRequirements
		if err := json.Unmarshal(reqData, &req); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing requirements JSON: %v\n", err)
			os.Exit(1)
		}
		if to == "" && req.PayTo != "" {
			to = req.PayTo
		}
		if value == "" && req.Amount != "" {
			value = req.Amount
		}
		if asset == "" && req.Asset != "" {
			asset = req.Asset
		}
		if domainName == "" {
			if name, ok := req.Extra["name"].(string); ok && name != "" {
				domainName = name
			}
		}
		if domainVersion == "" {
			if ver, ok := req.Extra["version"].(string); ok && ver != "" {
				domainVersion = ver
			}
		}
		if chainID == 0 && req.Network != "" {
			// Parse chain ID from CAIP-2 format (e.g. "eip155:84532")
			parts := strings.Split(req.Network, ":")
			if len(parts) == 2 {
				if id, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					chainID = id
				}
			}
		}
	}

	// Validate required flags
	if to == "" || value == "" {
		fmt.Fprintln(os.Stderr, "Error: --to and --value flags are required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli payload --to <address> --value <amount> [options]")
		payloadFlags.PrintDefaults()
		os.Exit(1)
	}

	// Resolve private key and from address
	var privateKey *ecdsa.PrivateKey
	if privateKeyHex != "" {
		if asset == "" || domainName == "" || domainVersion == "" || chainID == 0 {
			fmt.Fprintln(os.Stderr, "Error: --asset, --name, --version, and --chain-id are required when --private-key is provided")
			os.Exit(1)
		}
		pk, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing private key: %v\n", err)
			os.Exit(1)
		}
		privateKey = pk
		if from == "" {
			from = crypto.PubkeyToAddress(pk.PublicKey).Hex()
		}
	}

	// Resolve time window
	now := time.Now().Unix()
	if validAfter == 0 {
		validAfter = now
	}
	if validDuration > 0 {
		validBefore = validAfter + validDuration
	} else if validBefore == 0 {
		validBefore = validAfter + 600 // default 10 minutes
	}

	// Resolve nonce
	var nonceHex string
	if nonce != "" {
		nonceHex = nonce
	} else {
		var nonceBytes [32]byte
		if _, err := rand.Read(nonceBytes[:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating nonce: %v\n", err)
			os.Exit(1)
		}
		nonceHex = "0x" + hex.EncodeToString(nonceBytes[:])
	}

	// Build authorization
	auth := types.ExactEVMSchemeAuthorization{
		From:        from,
		To:          to,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonceHex,
	}

	// Sign if private key is provided and from is set
	signature := "0x" + strings.Repeat("00", 65)
	if privateKey != nil && from != "" {
		sig, err := utils.SignEIP3009(&auth, privateKey, asset, domainName, domainVersion, chainID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: signing failed: %v (using placeholder)\n", err)
		} else {
			signature = sig
		}
	}

	// Build output payload map
	payloadMap := map[string]any{
		"signature":     signature,
		"authorization": auth,
	}

	// Marshal only the inner payload (signature + authorization)
	jsonBytes, err := json.MarshalIndent(payloadMap, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting payload: %v\n", err)
		os.Exit(1)
	}

	// Output
	if output != "" {
		if err := os.WriteFile(output, jsonBytes, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Payload written to %s\n", output)
	} else {
		fmt.Println(string(jsonBytes))
	}
}

