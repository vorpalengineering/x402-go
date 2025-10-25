package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/vorpalengineering/x402-go/types"
)

type Client struct {
	facilitatorURL string
	httpClient     *http.Client
}

func NewClient(facilitatorURL string) *Client {
	return &Client{
		facilitatorURL: facilitatorURL,
		httpClient:     &http.Client{},
	}
}

func (client *Client) Verify(req *types.VerifyRequest) (*types.VerifyResponse, error) {
	// Build verify endpoint url
	url := fmt.Sprintf("%s/verify", client.facilitatorURL)

	// Encode request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to facilitator
	resp, err := client.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode response
	var verifyResp types.VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &verifyResp, nil
}

func (client *Client) Settle(req *types.SettleRequest) (*types.SettleResponse, error) {
	// Build settle endpoint url
	url := fmt.Sprintf("%s/settle", client.facilitatorURL)

	// Encode request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make request to facilitator
	resp, err := client.httpClient.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode response
	var settleResp types.SettleResponse
	if err := json.NewDecoder(resp.Body).Decode(&settleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &settleResp, nil
}

func (client *Client) Supported() (*types.SupportedResponse, error) {
	// Build supported endpoint url
	url := fmt.Sprintf("%s/supported", client.facilitatorURL)

	// Make request to facilitator
	resp, err := client.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Decode response
	var supportedResp types.SupportedResponse
	if err := json.NewDecoder(resp.Body).Decode(&supportedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &supportedResp, nil
}
