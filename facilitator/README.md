# x402 Facilitator

The facilitator service provides payment verification and on-chain settlement for the x402 protocol (v2). It validates EIP-3009 payment authorizations and executes `TransferWithAuthorization` transactions.

## Prerequisites

- Go 1.21 or later
- Access to blockchain RPC endpoints
- A funded Ethereum wallet (private key) for executing settlements

## Quick Start

### 1. Set Environment Variables

The facilitator requires a private key for signing and executing on-chain transactions:

```bash
export X402_FACILITATOR_PRIVATE_KEY=0x1234567890abcdef...
```

**Security Note**: Never commit your private key to version control. Use environment variables or a secure secret manager.

### 2. Configure the Facilitator

Copy the example configuration:

```bash
cp facilitator/config.example.yaml facilitator/config.yaml
```

Edit `facilitator/config.yaml` with your settings. See [Configuration](#configuration) below.

### 3. Run the Service

```bash
# Uses facilitator/config.yaml by default
go run ./cmd/facilitator

# Or specify a custom config path
go run ./cmd/facilitator --config=path/to/config.yaml
```

The service starts on the configured port (default: 4020).

## Configuration

The facilitator uses a YAML config file:

```yaml
server:
  host: "0.0.0.0"
  port: 4020

# Networks use CAIP-2 identifiers (namespace:reference)
networks:
  eip155:8453:
    rpc_url: "https://mainnet.base.org"
  eip155:1:
    rpc_url: "https://eth.llamarpc.com"

# Supported scheme-network combinations
supported:
  - scheme: "exact"
    network: "eip155:8453"
  - scheme: "exact"
    network: "eip155:1"

transaction:
  timeout_seconds: 120
  max_gas_price: "100000000000"  # 100 gwei in wei

log:
  level: "info"  # debug, info, warn, error
```

### Networks

Each network requires a CAIP-2 identifier as the key and an RPC URL:

| Network | CAIP-2 ID |
|---------|-----------|
| Ethereum mainnet | `eip155:1` |
| Base | `eip155:8453` |
| Base Sepolia | `eip155:84532` |
| Optimism | `eip155:10` |

### Supported Schemes

Only requests matching a configured scheme-network pair are processed. Currently supported schemes:
- `exact` — Fixed-amount EIP-3009 TransferWithAuthorization

## API Endpoints

### `GET /supported`

Returns supported configurations, extensions, and signer addresses.

**Response:**
```json
{
  "kinds": [
    {"x402Version": 2, "scheme": "exact", "network": "eip155:8453"},
    {"x402Version": 2, "scheme": "exact", "network": "eip155:1"}
  ],
  "extensions": [],
  "signers": {
    "eip155:8453": ["0xYourSignerAddress"],
    "eip155:1": ["0xYourSignerAddress"]
  }
}
```

### `POST /verify`

Verifies a payment payload against requirements.

**Request:**
```json
{
  "paymentPayload": {
    "x402Version": 2,
    "accepted": {
      "scheme": "exact",
      "network": "eip155:8453",
      "amount": "1000000",
      "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
      "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    },
    "payload": {
      "signature": "0x...",
      "authorization": {
        "from": "0x...",
        "to": "0x...",
        "value": "1000000",
        "validAfter": 1700000000,
        "validBefore": 1700003600,
        "nonce": "0x..."
      }
    }
  },
  "paymentRequirements": {
    "scheme": "exact",
    "network": "eip155:8453",
    "amount": "1000000",
    "payTo": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
  }
}
```

**Response:**
```json
{
  "isValid": true,
  "payer": "0xPayerAddress"
}
```

### `POST /settle`

Executes the payment on-chain via `TransferWithAuthorization`.

**Request:** Same structure as `/verify`.

**Response:**
```json
{
  "success": true,
  "transaction": "0xTransactionHash",
  "network": "eip155:8453",
  "payer": "0xPayerAddress"
}
```

## Docker

A multi-stage Dockerfile is provided at `cmd/facilitator/Dockerfile`. It produces a minimal Alpine-based image containing only the `facilitator` binary, exposing port 4020.

### Building

```bash
# From the project root
docker build -f cmd/facilitator/Dockerfile -t x402-facilitator .
```

### Running

```bash
# Run with a mounted config and private key
docker run --rm -p 4020:4020 \
  -v $(pwd)/facilitator/config.yaml:/etc/x402/config.yaml:ro \
  -e X402_FACILITATOR_PRIVATE_KEY=0x... \
  x402-facilitator --config /etc/x402/config.yaml
```

### Docker Compose

The facilitator is defined as a service in the project-level `docker-compose.yml`:

```bash
# Set your private key and start the facilitator
export X402_FACILITATOR_PRIVATE_KEY=0x...
docker compose up facilitator

# Rebuild after code changes
docker compose build facilitator
docker compose up facilitator
```

The compose service mounts `facilitator/config.yaml` read-only into the container at `/etc/x402/config.yaml`. Ensure your config file exists before running:

```bash
cp facilitator/config.example.yaml facilitator/config.yaml
# Edit facilitator/config.yaml with your RPC URLs and supported networks
```

## Facilitator Client

Use the client library to communicate with a facilitator from your resource server:

```go
import (
    "github.com/vorpalengineering/x402-go/facilitator/client"
    "github.com/vorpalengineering/x402-go/types"
)

c := client.NewClient("http://localhost:4020")

// Get supported schemes
supported, err := c.Supported()

// Verify a payment
verifyResp, err := c.Verify(&types.VerifyRequest{
    PaymentPayload:      paymentPayload,
    PaymentRequirements: requirements,
})

// Settle a payment
settleResp, err := c.Settle(&types.SettleRequest{
    PaymentPayload:      paymentPayload,
    PaymentRequirements: requirements,
})
```

## See Also

- [x402 Specification](https://github.com/coinbase/x402)
- [Resource Middleware](../resource/middleware) — Gin middleware that uses this facilitator
- [Resource Client](../resource/client) — Client for accessing x402-protected resources
