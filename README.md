# x402-go

Go implementation of the x402 protocol for verifiable payments.

## Packages

### Client (`/client`)

Client library for interacting with x402 facilitators.

```go
import "github.com/vorpalengineering/x402-go/client"

c := client.New("http://localhost:8080")
resp, err := c.Verify(&types.VerifyRequest{...})
```

### Facilitator (/facilitator)

Facilitator service implementation providing payment verification and settlement.

Run the facilitator service:
go run ./cmd/facilitator

Types (/types)

Shared types used by both client and facilitator packages.

Running the Facilitator

go run ./cmd/facilitator

The service will start on port 8080 with the following endpoints:
- POST /verify - Verify payment payloads
- POST /settle - Settle payments on-chain
EOF

Install dependencies

go get github.com/gin-gonic/gin
go mod tidy

Build the facilitator (optional)

go build -o bin/facilitator ./cmd/facilitator

Run the facilitator service

go run ./cmd/facilitator