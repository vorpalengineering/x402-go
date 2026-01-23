package facilitator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/types"
)

type Facilitator struct {
	config       *FacilitatorConfig
	router       *gin.Engine
	rpcClients   map[string]*ethclient.Client
	rpcClientsMu sync.RWMutex
}

func NewFacilitator(config *FacilitatorConfig) *Facilitator {
	// Set Gin mode based on log level
	if config.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.Default()

	// Create Facilitator instance
	f := &Facilitator{
		config:     config,
		router:     router,
		rpcClients: make(map[string]*ethclient.Client),
	}

	// Register routes
	f.registerRoutes()

	return f
}

func (f *Facilitator) Run(ctx context.Context) error {
	// Initialize RPC connections
	log.Println("Initializing RPC connections...")
	if err := f.DialRPCClients(); err != nil {
		return fmt.Errorf("failed to initialize RPC clients: %w", err)
	}
	log.Println("RPC connections established")

	// Start server
	addr := fmt.Sprintf("%s:%d", f.config.Server.Host, f.config.Server.Port)
	log.Printf("Starting x402 Facilitator service on %s", addr)
	log.Printf("Supported Schemes: %v", f.config.Supported)

	// Create HTTP server with our router
	srv := &http.Server{
		Addr:    addr,
		Handler: f.router,
	}

	// Channel to receive server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- fmt.Errorf("failed to start server: %w", err)
		}
	}()

	// Wait for context cancellation or server error
	select {
	case err := <-serverErrors:
		return err
	case <-ctx.Done():
		log.Println("Shutting down facilitator service...")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		log.Println("Facilitator service stopped")
		return nil
	}
}

func (f *Facilitator) Close() {
	f.closeAllRPCClients()
}

func (f *Facilitator) DialRPCClients() error {
	// Acquire write lock
	f.rpcClientsMu.Lock()
	defer f.rpcClientsMu.Unlock()

	// Dial eth client for each network in config
	for network := range f.config.Networks {
		networkCfg, err := f.config.GetNetworkConfig(network)
		if err != nil {
			return fmt.Errorf("failed to get config for %s: %w", network, err)
		}

		client, err := ethclient.Dial(networkCfg.RpcUrl)
		if err != nil {
			return fmt.Errorf("failed to connect to %s RPC: %w", network, err)
		}

		f.rpcClients[network] = client
	}

	return nil
}

func (f *Facilitator) getRPCClient(network string) (*ethclient.Client, error) {
	// Acquire read lock
	f.rpcClientsMu.RLock()
	if client, exists := f.rpcClients[network]; exists {
		f.rpcClientsMu.RUnlock()
		return client, nil
	}
	f.rpcClientsMu.RUnlock()

	// Lazy creation if not initialized
	f.rpcClientsMu.Lock()
	defer f.rpcClientsMu.Unlock()

	if client, exists := f.rpcClients[network]; exists {
		return client, nil
	}

	networkCfg, err := f.config.GetNetworkConfig(network)
	if err != nil {
		return nil, err
	}

	client, err := ethclient.Dial(networkCfg.RpcUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	f.rpcClients[network] = client
	return client, nil
}

func (f *Facilitator) closeAllRPCClients() {
	// Acquire write lock
	f.rpcClientsMu.Lock()
	defer f.rpcClientsMu.Unlock()

	// Close every eth client connection in pool
	for _, client := range f.rpcClients {
		client.Close()
	}
	f.rpcClients = make(map[string]*ethclient.Client)
}

func (f *Facilitator) registerRoutes() {
	f.router.POST("/verify", f.handleVerify)
	f.router.POST("/settle", f.handleSettle)
	f.router.GET("/supported", f.handleSupported)
}

func (f *Facilitator) handleVerify(ginCtx *gin.Context) {
	// Decode request
	var req types.VerifyRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check scheme-network pair is supported
	if !f.config.IsSupported(req.PaymentRequirements.Scheme, req.PaymentRequirements.Network) {
		res := types.VerifyResponse{
			IsValid:       false,
			InvalidReason: fmt.Sprintf("unsupported scheme-network: %s-%s", req.PaymentRequirements.Scheme, req.PaymentRequirements.Network),
		}
		ginCtx.JSON(http.StatusOK, res)
		return
	}

	// Extract context from HTTP request
	ctx := ginCtx.Request.Context()

	// Verify request
	isValid, invalidReason := f.verifyPayment(ctx, &req.PaymentPayload, &req.PaymentRequirements)

	// Craft response
	res := types.VerifyResponse{
		IsValid:       isValid,
		InvalidReason: invalidReason,
	}

	ginCtx.JSON(http.StatusOK, res)
}

func (f *Facilitator) handleSettle(ginCtx *gin.Context) {
	// Decode request
	var req types.SettleRequest
	if err := ginCtx.ShouldBindJSON(&req); err != nil {
		ginCtx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Extract context from HTTP request
	ctx := ginCtx.Request.Context()

	// Settle request
	resp := f.settlePayment(ctx, &req.PaymentPayload, &req.PaymentRequirements)

	ginCtx.JSON(http.StatusOK, resp)
}

func (f *Facilitator) handleSupported(ctx *gin.Context) {
	res := types.SupportedResponse{
		Kinds:      f.config.Supported,
		Extensions: []string{},
		Signers: map[string][]string{
			"eip155:*": []string{
				f.config.Signer.Address.String(),
			},
			"solana:*": []string{},
		},
	}

	ctx.JSON(http.StatusOK, res)
}
