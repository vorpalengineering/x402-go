package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vorpalengineering/x402-go/types"
)

// Client handles interactions with an x402 facilitator
type Client struct {
	facilitatorURL string
	httpClient     *http.Client
}

// Creates a new x402 client
func New(facilitatorURL string) *Client {
	return &Client{
		facilitatorURL: facilitatorURL,
		httpClient:     &http.Client{},
	}
}

// Verify sends a payment verification request to the facilitator
func (c *Client) Verify(req *types.VerifyRequest) (*types.VerifyResponse, error) {
	url := fmt.Sprintf("%s/verify", c.facilitatorURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var verifyResp types.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &verifyResp, nil
}

// Settle sends a payment settlement request to the facilitator
func (c *Client) Settle(req *types.SettleRequest) (*types.SettleResponse, error) {
	url := fmt.Sprintf("%s/settle", c.facilitatorURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var settleResp types.SettleResponse
	if err := json.NewDecoder(resp.Body).Decode(&settleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &settleResp, nil
}
