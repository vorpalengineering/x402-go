package client

import (
	"fmt"
	"net/http"

	"github.com/vorpalengineering/x402-go/types"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

type ClientOption func(*Client)

func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	return c.Do(req)
}

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

func (c *Client) handlePaymentRequired(originalReq *http.Request, paymentResp *types.PaymentRequiredResponse) (*http.Response, error) {
	// TODO: Implement payment generation logic
	// 1. Select payment requirements from Accepts
	// 2. Generate payment authorization (exact scheme, subscription, etc.)
	// 3. Create payment header
	// 4. Clone original request and add payment header
	// 5. Execute request with payment
	return nil, fmt.Errorf("not implemented")
}
