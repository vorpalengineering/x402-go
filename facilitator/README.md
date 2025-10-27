# x402 Facilitator

The facilitator service provides payment verification and settlement for the x402 protocol. It acts as a trusted intermediary that validates payment proofs and executes on-chain settlements.

## Prerequisites

- Go 1.21 or later
- Access to blockchain RPC endpoints (e.g., Base, Ethereum)
- A funded Ethereum wallet (private key) for executing settlements

## Quick Start

### 1. Set Environment Variables

The facilitator requires a private key for signing and executing on-chain transactions:

```bash
export X402_FACILITATOR_PRIVATE_KEY=0x1234567890abcdef...
```

**Security Note**: Never commit your private key to version control. Always use environment variables or secure secret management.

### 2. Configure the Facilitator

Copy the example configuration file:

```bash
cp config.facilitator.example.yaml config.facilitator.yaml
```

Edit `config.facilitator.yaml` with your settings. See [Configuration](#configuration) for details.

### 3. Run the Service

Run with default config (looks for `config.facilitator.yaml` in the current directory):

```bash
go run ./cmd/facilitator
```

Or specify a custom config path:

```bash
go run ./cmd/facilitator --config=path/to/config.facilitator.yaml
```

The service will start on the configured port (default: 8080).

## YAML Configuration

The facilitator uses a YAML configuration file. Here's a detailed explanation of each section:

### Server Configuration

```yaml
server:
  host: "0.0.0.0"  # Listen address (0.0.0.0 for all interfaces)
  port: 8080        # HTTP server port
```

### Network Configuration

Define RPC endpoints for each blockchain network you want to support:

```yaml
networks:
  base:
    rpc_url: "https://mainnet.base.org"
    chain_id: 8453
  ethereum:
    rpc_url: "https://eth.llamarpc.com"
    chain_id: 1
```

You can add additional networks as needed. Each network requires:
- `rpc_url`: HTTP(S) endpoint for the blockchain RPC
- `chain_id`: Network chain ID (e.g., 1 for Ethereum mainnet, 8453 for Base)

### Supported Schemes

List all scheme-network combinations your facilitator will accept:

```yaml
supported:
  - scheme: "exact"
    network: "base"
  - scheme: "exact"
    network: "ethereum"
```

Only requests matching these combinations will be processed. Others will be rejected.

### Transaction Settings

```yaml
transaction:
  timeout_seconds: 120              # Max time to wait for tx confirmation
  max_gas_price: "100000000000"     # Max gas price in wei (100 gwei)
```

- `timeout_seconds`: How long to wait for transaction confirmation before timing out
- `max_gas_price`: Maximum gas price you're willing to pay (prevents excessive fees)

### Logging

```yaml
log:
  level: "info"  # Options: debug, info, warn, error
```

- `debug`: Verbose logging for development
- `info`: Standard operational logging (recommended for production)
- `warn`: Only warnings and errors
- `error`: Only errors

## API Endpoints

The facilitator exposes the following HTTP endpoints:

### `GET /supported`

Returns the list of supported scheme-network combinations.

**Response**:
```json
{
  "supported": [
    {"scheme": "exact", "network": "base"},
    {"scheme": "exact", "network": "ethereum"}
  ]
}
```

### `POST /verify`

Verifies a payment payload against the provided requirements.

**Request**:
```json
{
  "x402_version": 1,
  "payment_header": "base64EncodedPaymentHeader",
  "payment_requirements": {
    "scheme": "exact",
    "network": "base",
    "max_amount_required": "1000000",
    "pay_to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
  }
}
```

**Response**:
```json
{
  "is_valid": true,
  "reason": ""
}
```

### `POST /settle`

Executes the payment on-chain and returns the transaction hash.

**Request**:
```json
{
  "x402_version": 1,
  "payment_header": "base64EncodedPaymentHeader",
  "payment_requirements": {
    "scheme": "exact",
    "network": "base",
    "max_amount_required": "1000000",
    "pay_to": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
  }
}
```

**Response**:
```json
{
  "success": true,
  "tx_hash": "0xabc123...",
  "error": ""
}
```

## Using as a Library

You can also use the facilitator as a library in your Go application:

```go
package main

import (
    "github.com/vorpalengineering/x402-go/facilitator"
    "github.com/vorpalengineering/x402-go/types"
)

func main() {
    // Load configuration
    config, err := facilitator.LoadConfig("config.facilitator.yaml")
    if err != nil {
        panic(err)
    }

    // Create facilitator instance
    f, err := facilitator.NewFacilitator(config)
    if err != nil {
        panic(err)
    }

    // Verify a payment
    verifyReq := &types.VerifyRequest{
        X402Version: 1,
        PaymentHeader: "base64EncodedPaymentHeader",
        PaymentRequirements: types.PaymentRequirements{
            Scheme: "exact",
            Network: "base",
            MaxAmountRequired: "1000000",
            PayTo: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
            Asset: "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
        },
    }

    verifyResp, err := f.Verify(verifyReq)
    if err != nil {
        panic(err)
    }

    if verifyResp.IsValid {
        // Payment is valid, proceed with settlement
        settleResp, err := f.Settle(verifyReq)
        if err != nil {
            panic(err)
        }
        println("Transaction hash:", settleResp.TxHash)
    }
}
```
