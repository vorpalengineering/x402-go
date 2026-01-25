package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vorpalengineering/x402-go/types"
)

func TestVerify(t *testing.T) {
	t.Run("successful verification", func(t *testing.T) {
		// Create mock server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and path
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if r.URL.Path != "/verify" {
				t.Errorf("Expected /verify path, got %s", r.URL.Path)
			}

			// Decode request body
			var req types.VerifyRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("Failed to decode request: %v", err)
			}

			// Return success response
			resp := types.VerifyResponse{
				IsValid: true,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		// Create client
		fc := NewFacilitatorClient(server.URL)

		// Make request
		req := &types.VerifyRequest{
			PaymentPayload: types.PaymentPayload{
				X402Version: 2,
				Accepted: types.PaymentRequirements{
					Scheme:  "exact",
					Network: "eip155:8453",
				},
			},
			PaymentRequirements: types.PaymentRequirements{
				Scheme:  "exact",
				Network: "eip155:8453",
			},
		}

		resp, err := fc.Verify(req)
		if err != nil {
			t.Fatalf("Verify failed: %v", err)
		}

		// Assertions
		if !resp.IsValid {
			t.Errorf("Expected IsValid=true, got false")
		}
	})

	t.Run("invalid payment", func(t *testing.T) {
		// Create mock server that returns invalid
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := types.VerifyResponse{
				IsValid:       false,
				InvalidReason: "insufficient amount",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		fc := NewFacilitatorClient(server.URL)
		req := &types.VerifyRequest{}

		resp, err := fc.Verify(req)
		if err != nil {
			t.Fatalf("Verify failed: %v", err)
		}

		if resp.IsValid {
			t.Errorf("Expected IsValid=false, got true")
		}
		if resp.InvalidReason != "insufficient amount" {
			t.Errorf("Expected InvalidReason='insufficient amount', got '%s'", resp.InvalidReason)
		}
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		fc := NewFacilitatorClient(server.URL)
		req := &types.VerifyRequest{}

		_, err := fc.Verify(req)
		if err == nil {
			t.Error("Expected error for 500 status, got nil")
		}
	})
}

func TestSettle(t *testing.T) {
	t.Run("successful settlement", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and path
			if r.Method != http.MethodPost {
				t.Errorf("Expected POST request, got %s", r.Method)
			}
			if r.URL.Path != "/settle" {
				t.Errorf("Expected /settle path, got %s", r.URL.Path)
			}

			// Return success response
			resp := types.SettleResponse{
				Success:     true,
				Transaction: "0xabc123",
				Network:     "eip155:8453",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		fc := NewFacilitatorClient(server.URL)
		req := &types.SettleRequest{
			PaymentPayload: types.PaymentPayload{
				X402Version: 2,
				Accepted: types.PaymentRequirements{
					Scheme:  "exact",
					Network: "eip155:8453",
				},
			},
			PaymentRequirements: types.PaymentRequirements{
				Scheme:  "exact",
				Network: "eip155:8453",
			},
		}

		resp, err := fc.Settle(req)
		if err != nil {
			t.Fatalf("Settle failed: %v", err)
		}

		// Assertions
		if !resp.Success {
			t.Errorf("Expected Success=true, got false")
		}
		if resp.Transaction != "0xabc123" {
			t.Errorf("Expected TxHash='0xabc123', got '%s'", resp.Transaction)
		}
		if resp.Network != "eip155:8453" {
			t.Errorf("Expected Network='eip155:8453', got '%s'", resp.Network)
		}
	})

	t.Run("settlement failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := types.SettleResponse{
				Success:     false,
				ErrorReason: "transaction reverted",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		fc := NewFacilitatorClient(server.URL)
		req := &types.SettleRequest{}

		resp, err := fc.Settle(req)
		if err != nil {
			t.Fatalf("Settle failed: %v", err)
		}

		if resp.Success {
			t.Errorf("Expected Success=false, got true")
		}
		if resp.ErrorReason != "transaction reverted" {
			t.Errorf("Expected Error='transaction reverted', got '%s'", resp.ErrorReason)
		}
	})
}

func TestSupported(t *testing.T) {
	t.Run("returns supported schemes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify request method and path
			if r.Method != http.MethodGet {
				t.Errorf("Expected GET request, got %s", r.Method)
			}
			if r.URL.Path != "/supported" {
				t.Errorf("Expected /supported path, got %s", r.URL.Path)
			}

			// Return supported schemes
			resp := types.SupportedResponse{
				Kinds: []types.SupportedKind{
					{Scheme: "exact", Network: "eip155:8453"},
					{Scheme: "exact", Network: "eip155:1"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		fc := NewFacilitatorClient(server.URL)
		resp, err := fc.Supported()
		if err != nil {
			t.Fatalf("Supported failed: %v", err)
		}

		// Assertions
		if len(resp.Kinds) != 2 {
			t.Errorf("Expected 2 supported kinds, got %d", len(resp.Kinds))
		}
		if resp.Kinds[0].Scheme != "exact" {
			t.Errorf("Expected first scheme='exact', got '%s'", resp.Kinds[0].Scheme)
		}
		if resp.Kinds[0].Network != "eip155:8453" {
			t.Errorf("Expected first network='eip155:8453', got '%s'", resp.Kinds[0].Network)
		}
	})

	t.Run("empty supported list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := types.SupportedResponse{
				Kinds: []types.SupportedKind{},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		fc := NewFacilitatorClient(server.URL)
		resp, err := fc.Supported()
		if err != nil {
			t.Fatalf("Supported failed: %v", err)
		}

		if len(resp.Kinds) != 0 {
			t.Errorf("Expected 0 supported kinds, got %d", len(resp.Kinds))
		}
	})
}
