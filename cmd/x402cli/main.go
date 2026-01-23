package main

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	facilitatorclient "github.com/vorpalengineering/x402-go/facilitator/client"
	"github.com/vorpalengineering/x402-go/resource/client"
	"github.com/vorpalengineering/x402-go/types"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Parse subcommand
	subcommand := os.Args[1]

	switch subcommand {
	case "check":
		checkCommand()
	case "supported":
		supportedCommand()
	case "verify":
		verifyCommand()
	case "payload":
		payloadCommand()
	case "requirements", "req":
		requirementsCommand()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", subcommand)
		printUsage()
		os.Exit(1)
	}
}

func checkCommand() {
	// Define flags for check command
	checkFlags := flag.NewFlagSet("check", flag.ExitOnError)
	var resource string
	checkFlags.StringVar(&resource, "resource", "", "URL of the resource to check (required)")
	checkFlags.StringVar(&resource, "r", "", "URL of the resource to check (required)")

	// Parse flags
	checkFlags.Parse(os.Args[2:])

	// Validate required flags
	if resource == "" {
		fmt.Fprintln(os.Stderr, "Error: --resource or -r flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli check --resource <url>")
		fmt.Fprintln(os.Stderr, "  x402cli check --r <url>")
		checkFlags.PrintDefaults()
		os.Exit(1)
	}

	// Create read-only client (no private key needed for checking)
	c := client.NewClient(nil)

	// Check if payment is required
	resp, requirements, err := c.CheckForPaymentRequired("GET", resource, "", nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Print results
	fmt.Printf("Resource: %s\n", resource)
	fmt.Printf("Status: %d %s\n\n", resp.StatusCode, resp.Status)

	if len(requirements) > 0 {
		fmt.Println("Payment Required (402)")
		fmt.Println("\nAccepts:")
		for i, req := range requirements {
			if i > 0 {
				fmt.Println("\n---")
			}
			printRequirement(&req)
		}
	} else if resp.StatusCode == 200 {
		fmt.Println("✓ Resource is accessible without payment")
	} else {
		fmt.Printf("Resource returned status %d (not payment-protected)\n", resp.StatusCode)
	}
}

func supportedCommand() {
	// Define flags for supported command
	supportedFlags := flag.NewFlagSet("supported", flag.ExitOnError)
	var facilitatorURL string
	supportedFlags.StringVar(&facilitatorURL, "facilitator", "", "URL of the facilitator service (required)")
	supportedFlags.StringVar(&facilitatorURL, "f", "", "URL of the facilitator service (required)")

	// Parse flags
	supportedFlags.Parse(os.Args[2:])

	// Validate required flags
	if facilitatorURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --facilitator or -f flag is required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli supported --facilitator <url>")
		fmt.Fprintln(os.Stderr, "  x402cli supported -f <url>")
		supportedFlags.PrintDefaults()
		os.Exit(1)
	}

	// Create facilitator client and call /supported
	c := facilitatorclient.NewClient(facilitatorURL)
	resp, err := c.Supported()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Pretty-print the response as JSON
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

func verifyCommand() {
	// Define flags for verify command
	verifyFlags := flag.NewFlagSet("verify", flag.ExitOnError)
	var facilitatorURL, payloadInput, requirementInput string
	verifyFlags.StringVar(&facilitatorURL, "facilitator", "", "URL of the facilitator service (required)")
	verifyFlags.StringVar(&facilitatorURL, "f", "", "URL of the facilitator service (required)")
	verifyFlags.StringVar(&payloadInput, "payload", "", "Payload object as JSON string or file path (required)")
	verifyFlags.StringVar(&payloadInput, "p", "", "Payload object as JSON string or file path (required)")
	verifyFlags.StringVar(&requirementInput, "requirement", "", "PaymentRequirements as JSON string or file path (required)")
	verifyFlags.StringVar(&requirementInput, "r", "", "PaymentRequirements as JSON string or file path (required)")

	// Parse flags
	verifyFlags.Parse(os.Args[2:])

	// Validate required flags
	if facilitatorURL == "" || payloadInput == "" || requirementInput == "" {
		fmt.Fprintln(os.Stderr, "Error: --facilitator, --payload, and --requirement flags are all required")
		fmt.Fprintln(os.Stderr, "\nUsage:")
		fmt.Fprintln(os.Stderr, "  x402cli verify -f <url> -p <json|file> -r <json|file>")
		verifyFlags.PrintDefaults()
		os.Exit(1)
	}

	// Parse payload (JSON string or file path)
	payloadData := readJSONOrFile(payloadInput)
	var payloadMap map[string]any
	if err := json.Unmarshal(payloadData, &payloadMap); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing payload JSON: %v\n", err)
		os.Exit(1)
	}

	// Parse requirements (JSON string or file path)
	requirementData := readJSONOrFile(requirementInput)
	var requirements types.PaymentRequirements
	if err := json.Unmarshal(requirementData, &requirements); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing requirement JSON: %v\n", err)
		os.Exit(1)
	}

	// Construct VerifyRequest
	req := types.VerifyRequest{
		PaymentPayload: types.PaymentPayload{
			X402Version: 2,
			Accepted:    requirements,
			Payload:     payloadMap,
		},
		PaymentRequirements: requirements,
	}

	// Call facilitator /verify
	c := facilitatorclient.NewClient(facilitatorURL)
	resp, err := c.Verify(&req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Pretty-print the response
	jsonBytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting response: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(jsonBytes))
}

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
		sig, err := signEIP3009(&auth, privateKey, asset, domainName, domainVersion, chainID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: signing failed: %v (using placeholder)\n", err)
		} else {
			signature = sig
		}
	}

	// Hardcoded defaults
	const (
		defaultScheme  = "exact"
		defaultNetwork = "eip155:84532"
		defaultAsset   = "0x036CbD53842c5426634e7929541eC2318f3dCF7e"
	)

	// Assemble PaymentPayload
	payload := types.PaymentPayload{
		X402Version: 2,
		Accepted: types.PaymentRequirements{
			Scheme:            defaultScheme,
			Network:           defaultNetwork,
			Amount:            value,
			Asset:             defaultAsset,
			PayTo:             to,
			MaxTimeoutSeconds: 30,
			Extra: map[string]any{
				"name":    "USDC",
				"version": "2",
			},
		},
		Payload: map[string]any{
			"signature":     signature,
			"authorization": auth,
		},
	}

	// Marshal only the inner payload (signature + authorization)
	jsonBytes, err := json.MarshalIndent(payload.Payload, "", "  ")
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

func signEIP3009(auth *types.ExactEVMSchemeAuthorization, privateKey *ecdsa.PrivateKey, asset string, domainName string, domainVersion string, chainID int64) (string, error) {
	// Parse addresses and values
	fromAddr := common.HexToAddress(auth.From)
	toAddr := common.HexToAddress(auth.To)
	val, ok := new(big.Int).SetString(auth.Value, 10)
	if !ok {
		return "", fmt.Errorf("invalid value: %s", auth.Value)
	}
	assetAddr := common.HexToAddress(asset)

	// Decode nonce
	nonceStr := strings.TrimPrefix(auth.Nonce, "0x")
	nonceBytes, err := hex.DecodeString(nonceStr)
	if err != nil {
		return "", fmt.Errorf("invalid nonce: %w", err)
	}
	var nonce [32]byte
	copy(nonce[32-len(nonceBytes):], nonceBytes)

	// EIP-712 Domain Separator
	domainTypeHash := crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
	))
	nameHash := crypto.Keccak256Hash([]byte(domainName))
	versionHash := crypto.Keccak256Hash([]byte(domainVersion))
	domainSeparator := crypto.Keccak256Hash(
		domainTypeHash.Bytes(),
		nameHash.Bytes(),
		versionHash.Bytes(),
		common.LeftPadBytes(big.NewInt(chainID).Bytes(), 32),
		common.LeftPadBytes(assetAddr.Bytes(), 32),
	)

	// TransferWithAuthorization struct hash
	transferTypeHash := crypto.Keccak256Hash([]byte(
		"TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)",
	))
	structHash := crypto.Keccak256Hash(
		transferTypeHash.Bytes(),
		common.LeftPadBytes(fromAddr.Bytes(), 32),
		common.LeftPadBytes(toAddr.Bytes(), 32),
		common.LeftPadBytes(val.Bytes(), 32),
		common.LeftPadBytes(big.NewInt(auth.ValidAfter).Bytes(), 32),
		common.LeftPadBytes(big.NewInt(auth.ValidBefore).Bytes(), 32),
		nonce[:],
	)

	// EIP-712 message hash
	messageHash := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator.Bytes(),
		structHash.Bytes(),
	)

	// Sign
	sig, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign: %w", err)
	}

	// Adjust v for Ethereum (add 27)
	sig[64] += 27

	return "0x" + hex.EncodeToString(sig), nil
}

func printRequirement(req *types.PaymentRequirements) {
	// Pretty-print the payment requirement as JSON
	jsonBytes, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting requirement: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "x402cli - CLI tool for interacting with x402-protected resources")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  x402cli <command> [flags]")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  check       Check if a resource requires payment")
	fmt.Fprintln(os.Stderr, "  supported   Query a facilitator for supported schemes/networks")
	fmt.Fprintln(os.Stderr, "  verify      Verify a payment payload against a facilitator")
	fmt.Fprintln(os.Stderr, "  payload     Generate a payment payload with EIP-3009 authorization")
	fmt.Fprintln(os.Stderr, "  req         Generate a payment requirements object")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Examples:")
	fmt.Fprintln(os.Stderr, "  x402cli check --resource http://localhost:3000/api/data")
	fmt.Fprintln(os.Stderr, "  x402cli supported --facilitator http://localhost:8080")
	fmt.Fprintln(os.Stderr, "  x402cli verify -f http://localhost:8080 -p payload.json -r requirements.json")
	fmt.Fprintln(os.Stderr, "  x402cli payload --to 0x... --value 10000 --private-key 0x...")
	fmt.Fprintln(os.Stderr, "  x402cli req --scheme exact --network eip155:84532 --amount 10000")
}
