package facilitator

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/vorpalengineering/x402-go/types"
	"github.com/vorpalengineering/x402-go/utils"
)

func (f *Facilitator) settlePayment(ctx context.Context, payload *types.PaymentPayload, requirements *types.PaymentRequirements) *types.SettleResponse {
	// Settle based on scheme
	switch payload.Accepted.Scheme {
	case "exact":
		return f.settleExactScheme(ctx, payload, requirements)
	default:
		return &types.SettleResponse{
			Success:     false,
			ErrorReason: fmt.Sprintf("unsupported scheme: %s", payload.Accepted.Scheme),
		}
	}
}

func (f *Facilitator) settleExactScheme(ctx context.Context, payload *types.PaymentPayload, requirements *types.PaymentRequirements) *types.SettleResponse {
	// Extract signature from payload
	signatureHex, ok := payload.Payload["signature"].(string)
	if !ok || signatureHex == "" {
		return &types.SettleResponse{
			Success:     false,
			ErrorReason: "missing signature",
		}
	}

	// Extract authorization from payload
	auth, err := utils.ExtractExactAuthorization(payload)
	if err != nil {
		return &types.SettleResponse{
			Success:     false,
			ErrorReason: fmt.Sprintf("invalid authorization: %v", err),
		}
	}

	// Get RPC client
	client, err := f.getRPCClient(requirements.Network)
	if err != nil {
		return &types.SettleResponse{
			Success:     false,
			ErrorReason: fmt.Sprintf("failed to connect to network: %v", err),
		}
	}

	// Build and send the transaction
	txHash, err := f.sendTransferWithAuthorization(ctx, client, auth, requirements, signatureHex)
	if err != nil {
		return &types.SettleResponse{
			Success:     false,
			ErrorReason: fmt.Sprintf("failed to settle payment: %v", err),
		}
	}

	// Return success response
	return &types.SettleResponse{
		Success:     true,
		Transaction: txHash,
		Network:     requirements.Network,
		Payer:       auth.From,
	}
}

func (f *Facilitator) sendTransferWithAuthorization(
	ctx context.Context,
	client *ethclient.Client,
	auth *types.ExactEVMSchemeAuthorization,
	requirements *types.PaymentRequirements,
	signatureHex string,
) (string, error) {
	// Parse the EIP-3009 ABI
	parsedABI, err := abi.JSON(strings.NewReader(utils.EIP3009TransferWithAuthABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Extract v, r, s from signature
	v, r, s, err := utils.ExtractVRS(signatureHex)
	if err != nil {
		return "", fmt.Errorf("failed to extract signature: %v", err)
	}

	// Parse addresses and value
	fromAddr := common.HexToAddress(auth.From)
	toAddr := common.HexToAddress(auth.To)
	value := new(big.Int)
	value.SetString(auth.Value, 10)

	// Parse nonce (should be bytes32)
	var authNonce [32]byte
	nonceBytes := common.FromHex(auth.Nonce)
	if len(nonceBytes) != 32 {
		return "", fmt.Errorf("invalid nonce length: expected 32 bytes, got %d", len(nonceBytes))
	}
	copy(authNonce[:], nonceBytes)

	// Encode the transferWithAuthorization call
	callData, err := parsedABI.Pack(
		"transferWithAuthorization",
		fromAddr,
		toAddr,
		value,
		big.NewInt(auth.ValidAfter),
		big.NewInt(auth.ValidBefore),
		authNonce,
		v,
		r,
		s,
	)
	if err != nil {
		return "", fmt.Errorf("failed to encode call: %v", err)
	}

	// Get nonce for facilitator address
	nonce, err := client.PendingNonceAt(ctx, f.config.Signer.Address)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get gas price: %w", err)
	}

	// Check gas price against max gas price from config
	maxGasPrice, ok := new(big.Int).SetString(f.config.Transaction.MaxGasPrice, 10)
	if !ok {
		return "", fmt.Errorf("failed to parse max gas price: %s", f.config.Transaction.MaxGasPrice)
	}

	if gasPrice.Cmp(maxGasPrice) > 0 {
		return "", fmt.Errorf("gas price too high: suggested %s wei exceeds max %s wei", gasPrice.String(), maxGasPrice.String())
	}

	// Estimate gas
	tokenAddress := common.HexToAddress(requirements.Asset)
	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: f.config.Signer.Address,
		To:   &tokenAddress,
		Data: callData,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Create transaction
	tx := ethtypes.NewTransaction(
		nonce,
		tokenAddress,
		big.NewInt(0), // No ETH value, just calling contract
		gasLimit,
		gasPrice,
		callData,
	)

	// Get chain ID
	chainID, err := utils.GetChainID(requirements.Network)
	if err != nil {
		return "", fmt.Errorf("failed to get chain id: %w", err)
	}

	// Sign transaction
	signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(chainID), f.config.Signer.PrivateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	// Return transaction hash
	return signedTx.Hash().Hex(), nil
}
