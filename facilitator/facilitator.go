package facilitator

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/types"
)

var cfg *FacilitatorConfig

var (
	rpcClients   = make(map[string]*ethclient.Client)
	rpcClientsMu sync.RWMutex
)

func InitializeRPCClients() error {
	// Acquire write lock
	rpcClientsMu.Lock()
	defer rpcClientsMu.Unlock()

	// Dial eth client for each network in config
	for network := range cfg.Networks {
		networkCfg, err := cfg.GetNetworkConfig(network)
		if err != nil {
			return fmt.Errorf("failed to get config for %s: %w", network, err)
		}

		client, err := ethclient.Dial(networkCfg.RpcUrl)
		if err != nil {
			return fmt.Errorf("failed to connect to %s RPC: %w", network, err)
		}

		rpcClients[network] = client
	}

	return nil
}

func getRPCClient(network string) (*ethclient.Client, error) {
	// Acquire read lock
	rpcClientsMu.RLock()
	if client, exists := rpcClients[network]; exists {
		rpcClientsMu.RUnlock()
		return client, nil
	}
	rpcClientsMu.RUnlock()

	// Lazy creation if not initialized
	rpcClientsMu.Lock()
	defer rpcClientsMu.Unlock()

	if client, exists := rpcClients[network]; exists {
		return client, nil
	}

	networkCfg, err := cfg.GetNetworkConfig(network)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(networkCfg.RpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	rpcClients[network] = client
	return client, nil
}

func CloseAllRPCClients() {
	// Acquire write lock
	rpcClientsMu.Lock()
	defer rpcClientsMu.Unlock()

	// Close every eth client connection in pool
	for _, client := range rpcClients {
		client.Close()
	}
	rpcClients = make(map[string]*ethclient.Client)
}

func RegisterRoutes(router *gin.Engine, config *FacilitatorConfig) {
	cfg = config

	router.POST("/verify", handleVerify)
	router.POST("/settle", handleSettle)
	router.GET("/supported", handleSupported)
}

func handleVerify(ctx *gin.Context) {
	// Decode request
	var req types.VerifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check scheme-network pair is supported
	if !cfg.IsSupported(req.PaymentRequirements.Scheme, req.PaymentRequirements.Network) {
		res := types.VerifyResponse{
			IsValid:       false,
			InvalidReason: fmt.Sprintf("unsupported scheme-network: %s-%s", req.PaymentRequirements.Scheme, req.PaymentRequirements.Network),
		}
		ctx.JSON(http.StatusOK, res)
		return
	}

	// Verify request
	isValid, invalidReason := VerifyPayment(&req)

	// Craft response
	res := types.VerifyResponse{
		IsValid:       isValid,
		InvalidReason: invalidReason,
	}

	ctx.JSON(http.StatusOK, res)
}

func handleSettle(ctx *gin.Context) {
	// Decode request
	var req types.SettleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Settle request
	resp := SettlePayment(&req)

	ctx.JSON(http.StatusOK, resp)
}

func handleSupported(ctx *gin.Context) {
	res := types.SupportedResponse{
		Kinds: cfg.Supported,
	}

	ctx.JSON(http.StatusOK, res)
}
