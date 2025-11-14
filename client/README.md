# X402 Resource Client

Client library for accessing x402-protected resources with explicit control over the payment flow.

## Features

- **Explicit payment flow**: Discrete functions for each step (check, generate, pay)
- **Full control**: Inspect requirements, check balance, decide whether to pay
- **EIP-3009 signing**: Generates USDC payment authorizations
- **Multi-network**: Supports Base, Ethereum, and more

## Installation

```bash
go get github.com/vorpalengineering/x402-go/client
```

## Quick Start

```go
package main

import (
    "fmt"
    "io"
    "log"

    "github.com/ethereum/go-ethereum/crypto"
    "github.com/vorpalengineering/x402-go/client"
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
        // Inspect requirements
        selected := requirements[0]
        log.Printf("Payment required: %s on %s", selected.MaxAmountRequired, selected.Network)

        // TODO: Check balance here before paying

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

Creates a new resource client. The private key is used to sign payment authorizations.

### CheckForPaymentRequired

```go
func (c *Client) CheckForPaymentRequired(method, url, contentType string, body []byte) (*http.Response, []types.PaymentRequirements, error)
```

Makes an HTTP request and checks if payment is required.

**Returns:**
- `*http.Response` - The HTTP response (body is closed if 402)
- `[]types.PaymentRequirements` - Array of acceptable payment options (empty if not 402)
- `error` - Any error that occurred

**Behavior:**
- If response is **200 OK**: Returns response with empty requirements
- If response is **402 Payment Required**: Parses `Accepts` field and returns requirements array
- If response is **other**: Returns response with empty requirements

This allows you to:
1. Inspect payment requirements before paying
2. Select preferred payment option (network, scheme)
3. Check token balance before committing
4. Handle non-payment responses normally

### GeneratePayment

```go
func (c *Client) GeneratePayment(requirements *types.PaymentRequirements) (string, error)
```

Generates a payment header for the given requirements without making a request.

**What it does:**
1. Validates payment scheme (currently only "exact" is supported)
2. Parses amount, recipient address, and token address
3. Creates EIP-3009 `transferWithAuthorization`:
   - Generates random nonce
   - Creates EIP-712 typed data
   - Signs with private key
   - Sets 1-hour validity window
4. Encodes as base64 JSON payment header

**Returns:**
- `string` - Base64-encoded payment header for `X-Payment` header
- `error` - Validation or signing error

Use this if you want to:
- Generate payment header separately
- Inspect the payment before sending
- Store/cache payment headers
- Manually construct HTTP requests

### PayForResource

```go
func (c *Client) PayForResource(method, url, contentType string, body []byte, requirements *types.PaymentRequirements) (*http.Response, error)
```

Generates payment and makes HTTP request with payment header in one step.

**What it does:**
1. Calls `GeneratePayment()` to create payment header
2. Makes HTTP request with `X-Payment` header set
3. Returns the response

**Use this when:**
- You've already checked requirements with `CheckForPaymentRequired()`
- You've verified you have sufficient balance
- You're ready to pay and access the resource

## Usage Examples

### Full Payment Flow

```go
c := client.NewClient(privateKey)
url := "https://api.example.com/data"

// Check for payment
resp, requirements, err := c.CheckForPaymentRequired("GET", url, "", nil)
if err != nil {
    log.Fatal(err)
}

// If payment required
if len(requirements) > 0 {
    // Select payment option (e.g., prefer Base network)
    var selected *types.PaymentRequirements
    for _, req := range requirements {
        if req.Network == "base" && req.Scheme == "exact" {
            selected = &req
            break
        }
    }

    if selected == nil {
        log.Fatal("No compatible payment option")
    }

    // Inspect amount
    log.Printf("Paying %s USDC on %s", selected.MaxAmountRequired, selected.Network)

    // TODO: Check USDC balance here

    // Pay for resource
    resp, err = c.PayForResource("GET", url, "", nil, selected)
    if err != nil {
        log.Fatal(err)
    }
}

// Use response
defer resp.Body.Close()
```

### POST Request with Payment

```go
c := client.NewClient(privateKey)
url := "https://api.example.com/submit"
body := []byte(`{"name": "Alice", "amount": 100}`)

// Check if payment required
resp, requirements, err := c.CheckForPaymentRequired("POST", url, "application/json", body)
if err != nil {
    log.Fatal(err)
}

// Pay if needed
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

// Check for payment
_, requirements, err := c.CheckForPaymentRequired("GET", url, "", nil)
if len(requirements) > 0 {
    // Generate payment header
    paymentHeader, err := c.GeneratePayment(&requirements[0])
    if err != nil {
        log.Fatal(err)
    }

    // Manually create request
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("X-Payment", paymentHeader)
    req.Header.Set("Custom-Header", "value")

    resp, _ := http.DefaultClient.Do(req)
    defer resp.Body.Close()
}
```

## Payment Flow

The three-step payment flow:

```
1. CheckForPaymentRequired()
   ├─> Makes HTTP request
   ├─> If 402: Parse requirements
   └─> Returns response + requirements

2. [Your code]
   ├─> Inspect requirements
   ├─> Check balance
   ├─> Select payment option
   └─> Decide whether to pay

3. PayForResource()
   ├─> Generate EIP-3009 authorization
   ├─> Sign with private key
   ├─> Make request with X-Payment header
   └─> Return response
```

Or use `GeneratePayment()` separately:

```
1. CheckForPaymentRequired()
2. GeneratePayment()  <- Returns payment header string
3. [Make your own HTTP request with header]
```

## Payment Requirements Structure

When a resource requires payment, `CheckForPaymentRequired()` returns:

```go
[]types.PaymentRequirements{
    {
        Scheme:            "exact",
        Network:           "base",
        MaxAmountRequired: "1000000",  // 1 USDC (6 decimals)
        PayTo:             "0x123...",
        Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC on Base
        Resource:          "/api/data",
        Description:       "API access",
        MaxTimeoutSeconds: 120,
    },
    // ... more options if server provides alternatives
}
```
