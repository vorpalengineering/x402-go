package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/vorpalengineering/x402-go/types"
)

const ERC20BalanceOfABI = `[{
	"constant": true,
	"inputs": [{"name": "account", "type": "address"}],
	"name": "balanceOf",
	"outputs": [{"name": "", "type": "uint256"}],
	"type": "function"
}]`

const EIP3009TransferWithAuthABI = `[{
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

func GetChainID(network string) *big.Int {
	switch network {
	case "base":
		return big.NewInt(8453)
	case "base-sepolia":
		return big.NewInt(84532)
	default:
		return big.NewInt(1) // Default to mainnet
	}
}

func DecodePaymentHeader(header string) (*types.PaymentPayload, error) {
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

func ExtractExactAuthorization(payload *types.PaymentPayload) (*types.ExactEVMSchemeAuthorization, error) {
	// Get authorization object
	authData, ok := payload.Payload["authorization"]
	if !ok {
		return nil, fmt.Errorf("missing authorization")
	}

	// Convert to JSON and back to struct
	authJSON, err := json.Marshal(authData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization: %w", err)
	}

	var auth types.ExactEVMSchemeAuthorization
	if err := json.Unmarshal(authJSON, &auth); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization: %w", err)
	}

	return &auth, nil
}

func ExtractVRS(signatureHex string) (v uint8, r [32]byte, s [32]byte, err error) {
	// Remove 0x prefix if present
	if len(signatureHex) > 2 && signatureHex[:2] == "0x" {
		signatureHex = signatureHex[2:]
	}

	// Decode hex signature
	signature, err := hexutil.Decode("0x" + signatureHex)
	if err != nil {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("invalid signature format: %w", err)
	}

	// Signature should be 65 bytes (r: 32, s: 32, v: 1)
	if len(signature) != 65 {
		return 0, [32]byte{}, [32]byte{}, fmt.Errorf("invalid signature length: expected 65, got %d", len(signature))
	}

	// Extract r (first 32 bytes)
	copy(r[:], signature[0:32])

	// Extract s (next 32 bytes)
	copy(s[:], signature[32:64])

	// Extract v (last byte)
	v = signature[64]

	// Ethereum uses v = 27 or 28, ensure it's in that range
	if v < 27 {
		v += 27
	}

	return v, r, s, nil
}

// TODO: add params for other tokens
func BuildEIP712TypedData(auth *types.ExactEVMSchemeAuthorization, requirements *types.PaymentRequirements) apitypes.TypedData {
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
			Name:              "USDC", // This should match the token contract
			Version:           "2",    // USDC version
			ChainId:           (*math.HexOrDecimal256)(GetChainID(requirements.Network)),
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
