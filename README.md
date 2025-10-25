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

Run the facilitator service:
```bash
go run ./cmd/facilitator
```

The service will start on port 8080 with the following endpoints:
- `GET /supported` - Get supported scheme-network combinations
- `POST /verify` - Verify payment payloads
- `POST /settle` - Settle payments on-chain

## x402 Protocol

This implementation follows the [x402 specification](https://github.com/coinbase/x402) for verifiable on-chain payments.
