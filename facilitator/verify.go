package facilitator

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/vorpalengineering/x402-go/types"
)

const erc20BalanceOfABI = `[{
	"constant": true,
	"inputs": [{"name": "account", "type": "address"}],
	"name": "balanceOf",
	"outputs": [{"name": "", "type": "uint256"}],
	"type": "function"
}]`

const eip3009TransferWithAuthABI = `[{
	"inputs": [
		{"name": "from", "type": "address"},
		{"name": "to", "type": "address"},
		{"name": "value", "type": "uint256"},
		{"name": "validAfter", "type": "uint256"},
		{"name": "validBefore", "type": "uint256"},
		{"name": "nonce", "type": "bytes32"},
		{"name": "v", "type": "uint8"},
		{"name": "r", "type": "bytes32"},
		{"name": "s", "type": "bytes32"}
	],
	"name": "transferWithAuthorization",
	"outputs": [],
	"stateMutability": "nonpayable",
	"type": "function"
}]`

func VerifyPayment(req *types.VerifyRequest) (bool, string) {
	// Decode the payment header from base64
	paymentPayload, err := decodePaymentHeader(req.PaymentHeader)
	if err != nil {
		return false, fmt.Sprintf("failed to decode payment header: %v", err)
	}

	// Verify based on scheme
	switch paymentPayload.Scheme {
	case "exact":
		return verifyExactScheme(paymentPayload, &req.PaymentRequirements)
	default:
		return false, fmt.Sprintf("unsupported scheme: %s", paymentPayload.Scheme)
	}
}

func decodePaymentHeader(header string) (*types.PaymentPayload, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(header)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	// Parse JSON
	var payload types.PaymentPayload
	if err := json.Unmarshal(decoded, &payload); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	return &payload, nil
}

func verifyExactScheme(payload *types.PaymentPayload, requirements *types.PaymentRequirements) (bool, string) {
	// Extract signature from payload (we need it for multiple steps)
	signatureHex, ok := payload.Payload["signature"].(string)
	if !ok || signatureHex == "" {
		return false, "missing signature"
	}

	// Extract authorization from payload
	auth, err := extractExactAuthorization(payload)
	if err != nil {
		return false, fmt.Sprintf("invalid authorization: %v", err)
	}

	// Step 1: Signature Validation
	if valid, reason := verifySignature(auth, payload, requirements); !valid {
		return false, reason
	}

	// Step 2: Balance Verification
	if valid, reason := verifyBalance(auth, requirements); !valid {
		return false, reason
	}

	// Step 3: Amount Validation
	if valid, reason := verifyAmount(auth, requirements); !valid {
		return false, reason
	}

	// Step 4: Time Window Check
	if valid, reason := verifyTimeWindow(auth); !valid {
		return false, reason
	}

	// Step 5: Parameter Matching
	if valid, reason := verifyParameters(auth, requirements); !valid {
		return false, reason
	}

	// Step 6: Transaction Simulation
	if valid, reason := simulateTransaction(auth, requirements, signatureHex); !valid {
		return false, reason
	}

	return true, ""
}

func extractExactAuthorization(payload *types.PaymentPayload) (*types.ExactSchemeAuthorization, error) {
	// Get signature
	_, ok := payload.Payload["signature"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid signature")
	}

	// Get authorization object
	authData, ok := payload.Payload["authorization"]
	if !ok {
		return nil, fmt.Errorf("missing authorization")
	}

	// Convert to JSON and back to struct (handles type conversions)
	authJSON, err := json.Marshal(authData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization: %w", err)
	}

	var auth types.ExactSchemeAuthorization
	if err := json.Unmarshal(authJSON, &auth); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization: %w", err)
	}

	// Store signature in a place we can access it
	// (We'll need to add a Signature field to ExactSchemeAuthorization)
	// For now, we'll handle it separately

	return &auth, nil
}

func verifySignature(auth *types.ExactSchemeAuthorization, payload *types.PaymentPayload, requirements *types.PaymentRequirements) (bool, string) {
	// Step 1: Extract signature from payload
	signatureHex, ok := payload.Payload["signature"].(string)
	if !ok || signatureHex == "" {
		return false, "missing signature"
	}

	// Remove 0x prefix if present
	if len(signatureHex) > 2 && signatureHex[:2] == "0x" {
		signatureHex = signatureHex[2:]
	}

	// Decode hex signature
	signature, err := hexutil.Decode("0x" + signatureHex)
	if err != nil {
		return false, fmt.Sprintf("invalid signature format: %v", err)
	}

	// Signature should be 65 bytes (r: 32, s: 32, v: 1)
	if len(signature) != 65 {
		return false, fmt.Sprintf("invalid signature length: expected 65, got %d", len(signature))
	}

	// Step 2: Build EIP-712 typed data
	typedData := buildEIP712TypedData(auth, requirements)

	// Step 3: Hash the typed data according to EIP-712
	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return false, fmt.Sprintf("failed to hash domain: %v", err)
	}

	messageHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return false, fmt.Sprintf("failed to hash message: %v", err)
	}

	// EIP-712 final hash: keccak256("\x19\x01" ‖ domainSeparator ‖ messageHash)
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(messageHash)))
	hash := crypto.Keccak256Hash(rawData)

	// Step 4: Adjust v value (Ethereum uses 27/28, but ecrecover expects 0/1)
	if signature[64] == 27 || signature[64] == 28 {
		signature[64] -= 27
	}

	// Step 5: Recover the public key from the signature
	pubKey, err := crypto.SigToPub(hash.Bytes(), signature)
	if err != nil {
		return false, fmt.Sprintf("failed to recover public key: %v", err)
	}

	// Step 6: Get the address from the public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Step 7: Verify the recovered address matches auth.From
	expectedAddr := common.HexToAddress(auth.From)
	if recoveredAddr != expectedAddr {
		return false, fmt.Sprintf("signature mismatch: recovered %s, expected %s",
			recoveredAddr.Hex(), expectedAddr.Hex())
	}

	return true, ""
}

func buildEIP712TypedData(auth *types.ExactSchemeAuthorization, requirements *types.PaymentRequirements) apitypes.TypedData {
	// Parse value as big.Int
	value := new(big.Int)
	value.SetString(auth.Value, 10)

	return apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"TransferWithAuthorization": []apitypes.Type{
				{Name: "from", Type: "address"},
				{Name: "to", Type: "address"},
				{Name: "value", Type: "uint256"},
				{Name: "validAfter", Type: "uint256"},
				{Name: "validBefore", Type: "uint256"},
				{Name: "nonce", Type: "bytes32"},
			},
		},
		PrimaryType: "TransferWithAuthorization",
		Domain: apitypes.TypedDataDomain{
			Name:              "USD Coin", // This should match the token contract
			Version:           "2",        // USDC version
			ChainId:           (*math.HexOrDecimal256)(getChainID(requirements.Network)),
			VerifyingContract: requirements.Asset,
		},
		Message: apitypes.TypedDataMessage{
			"from":        auth.From,
			"to":          auth.To,
			"value":       value.String(),
			"validAfter":  fmt.Sprintf("%d", auth.ValidAfter),
			"validBefore": fmt.Sprintf("%d", auth.ValidBefore),
			"nonce":       auth.Nonce,
		},
	}
}

func verifyBalance(auth *types.ExactSchemeAuthorization, requirements *types.PaymentRequirements) (bool, string) {
	// Parse the payment amount
	paymentAmount, ok := new(big.Int).SetString(auth.Value, 10)
	if !ok {
		return false, "invalid payment amount format"
	}

	// Get RPC client for the network
	client, err := getRPCClient(requirements.Network)
	if err != nil {
		return false, fmt.Sprintf("failed to connect to network: %v", err)
	}

	// Parse the ERC-20 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20BalanceOfABI))
	if err != nil {
		return false, fmt.Sprintf("failed to parse ABI: %v", err)
	}

	// Encode the balanceOf call
	fromAddress := common.HexToAddress(auth.From)
	callData, err := parsedABI.Pack("balanceOf", fromAddress)
	if err != nil {
		return false, fmt.Sprintf("failed to encode balanceOf call: %v", err)
	}

	// Create the call message
	tokenAddress := common.HexToAddress(requirements.Asset)
	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: callData,
	}

	// Execute the call
	ctx := context.Background()
	result, err := client.CallContract(ctx, msg, nil) // nil = latest block
	if err != nil {
		return false, fmt.Sprintf("failed to call balanceOf: %v", err)
	}

	// Decode the result
	var balance *big.Int
	err = parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return false, fmt.Sprintf("failed to decode balance: %v", err)
	}

	// Check if balance is sufficient
	if balance.Cmp(paymentAmount) < 0 {
		return false, fmt.Sprintf("insufficient balance: has %s, needs %s", balance.String(), paymentAmount.String())
	}

	return true, ""
}

func verifyAmount(auth *types.ExactSchemeAuthorization, requirements *types.PaymentRequirements) (bool, string) {
	// Parse amounts as big.Int for safe comparison
	paymentAmount, ok := new(big.Int).SetString(auth.Value, 10)
	if !ok {
		return false, "invalid payment amount format"
	}

	requiredAmount, ok := new(big.Int).SetString(requirements.MaxAmountRequired, 10)
	if !ok {
		return false, "invalid required amount format"
	}

	// Payment must be >= required amount
	if paymentAmount.Cmp(requiredAmount) < 0 {
		return false, fmt.Sprintf("insufficient amount: got %s, required %s", auth.Value, requirements.MaxAmountRequired)
	}

	return true, ""
}

func verifyTimeWindow(auth *types.ExactSchemeAuthorization) (bool, string) {
	now := time.Now().Unix()

	// Check validAfter
	if now < auth.ValidAfter {
		return false, fmt.Sprintf("payment not yet valid (valid after %d)", auth.ValidAfter)
	}

	// Check validBefore
	if now > auth.ValidBefore {
		return false, fmt.Sprintf("payment expired (valid before %d)", auth.ValidBefore)
	}

	return true, ""
}

func verifyParameters(auth *types.ExactSchemeAuthorization, requirements *types.PaymentRequirements) (bool, string) {
	// Verify recipient address matches
	if auth.To != requirements.PayTo {
		return false, fmt.Sprintf("recipient mismatch: got %s, expected %s", auth.To, requirements.PayTo)
	}

	// Additional parameter checks can be added here

	return true, ""
}

func simulateTransaction(auth *types.ExactSchemeAuthorization, requirements *types.PaymentRequirements, signatureHex string) (bool, string) {
	// Get RPC client
	client, err := getRPCClient(requirements.Network)
	if err != nil {
		return false, fmt.Sprintf("failed to connect to network: %v", err)
	}

	// Parse the EIP-3009 ABI
	parsedABI, err := abi.JSON(strings.NewReader(eip3009TransferWithAuthABI))
	if err != nil {
		return false, fmt.Sprintf("failed to parse ABI: %v", err)
	}

	// Extract v, r, s from signature
	v, r, s, err := extractVRS(signatureHex)
	if err != nil {
		return false, fmt.Sprintf("failed to extract signature components: %v", err)
	}

	// Parse addresses and value
	fromAddr := common.HexToAddress(auth.From)
	toAddr := common.HexToAddress(auth.To)
	value := new(big.Int)
	value.SetString(auth.Value, 10)

	// Parse nonce (should be bytes32)
	var nonce [32]byte
	nonceBytes, err := hexutil.Decode(auth.Nonce)
	if err != nil {
		return false, fmt.Sprintf("invalid nonce format: %v", err)
	}
	copy(nonce[:], nonceBytes)

	// Encode the transferWithAuthorization call
	callData, err := parsedABI.Pack(
		"transferWithAuthorization",
		fromAddr,
		toAddr,
		value,
		big.NewInt(auth.ValidAfter),
		big.NewInt(auth.ValidBefore),
		nonce,
		v,
		r,
		s,
	)
	if err != nil {
		return false, fmt.Sprintf("failed to encode call: %v", err)
	}

	// Create the call message
	tokenAddress := common.HexToAddress(requirements.Asset)
	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: callData,
	}

	// Simulate the transaction
	ctx := context.Background()
	_, err = client.CallContract(ctx, msg, nil) // nil = latest block
	if err != nil {
		return false, fmt.Sprintf("transaction would fail: %v", err)
	}

	// If we got here, the transaction simulation succeeded
	return true, ""
}
