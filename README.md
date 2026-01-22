# x402-go

Go implementation of the x402 protocol for verifiable payments.

## Packages

### CLI Tool (`/cmd/x402cli`)

Command-line tool for interacting with x402-protected resources.

```bash
# Check if a resource requires payment
go run ./cmd/x402cli check --resource http://localhost:3000/api/data

# Or build and install
go build ./cmd/x402cli
./x402cli check --resource http://localhost:3000/api/data

# Or install to $GOPATH/bin
go install ./cmd/x402cli
x402cli check --resource http://localhost:3000/api/data
```

**Example output:**
```
Resource: http://localhost:3000/api/data
Status: 402 Payment Required

Payment Required (402)

Accepts:
{
  "scheme": "exact",
  "network": "base",
  "maxAmountRequired": "1000000",
  "payTo": "0x123...",
  "asset": "0x833...",
  "resource": "/api/data",
  "description": "API access",
  "maxTimeoutSeconds": 120
}
```

### Client (`/client`)

Client library for accessing x402-protected resources with explicit payment flow control.

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/vorpalengineering/x402-go/client"
)

// Load your private key
privateKey, err := crypto.HexToECDSA("your_private_key_hex")
if err != nil {
    log.Fatal(err)
}

// Create resource client
c := client.NewClient(privateKey)

// Step 1: Check if resource requires payment
resp, requirements, err := c.CheckForPaymentRequired("GET", "https://api.example.com/data", "", nil)
if err != nil {
    log.Fatal(err)
}

// Step 2: If payment required, inspect requirements and decide to pay
if len(requirements) > 0 {
    selected := requirements[0]

    // Check balance, validate amount, etc.
    log.Printf("Payment required: %s %s", selected.MaxAmountRequired, selected.Network)

    // Step 3: Pay for resource
    resp, err = c.PayForResource("GET", "https://api.example.com/data", "", nil, &selected)
    if err != nil {
        log.Fatal(err)
    }
}

// Use the response
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
```

### Middleware (`/middleware`)

Middleware for adding x402 payment verification to any Gin-based API.

```go
import "github.com/vorpalengineering/x402-go/middleware"

// Configure middleware
x402 := middleware.NewX402Middleware(&middleware.Config{
    FacilitatorURL: "http://localhost:8080",
    DefaultRequirements: types.PaymentRequirements{
        Scheme:            "exact",
        Network:           "base",
        MaxAmountRequired: "1000000",
        PayTo:             "0x123...", // Your seller address here
        Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
        // ... other fields
    },
    ProtectedPaths: []string{"/api/*"},
})

// Apply to your Gin router
router.Use(x402.Handler())
```

See [middleware/README.md](./middleware/README.md) for detailed documentation.

### Facilitator (`/facilitator`)

Facilitator service implementation providing payment verification and settlement.

Make sure you have set your X402_FACILITATOR_PRIVATE_KEY env variable first.

```bash
export X402_FACILITATOR_PRIVATE_KEY=0x123abc
```

Run the facilitator service:
```bash
# Copy example config first
cp config.facilitator.example.yaml config.facilitator.yaml
# Edit config.facilitator.yaml with your settings

# Run with default config path
go run ./cmd/facilitator

# Or with a custom config path
go run ./cmd/facilitator --config=path/to/config.facilitator.yaml
```

The service will start on the configured port with the following endpoints:
- `GET /supported` - Get supported scheme-network combinations
- `POST /verify` - Verify payment payloads
- `POST /settle` - Settle payments on-chain

### Facilitator Client (`/facilitator/client`)

Client library for interacting with x402 facilitators.

```go
import "github.com/vorpalengineering/x402-go/facilitator/client"

// Create new facilitator client
c := client.NewClient("http://localhost:8080")

// Get supported schemes
supported, err := c.Supported()
if err != nil {
    log.Fatal(err)
}

// Verify Payment
verifyReq := &types.VerifyRequest{
    X402Version: 1,
    PaymentHeader: "base64EncodedPaymentHeader",
    PaymentRequirements: types.PaymentRequirements{
        Scheme: "exact",
        Network: "base",
        // ... other fields
    },
}
verifyResp, err := c.Verify(verifyReq)
if verifyResp.IsValid {
    // Payment is valid
}

// Settle Payment
settleReq := &types.SettleRequest{
    X402Version: 1,
    PaymentHeader: "base64EncodedPaymentHeader",
    PaymentRequirements: types.PaymentRequirements{
        Scheme: "exact",
        Network: "base",
        // ... other fields
    },
}
settleResp, err := c.Settle(settleReq)
if settleResp.Success {
    // Payment is successful
    fmt.Printf("Transaction hash: %s\n", settleResp.TxHash)
}
```

## x402 Protocol

This implementation follows the [x402 specification](https://github.com/coinbase/x402) for verifiable on-chain payments.
