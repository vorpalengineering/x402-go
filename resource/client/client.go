package client

import (
	"fmt"
	"net/http"

	"github.com/vorpalengineering/x402-go/types"
)

// Client is an HTTP client that handles x402 payment protocol
// for accessing protected resources
type Client struct {
	httpClient *http.Client
	// TODO: Add payment strategy (exact, subscription, etc.)
	// TODO: Add facilitator client for payment verification
}

// NewClient creates a new resource client
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// ClientOption is a function that configures a Client
type ClientOption func(*Client)

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// Get performs a GET request to a protected resource
func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.Do(req)
}

// Do executes an HTTP request with x402 payment handling
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Make initial request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// If 402 Payment Required, handle payment flow
	if resp.StatusCode == http.StatusPaymentRequired {
		resp.Body.Close()
		return nil, fmt.Errorf("payment required but payment handling not yet implemented")
		// TODO: Implement payment flow:
		// 1. Parse PaymentRequiredResponse from response body
		// 2. Generate payment based on requirements
		// 3. Add payment header to request
		// 4. Retry request with payment
	}

	return resp, nil
}

// handlePaymentRequired processes a 402 response and retries with payment
func (c *Client) handlePaymentRequired(originalReq *http.Request, paymentResp *types.PaymentRequiredResponse) (*http.Response, error) {
	// TODO: Implement payment generation logic
	// 1. Select payment requirements from Accepts
	// 2. Generate payment authorization (exact scheme, subscription, etc.)
	// 3. Create payment header
	// 4. Clone original request and add payment header
	// 5. Execute request with payment
	return nil, fmt.Errorf("not implemented")
}
