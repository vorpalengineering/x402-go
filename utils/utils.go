package utils

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
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

func GetChainID(network string) (*big.Int, error) {
	// network string is in CAIP-2 format (e.g. "eip155:8453")
	substrings := strings.Split(network, ":")
	if len(substrings) != 2 {
		return nil, fmt.Errorf("invalid CAIP-2 network string")
	}
	networkId := substrings[1]
	chainId, ok := new(big.Int).SetString(networkId, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse CAIP-2 network string: %s", network)
	}
	return chainId, nil
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

func BuildEIP712TypedData(auth *types.ExactEVMSchemeAuthorization, requirements *types.PaymentRequirements) (*apitypes.TypedData, error) {
	// Parse value as big.Int
	value := new(big.Int)
	value.SetString(auth.Value, 10)

	// Get Chain ID
	chainID, err := GetChainID(requirements.Network)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain id: %w", err)
	}

	// Get EIP712 Domain data from payment requirements extra field
	name, ok := requirements.Extra["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("missing EIP712 Domain name in extra field")
	}
	version, ok := requirements.Extra["version"].(string)
	if !ok || version == "" {
		return nil, fmt.Errorf("missing EIP712 Domain version in extra field")
	}

	return &apitypes.TypedData{
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
			Name:              name,    // This should match the token contract
			Version:           version, // USDC version
			ChainId:           (*math.HexOrDecimal256)(chainID),
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
	}, nil
}

func SignEIP3009(auth *types.ExactEVMSchemeAuthorization, privateKey *ecdsa.PrivateKey, asset, domainName, domainVersion string, chainID int64) (string, error) {
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
