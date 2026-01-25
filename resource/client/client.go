package client

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/vorpalengineering/x402-go/types"
	"github.com/vorpalengineering/x402-go/utils"
)

type ResourceClient struct {
	httpClient *http.Client
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

func NewResourceClient(privateKey *ecdsa.PrivateKey) *ResourceClient {
	rc := &ResourceClient{
		httpClient: &http.Client{},
		privateKey: privateKey,
	}

	// Only derive address if we have a private key
	if privateKey != nil {
		rc.address = crypto.PubkeyToAddress(privateKey.PublicKey)
	}

	return rc
}

func (rc *ResourceClient) Browse(baseURL string) (*types.DiscoveryResponse, error) {
	// Build discovery URL
	discoveryURL := strings.TrimSuffix(baseURL, "/") + "/.well-known/x402"

	// Make HTTP GET request
	resp, err := rc.httpClient.Get(discoveryURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("discovery endpoint returned status %d", resp.StatusCode)
	}

	// Parse response
	var discovery types.DiscoveryResponse
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to parse discovery response: %w", err)
	}

	return &discovery, nil
}

func (rc *ResourceClient) Check(
	method string,
	url string,
	contentType string,
	body []byte,
) (*http.Response, *types.PaymentRequired, error) {
	// Make HTTP request
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	// If not 402, return response with no payment required
	if resp.StatusCode != http.StatusPaymentRequired {
		return resp, nil, nil
	}

	// Parse 402 Payment Required response
	respBody, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read 402 response: %w", err)
	}

	var paymentResp types.PaymentRequired
	if err := json.Unmarshal(respBody, &paymentResp); err != nil {
		return nil, nil, fmt.Errorf("failed to parse payment requirements: %w", err)
	}

	return resp, &paymentResp, nil
}

func (rc *ResourceClient) Pay(
	method string,
	url string,
	contentType string,
	body []byte,
	requirements *types.PaymentRequirements,
) (*http.Response, error) {
	// Generate payment header
	paymentHeader, err := rc.GeneratePayment(requirements)
	if err != nil {
		return nil, err
	}

	// Make request with payment header
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("PAYMENT-SIGNATURE", paymentHeader)

	resp, err := rc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request with payment failed: %w", err)
	}

	return resp, nil
}

func (rc *ResourceClient) GeneratePayment(requirements *types.PaymentRequirements) (string, error) {
	// Check that we have a private key for payment generation
	if rc.privateKey == nil {
		return "", fmt.Errorf("cannot generate payment: client was created without a private key")
	}

	// Validate scheme
	if requirements.Scheme != "exact" {
		return "", fmt.Errorf("unsupported payment scheme: %s (only 'exact' is supported)", requirements.Scheme)
	}

	// Parse amount
	value, ok := new(big.Int).SetString(requirements.Amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid amount: %s", requirements.Amount)
	}

	// Parse recipient address
	toAddress := common.HexToAddress(requirements.PayTo)
	if toAddress == (common.Address{}) {
		return "", fmt.Errorf("invalid recipient address: %s", requirements.PayTo)
	}

	// Parse asset (token contract) address
	assetAddress := common.HexToAddress(requirements.Asset)
	if assetAddress == (common.Address{}) {
		return "", fmt.Errorf("invalid asset address: %s", requirements.Asset)
	}

	// Get chain ID for the network
	chainID, err := utils.GetChainID(requirements.Network)
	if err != nil {
		return "", fmt.Errorf("failed to get chain id: %s", err)
	}

	// Generate EIP-3009 authorization
	auth, err := CreateEIP3009Authorization(
		rc.privateKey,
		rc.address,
		toAddress,
		value,
		assetAddress,
		chainID.Int64(),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create EIP-3009 authorization: %w", err)
	}

	// Build payment payload
	payload := types.PaymentPayload{
		X402Version: 2,
		Accepted:    *requirements,
		Payload: map[string]any{
			"signature": encodeSignature(auth.V, auth.R, auth.S),
			"authorization": types.ExactEVMSchemeAuthorization{
				From:        auth.From.Hex(),
				To:          auth.To.Hex(),
				Value:       auth.Value.String(),
				ValidAfter:  auth.ValidAfter.Int64(),
				ValidBefore: auth.ValidBefore.Int64(),
				Nonce:       "0x" + hex.EncodeToString(auth.Nonce[:]),
			},
		},
	}

	// Encode to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payment payload: %w", err)
	}

	// Encode to base64
	return base64.StdEncoding.EncodeToString(payloadJSON), nil
}

// TODO: make this a generic function for all tokens
func createDomainSeparator(verifyingContract common.Address, chainID *big.Int, name string, version string) common.Hash {
	// EIP-712 Domain typeHash
	// keccak256("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)")
	domainTypeHash := crypto.Keccak256Hash([]byte(
		"EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)",
	))

	// name = "USD Coin"
	nameHash := crypto.Keccak256Hash([]byte(name))

	// version = "2"
	versionHash := crypto.Keccak256Hash([]byte(version))

	// Encode domain separator
	domainSeparator := crypto.Keccak256Hash(
		domainTypeHash.Bytes(),
		nameHash.Bytes(),
		versionHash.Bytes(),
		common.LeftPadBytes(chainID.Bytes(), 32),
		common.LeftPadBytes(verifyingContract.Bytes(), 32),
	)

	return domainSeparator
}

// encodeSignature converts EIP-3009 signature components (v, r, s) to hex string
func encodeSignature(v uint8, r, s [32]byte) string {
	sig := make([]byte, 65)
	copy(sig[0:32], r[:])
	copy(sig[32:64], s[:])
	sig[64] = v - 27 // Convert from Ethereum's v (27/28) to standard (0/1)
	return "0x" + hex.EncodeToString(sig)
}

func generateNonce() ([32]byte, error) {
	var nonce [32]byte
	_, err := rand.Read(nonce[:])
	if err != nil {
		return nonce, fmt.Errorf("failed to generate nonce: %w", err)
	}
	return nonce, nil
}

func CreateEIP3009Authorization(
	privateKey *ecdsa.PrivateKey,
	from common.Address,
	to common.Address,
	value *big.Int,
	usdcContract common.Address,
	chainID int64,
) (*types.EIP3009Authorization, error) {
	// Generate nonce
	nonce, err := generateNonce()
	if err != nil {
		return nil, err
	}

	// Set validity period (valid from 1 hour ago to 1 hour from now)
	validAfter := big.NewInt(time.Now().Add(-1 * time.Hour).Unix())
	validBefore := big.NewInt(time.Now().Add(1 * time.Hour).Unix())

	// EIP-712 Domain Separator
	domainSeparator := createDomainSeparator(usdcContract, big.NewInt(chainID), "USDC", "2")

	// Transfer With Authorization typeHash
	// keccak256("TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)")
	transferTypeHash := crypto.Keccak256Hash([]byte(
		"TransferWithAuthorization(address from,address to,uint256 value,uint256 validAfter,uint256 validBefore,bytes32 nonce)",
	))

	// Encode the struct hash
	structHash := crypto.Keccak256Hash(
		transferTypeHash.Bytes(),
		common.LeftPadBytes(from.Bytes(), 32),
		common.LeftPadBytes(to.Bytes(), 32),
		common.LeftPadBytes(value.Bytes(), 32),
		common.LeftPadBytes(validAfter.Bytes(), 32),
		common.LeftPadBytes(validBefore.Bytes(), 32),
		nonce[:],
	)

	// Create the final message hash (EIP-712)
	messageHash := crypto.Keccak256Hash(
		[]byte("\x19\x01"),
		domainSeparator.Bytes(),
		structHash.Bytes(),
	)

	// Sign the message
	signature, err := crypto.Sign(messageHash.Bytes(), privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign message: %w", err)
	}

	// Extract v, r, s from signature
	var r, s [32]byte
	copy(r[:], signature[0:32])
	copy(s[:], signature[32:64])
	v := signature[64] + 27 // Add 27 for Ethereum compatibility

	auth := &types.EIP3009Authorization{
		From:        from,
		To:          to,
		Value:       value,
		ValidAfter:  validAfter,
		ValidBefore: validBefore,
		Nonce:       nonce,
		V:           v,
		R:           r,
		S:           s,
	}

	return auth, nil
}
