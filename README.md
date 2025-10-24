# x402-go

Go implementation of the x402 protocol for verifiable payments.

## Packages

### Client (`/client`)

Client library for interacting with x402 facilitators.

```go
import "github.com/vorpalengineering/x402-go/client"

// Create new x402 client
c := client.New("http://localhost:8080")

// Verify Payment
resp, err := c.Verify(&types.VerifyRequest{...})

// Settle Payment
res, err = c.Settle(&types.SettleRequest{...})
```

### Facilitator (`/facilitator`)

Facilitator service implementation providing payment verification and settlement.

Run the facilitator service:
```
go run ./cmd/facilitator
```

The service will start on port 8080 with the following endpoints:
- POST /verify - Verify payment payloads
- POST /settle - Settle payments on-chain

### Types (`/types`)

Shared types used by both client and facilitator packages.

```go
// Verify types
type VerifyRequest struct {
	PaymentPayload interface{} `json:"payment_payload"`
	Requirements   interface{} `json:"requirements"`
}
type VerifyResponse struct {
	Valid   bool   `json:"valid"`
	Message string `json:"message,omitempty"`
}

// Settle types
type SettleRequest struct {
	PaymentPayload interface{} `json:"payment_payload"`
}
type SettleResponse struct {
	TxHash  string `json:"tx_hash,omitempty"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
```