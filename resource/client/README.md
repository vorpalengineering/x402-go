# x402 Resource Client

Client library for accessing x402-protected resources with explicit control over the payment flow.

## Features

- **Explicit payment flow**: Discrete functions for each step (check, generate, pay)
- **Full control**: Inspect requirements, check balance, decide whether to pay
- **EIP-3009 signing**: Generates `TransferWithAuthorization` payment proofs
- **CAIP-2 networks**: Supports Base, Ethereum, and any EVM chain

## Installation

```bash
go get github.com/vorpalengineering/x402-go/resource/client
```

## Quick Start

```go
package main

import (
    "fmt"
    "io"
    "log"

    "github.com/ethereum/go-ethereum/crypto"
    "github.com/vorpalengineering/x402-go/resource/client"
)

func main() {
    // Load your private key
    privateKey, err := crypto.HexToECDSA("your_private_key_hex")
    if err != nil {
        log.Fatal(err)
    }

    // Create client
    c := client.NewClient(privateKey)

    url := "https://api.example.com/protected/data"

    // Step 1: Check if payment is required
    resp, requirements, err := c.CheckForPaymentRequired("GET", url, "", nil)
    if err != nil {
        log.Fatal(err)
    }

    // Step 2: Handle payment if required
    if len(requirements) > 0 {
        selected := requirements[0]
        log.Printf("Payment required: %s on %s", selected.Amount, selected.Network)

        // Step 3: Pay for resource
        resp, err = c.PayForResource("GET", url, "", nil, &selected)
        if err != nil {
            log.Fatal(err)
        }
    }

    // Use the response
    defer resp.Body.Close()
    body, _ := io.ReadAll(resp.Body)
    fmt.Printf("Response: %s\n", body)
}
```

## API

### NewClient

```go
func NewClient(privateKey *ecdsa.PrivateKey) *Client
```

Creates a new resource client. Pass `nil` for read-only usage (checking requirements without paying).

### Browse

```go
func (c *Client) Browse(baseURL string) (*types.DiscoveryResponse, error)
```

Fetches the `/.well-known/x402` discovery endpoint for a server and returns the available protected endpoints.

**Parameters:**
- `baseURL` — The base URL of the server (trailing slash is optional)

**Returns:**
- `*types.DiscoveryResponse` — The discovery document containing protected endpoints
- `error` — Any error that occurred

### CheckForPaymentRequired

```go
func (c *Client) CheckForPaymentRequired(method, url, contentType string, body []byte) (*http.Response, []types.PaymentRequirements, error)
```

Makes an HTTP request and checks if payment is required.

**Returns:**
- `*http.Response` — The HTTP response (body is closed if 402)
- `[]types.PaymentRequirements` — Acceptable payment options (empty if not 402)
- `error` — Any error that occurred

### Requirements

```go
func (c *ResourceClient) Requirements(method, url, contentType string, body []byte, index int) (*types.PaymentRequirements, error)
```

Fetches payment requirements from a resource URL. Calls `Check()` and extracts a single `PaymentRequirements` from the accepts array.

**Parameters:**
- `method` — HTTP method (GET, POST, etc.)
- `url` — URL of the resource to check
- `contentType` — Content-Type header (empty string if not needed)
- `body` — Request body (nil for GET requests)
- `index` — Index into the accepts array (usually 0)

**Returns:**
- `*types.PaymentRequirements` — The selected payment requirements
- `error` — If resource doesn't require payment (non-402) or index is out of bounds

**Example:**
```go
c := client.NewResourceClient(nil) // No private key needed
req, err := c.Requirements("GET", "https://api.example.com/data", "", nil, 0)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Payment required: %s %s on %s\n", req.Amount, req.Asset, req.Network)
```

### GeneratePayment

```go
func (c *Client) GeneratePayment(requirements *types.PaymentRequirements) (string, error)
```

Generates a base64-encoded payment payload for the `PAYMENT-SIGNATURE` header.

**What it does:**
1. Validates payment scheme (currently only `exact`)
2. Parses amount, recipient, and token contract addresses
3. Creates EIP-3009 `TransferWithAuthorization` (random nonce, 1-hour validity window)
4. Signs with EIP-712 typed data
5. Returns base64-encoded JSON payload

### PayForResource

```go
func (c *Client) PayForResource(method, url, contentType string, body []byte, requirements *types.PaymentRequirements) (*http.Response, error)
```

Generates payment and makes the HTTP request with the `PAYMENT-SIGNATURE` header in one step.

## Usage Examples

### Discovering Protected Endpoints

```go
c := client.NewClient(nil) // No private key needed for discovery
baseURL := "https://api.example.com"

discovery, err := c.Browse(baseURL)
if err != nil {
    log.Fatal(err)
}

for _, endpoint := range discovery.Endpoints {
    fmt.Printf("%s %s - %s\n", endpoint.Method, endpoint.Path, endpoint.Description)
}
```

### Selecting a Payment Option

```go
c := client.NewClient(privateKey)
url := "https://api.example.com/data"

resp, requirements, err := c.CheckForPaymentRequired("GET", url, "", nil)
if err != nil {
    log.Fatal(err)
}

if len(requirements) > 0 {
    // Select preferred network
    var selected *types.PaymentRequirements
    for _, req := range requirements {
        if req.Network == "eip155:8453" && req.Scheme == "exact" {
            selected = &req
            break
        }
    }

    if selected == nil {
        log.Fatal("No compatible payment option")
    }

    log.Printf("Paying %s on %s", selected.Amount, selected.Network)

    resp, err = c.PayForResource("GET", url, "", nil, selected)
    if err != nil {
        log.Fatal(err)
    }
}

defer resp.Body.Close()
```

### POST Request with Payment

```go
c := client.NewClient(privateKey)
url := "https://api.example.com/submit"
body := []byte(`{"name": "Alice"}`)

resp, requirements, err := c.CheckForPaymentRequired("POST", url, "application/json", body)
if err != nil {
    log.Fatal(err)
}

if len(requirements) > 0 {
    resp, err = c.PayForResource("POST", url, "application/json", body, &requirements[0])
    if err != nil {
        log.Fatal(err)
    }
}

defer resp.Body.Close()
```

### Generate Payment Separately

```go
c := client.NewClient(privateKey)

_, requirements, err := c.CheckForPaymentRequired("GET", url, "", nil)
if len(requirements) > 0 {
    paymentHeader, err := c.GeneratePayment(&requirements[0])
    if err != nil {
        log.Fatal(err)
    }

    // Manually construct request with custom headers
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("PAYMENT-SIGNATURE", paymentHeader)
    req.Header.Set("Custom-Header", "value")

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
}
```

## Payment Requirements

When a resource requires payment, `CheckForPaymentRequired()` returns:

```go
[]types.PaymentRequirements{
    {
        Scheme:            "exact",
        Network:           "eip155:8453",
        Amount:            "1000000",  // 1 USDC (6 decimals)
        PayTo:             "0x123...",
        Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
        MaxTimeoutSeconds: 120,
        Extra: map[string]any{
            "name":    "USDC",
            "version": "2",
        },
    },
}
```

## See Also

- [x402 Specification](https://github.com/coinbase/x402)
- [Resource Middleware](../middleware) — Gin middleware for protecting resources
- [Facilitator](../../facilitator) — Payment verification and settlement service
