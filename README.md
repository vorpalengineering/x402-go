# x402-go

Go implementation of the [x402 protocol](https://github.com/coinbase/x402) (v2) for verifiable on-chain payments.

## Using x402-go...

### As A Buyer

1. Browse available resources with `x402cli browse` or `ResourceClient.Browse()`
2. Check available payment requirements with `x402 check` or `ResourceClient.Check()`
3. Select and fetch payment requirement with `x402 req` or `ResourceClient.Requirements()`
4. Create and sign payment payload with `x402 payload` or `ResourceClient.Payload()`
5. Pay for resource with `x402 pay` or `ResourceClient.Pay()`

### As A Seller

1. Integrate the `X402Middleware` handler into your Gin API.
2. Configure the middleware to protect API routes with payment requirements.
3. Generate and verify ownership proofs with `x402cli proof gen` and `x402cli proof verify`
4. Test your configuration with the `x402cli` or `ResourceClient`.

OR

1. Integrate the `FacilitatorClient` directly into your Golang API.
2. Configure the client to handle payment actions.
3. Check facilitator supported networks and schemes with `FacilitatorClient.Supported()`
4. Verify payment payload validity with `FacilitatorClient.Verify()`
5. Settle payment payload with `FacilitatorClient.Settle()`

### As An Operator

1. Configure and run the `Facilitator` service.

## Project Structure

```
x402-go/
├── cmd/
│   ├── facilitator/       # Facilitator service binary
│   └── x402cli/           # CLI tool for checking x402-protected resources
├── facilitator/           # Facilitator server, verification, and settlement logic
│   ├── client/            # Client library for interacting with a facilitator
│   └── config.example.yaml
├── resource/              # Resource server components
│   ├── client/            # Client library for accessing x402-protected resources
│   └── middleware/        # Gin middleware for protecting resources with x402
├── types/                 # Shared x402 protocol types
└── utils/                 # Shared utilities (EIP-712, CAIP-2 parsing, etc.)
```

## Packages

### CLI Tool (`cmd/x402cli`)

Command-line tool for checking x402-protected resources.

```bash
# Build
go build ./cmd/x402cli

# Get supported data from facilitator
./x402cli supported -u http://localhost:4020

```

**Example output:**
```
{
  "kinds": [
    {
      "x402Version": 0,
      "scheme": "exact",
      "network": "eip155:84532"
    }
  ],
  "extensions": [],
  "signers": {
    "eip155:*": [
      "0x123..."
    ],
    "solana:*": []
  }
}
```

### Resource Client (`resource/client`)

Client library for accessing x402-protected resources. Handles EIP-3009 `TransferWithAuthorization` signing.

```go
import (
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/vorpalengineering/x402-go/resource/client"
)

// Load your private key
privateKey, err := crypto.HexToECDSA("your_private_key_hex")
if err != nil {
    log.Fatal(err)
}

// Create resource client
rc := client.NewResourceClient(privateKey)

// Step 1: Check if resource requires payment
resp, paymentRequired, err := rc.Check("GET", "https://api.example.com/data", "", nil)
if err != nil {
    log.Fatal(err)
}

// Step 2: If payment required, inspect requirements and decide to pay
if paymentRequired != nil {
    selected := paymentRequired.Accepts[0]

    log.Printf("Payment required: %s on %s", selected.Amount, selected.Network)

    // Step 3: Pay for resource (generates EIP-3009 authorization and sends PAYMENT-SIGNATURE header)
    resp, err = rc.Pay("GET", "https://api.example.com/data", "", nil, &selected)
    if err != nil {
        log.Fatal(err)
    }
}

// Use the response
defer resp.Body.Close()
body, _ := io.ReadAll(resp.Body)
```

### Resource Middleware (`resource/middleware`)

Gin middleware for protecting API routes with x402 payment verification.

```go
import (
    "github.com/vorpalengineering/x402-go/resource/middleware"
    "github.com/vorpalengineering/x402-go/types"
)

// Configure middleware
x402 := middleware.NewX402Middleware(&middleware.MiddlewareConfig{
    FacilitatorURL: "http://localhost:4020",
    DefaultRequirements: types.PaymentRequirements{
        Scheme:  "exact",
        Network: "eip155:8453",
        Amount:  "1000000",
        PayTo:   "0x123...",
        Asset:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
    },
    ProtectedPaths: []string{"/api/*"},
})

// Apply to your Gin router
router.Use(x402.Handler())
```

The middleware implements the full x402 payment flow:
1. Returns `402` with `PAYMENT-REQUIRED` header when no payment is provided
2. Verifies payment with the facilitator
3. Executes the handler if payment is valid
4. Settles payment on-chain after successful response
5. Returns `PAYMENT-RESPONSE` header with settlement details

### Facilitator (`facilitator`)

Facilitator service for payment verification and on-chain settlement.

```bash
# Set your facilitator's private key
export X402_FACILITATOR_PRIVATE_KEY=0x123abc...

# Copy and edit config
cp facilitator/config.example.yaml facilitator/config.yaml

# Run the facilitator (uses facilitator/config.yaml by default)
go run ./cmd/facilitator
go run ./cmd/facilitator --config=path/to/config.yaml
```

**Endpoints:**
- `GET /supported` - Returns supported scheme/network combinations, extensions, and signer addresses
- `POST /verify` - Verifies a payment payload against requirements
- `POST /settle` - Settles a verified payment on-chain via EIP-3009 `TransferWithAuthorization`

**Configuration (`facilitator/config.yaml`):**
```yaml
server:
  host: "0.0.0.0"
  port: 4020

# Networks use CAIP-2 identifiers
networks:
  eip155:8453:
    rpc_url: "https://mainnet.base.org"
  eip155:84532:
    rpc_url: "https://sepolia.base.org"

supported:
  - scheme: "exact"
    network: "eip155:8453"

transaction:
  timeout_seconds: 120
  max_gas_price: "100000000000"

log:
  level: "info"
```

### Facilitator Client (`facilitator/client`)

Client library for communicating with an x402 facilitator.

```go
import (
    "github.com/vorpalengineering/x402-go/facilitator/client"
    "github.com/vorpalengineering/x402-go/types"
)

fc := client.NewFacilitatorClient("http://localhost:4020")

// Get supported schemes
supported, err := fc.Supported()

// Verify a payment
verifyResp, err := fc.Verify(&types.VerifyRequest{
    PaymentPayload: paymentPayload,
    PaymentRequirements: types.PaymentRequirements{
        Scheme:  "exact",
        Network: "eip155:8453",
        Amount:  "1000000",
        PayTo:   "0x123...",
        Asset:   "0x833...",
    },
})
if verifyResp.IsValid {
    // Payment is valid, payer: verifyResp.Payer
}

// Settle a payment
settleResp, err := fc.Settle(&types.SettleRequest{
    PaymentPayload:      paymentPayload,
    PaymentRequirements: requirements,
})
if settleResp.Success {
    fmt.Printf("Settled: tx=%s, network=%s\n", settleResp.Transaction, settleResp.Network)
}
```

## Docker

Both the facilitator and CLI can be run as containers via Docker Compose.

```bash
# Start the facilitator service
export X402_FACILITATOR_PRIVATE_KEY=0x...
docker compose up facilitator

# Run a CLI command
docker compose run --rm x402cli supported -u http://facilitator:4020

# Build images without starting
docker compose build
```

The facilitator mounts `facilitator/config.yaml` into the container. Ensure the file exists before running (copy from `facilitator/config.example.yaml`).

See the [facilitator README](facilitator/README.md#docker) and [CLI README](cmd/x402cli/README.md#docker) for more details.

## v2 Protocol

This implementation follows x402 protocol version 2. Key aspects:

- **CAIP-2 network identifiers** (e.g., `eip155:8453` for Base, `eip155:1` for Ethereum mainnet)
- **EIP-3009 TransferWithAuthorization** for gasless token transfers
- **Transport headers**: `PAYMENT-SIGNATURE` (client request), `PAYMENT-REQUIRED` (402 response), `PAYMENT-RESPONSE` (success response)
- **Payment payload** carries the accepted requirements and signed authorization as a base64-encoded JSON object

See the [x402 specification](https://github.com/coinbase/x402) for full protocol details.
