package facilitator

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vorpalengineering/x402-go/types"
)

func TestSupported(t *testing.T) {
	// Create test config
	testConfig := &FacilitatorConfig{
		Server: ServerConfig{
			Host: "localhost",
			Port: 4020,
		},
		Networks: map[string]NetworkConfig{
			"base": {
				RpcUrl:  "https://mainnet.base.org",
				ChainId: "8453",
			},
			"ethereum": {
				RpcUrl:  "https://eth.llamarpc.com",
				ChainId: "1",
			},
		},
		Supported: []types.SchemeNetworkPair{
			{Scheme: "exact", Network: "base"},
			{Scheme: "exact", Network: "ethereum"},
		},
		Transaction: TransactionConfig{
			TimeoutSeconds: 120,
			MaxGasPrice:    "100000000000",
		},
		Log: LogConfig{
			Level: "info",
		},
		PrivateKey: "0x0000000000000000000000000000000000000000000000000000000000000001",
	}

	// Create facilitator
	f := NewFacilitator(testConfig)

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/supported", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create a response recorder to capture the response
	recorder := httptest.NewRecorder()

	// Serve the request
	f.router.ServeHTTP(recorder, req)

	// Check the status code
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	// Parse the response body
	var response types.SupportedResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify the response contains the expected supported schemes
	expectedCount := 2
	if len(response.Kinds) != expectedCount {
		t.Errorf("Expected %d supported kinds, got %d", expectedCount, len(response.Kinds))
	}

	// Verify specific scheme-network pairs
	hasBaseExact := false
	hasEthereumExact := false

	for _, kind := range response.Kinds {
		if kind.Scheme == "exact" && kind.Network == "base" {
			hasBaseExact = true
		}
		if kind.Scheme == "exact" && kind.Network == "ethereum" {
			hasEthereumExact = true
		}
	}

	if !hasBaseExact {
		t.Error("Expected to find exact-base in supported kinds")
	}

	if !hasEthereumExact {
		t.Error("Expected to find exact-ethereum in supported kinds")
	}
}

func TestSupportedEmpty(t *testing.T) {
	// Create config with no supported schemes
	testConfig := &FacilitatorConfig{
		Supported: []types.SchemeNetworkPair{},
		Log: LogConfig{
			Level: "info",
		},
	}

	// Create facilitator
	f := NewFacilitator(testConfig)

	req, _ := http.NewRequest("GET", "/supported", nil)
	recorder := httptest.NewRecorder()

	f.router.ServeHTTP(recorder, req)

	// Should still return 200 with empty array
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, recorder.Code)
	}

	var response types.SupportedResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	if len(response.Kinds) != 0 {
		t.Errorf("Expected 0 supported kinds, got %d", len(response.Kinds))
	}
}

func TestSupportedMultipleSchemes(t *testing.T) {
	testConfig := &FacilitatorConfig{
		Supported: []types.SchemeNetworkPair{
			{Scheme: "exact", Network: "base"},
			{Scheme: "exact", Network: "ethereum"},
			{Scheme: "exact", Network: "optimism"},
			{Scheme: "subscription", Network: "base"}, // Different scheme
		},
		Log: LogConfig{
			Level: "info",
		},
	}

	// Create facilitator
	f := NewFacilitator(testConfig)

	req, _ := http.NewRequest("GET", "/supported", nil)
	recorder := httptest.NewRecorder()

	f.router.ServeHTTP(recorder, req)

	var response types.SupportedResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	if len(response.Kinds) != 4 {
		t.Errorf("Expected 4 supported kinds, got %d", len(response.Kinds))
	}
}

func TestDialRPCClients(t *testing.T) {
	testConfig := &FacilitatorConfig{
		Networks: map[string]NetworkConfig{
			"base": {
				RpcUrl:  "https://mainnet.base.org",
				ChainId: "8453",
			},
			"ethereum": {
				RpcUrl:  "https://eth.llamarpc.com",
				ChainId: "1",
			},
		},
		Log: LogConfig{
			Level: "info",
		},
	}

	// Create facilitator
	f := NewFacilitator(testConfig)
	defer f.Close()

	// Verify RPC clients are not initialized yet
	f.rpcClientsMu.RLock()
	initialCount := len(f.rpcClients)
	f.rpcClientsMu.RUnlock()

	if initialCount != 0 {
		t.Errorf("Expected 0 RPC clients before initialization, got %d", initialCount)
	}

	// Initialize RPC connections
	err := f.DialRPCClients()
	if err != nil {
		t.Fatalf("Failed to initialize RPC clients: %v", err)
	}

	// Verify RPC clients were created for each network
	f.rpcClientsMu.RLock()
	clientCount := len(f.rpcClients)
	f.rpcClientsMu.RUnlock()

	expectedCount := len(testConfig.Networks)
	if clientCount != expectedCount {
		t.Errorf("Expected %d RPC clients, got %d", expectedCount, clientCount)
	}

	// Verify we can get each RPC client
	for network := range testConfig.Networks {
		client, err := f.getRPCClient(network)
		if err != nil {
			t.Errorf("Failed to get RPC client for network %s: %v", network, err)
		}
		if client == nil {
			t.Errorf("RPC client for network %s is nil", network)
		}
	}
}
