# x402-go

Go implementation of the x402 protocol for verifiable payments.

## Packages

### Client (`/client`)

Client library for interacting with x402 facilitators.

```go
import "github.com/vorpalengineering/x402-go/client"

// Create new x402 client
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

# Run with default config
go run ./cmd/facilitator

# Or with a custom config path
go run ./cmd/facilitator --config=path/to/config.facilitator.yaml
```

The service will start on port 8080 with the following endpoints:
- `GET /supported` - Get supported scheme-network combinations
- `POST /verify` - Verify payment payloads
- `POST /settle` - Settle payments on-chain

### Middleware (`/middleware`)

Standalone Gin middleware for adding x402 payment verification to any Gin-based API.

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

## x402 Protocol

This implementation follows the [x402 specification](https://github.com/coinbase/x402) for verifiable on-chain payments.
