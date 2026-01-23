# x402 Middleware for Gin

Gin middleware for x402 (v2) payment verification. This package adds payment-gated access to any Gin-based API.

## Features

- Drop-in Gin middleware for payment verification and settlement
- Complete payment flow: verify, fulfill, settle, respond
- Buffered response ensures payment before access
- Flexible path protection using glob patterns
- Route-specific payment requirements
- v2 transport headers (`PAYMENT-REQUIRED`, `PAYMENT-RESPONSE`)
- Integration with x402 facilitator services

## Installation

```bash
go get github.com/vorpalengineering/x402-go/resource/middleware
```

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/vorpalengineering/x402-go/resource/middleware"
    "github.com/vorpalengineering/x402-go/types"
)

func main() {
    router := gin.Default()

    // Configure x402 middleware
    x402 := middleware.NewX402Middleware(&middleware.MiddlewareConfig{
        FacilitatorURL: "http://localhost:4020",
        DefaultRequirements: types.PaymentRequirements{
            Scheme:  "exact",
            Network: "eip155:8453",
            Amount:  "1000000", // 1 USDC (6 decimals)
            PayTo:   "0x123...",
            Asset:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC on Base
        },
        ProtectedPaths: []string{"/api/*"},
    })

    // Apply middleware globally
    router.Use(x402.Handler())

    // Your routes
    router.GET("/api/data", func(c *gin.Context) {
        c.JSON(200, gin.H{"data": "protected content"})
    })

    router.Run(":3000")
}
```

## Configuration

### MiddlewareConfig Structure

```go
type MiddlewareConfig struct {
    // FacilitatorURL is the base URL of the x402 facilitator service
    FacilitatorURL string

    // DefaultRequirements specifies default payment requirements
    DefaultRequirements types.PaymentRequirements

    // ProtectedPaths is a list of path patterns requiring payment
    // Supports glob patterns like "/api/*" or exact paths like "/data"
    ProtectedPaths []string

    // RouteRequirements maps specific routes to custom payment requirements
    RouteRequirements map[string]types.PaymentRequirements

    // RouteResources maps specific routes to ResourceInfo metadata
    RouteResources map[string]*types.ResourceInfo

    // PaymentHeaderName is the HTTP header containing payment
    // Defaults to "PAYMENT-SIGNATURE"
    PaymentHeaderName string
}
```

### Path Protection

Protect specific paths using glob patterns:

```go
ProtectedPaths: []string{
    "/api/*",           // All routes under /api/
    "/v1/premium/*",    // All routes under /v1/premium/
    "/data/sensitive",  // Exact path
}
```

### Route-Specific Requirements

Override default requirements for specific routes:

```go
RouteRequirements: map[string]types.PaymentRequirements{
    "/api/premium": {
        Scheme:  "exact",
        Network: "eip155:8453",
        Amount:  "5000000", // 5 USDC for premium
        PayTo:   "0x123...",
        Asset:   "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
    },
}
```

## Usage Patterns

### Global Middleware

Apply to all routes:

```go
router.Use(x402.Handler())
```

### Route-Specific Middleware

Apply only to specific routes:

```go
router.GET("/api/data", x402.Handler(), yourHandler)
router.POST("/api/submit", x402.Handler(), anotherHandler)
```

### Group Middleware

Apply to route groups:

```go
apiGroup := router.Group("/api")
apiGroup.Use(x402.Handler())
{
    apiGroup.GET("/data", getData)
    apiGroup.POST("/submit", submitData)
}
```

## Payment Flow

The middleware implements the full x402 v2 payment flow:

1. **Request without payment** - Returns 402 with payment requirements and `PAYMENT-REQUIRED` header
2. **Request with `PAYMENT-SIGNATURE` header** - Verifies payment with facilitator
3. **Invalid payment** - Returns 402 with error details and `PAYMENT-REQUIRED` header
4. **Valid payment** - Handler executes (response is buffered)
5. **Handler succeeds (2xx)** - Settles payment on-chain via facilitator
6. **Settlement succeeds** - Sends buffered response with `PAYMENT-RESPONSE` header
7. **Settlement fails** - Returns error (buffered response is discarded)

The response is only sent to the client AFTER successful payment settlement.

## Transport Headers

| Header | Direction | Description |
|--------|-----------|-------------|
| `PAYMENT-SIGNATURE` | Client -> Server | Base64-encoded payment payload |
| `PAYMENT-REQUIRED` | Server -> Client | Base64-encoded payment requirements (on 402) |
| `PAYMENT-RESPONSE` | Server -> Client | Base64-encoded settlement response (on success) |

## Response Formats

### 402 Payment Required (No Payment)

```json
{
  "x402Version": 2,
  "resource": {
    "url": "/api/data"
  },
  "accepts": [
    {
      "scheme": "exact",
      "network": "eip155:8453",
      "amount": "1000000",
      "payTo": "0x123...",
      "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913",
      "maxTimeoutSeconds": 120
    }
  ]
}
```

### 402 Payment Required (Invalid Payment)

```json
{
  "x402Version": 2,
  "accepts": [...],
  "error": "insufficient amount: got 500000, required 1000000"
}
```

## Context Values

### During Handler Execution (After Verification)

```go
verified, _ := c.Get("x402_payment_verified")       // bool
paymentHeader, _ := c.Get("x402_payment_header")    // string
requirements, _ := c.Get("x402_payment_requirements") // types.PaymentRequirements
```

### After Settlement

```go
txHash, _ := c.Get("x402_settlement_tx")        // string
network, _ := c.Get("x402_settlement_network")  // string (CAIP-2)
payer, _ := c.Get("x402_settlement_payer")       // string
```

Settlement context values are set after the handler completes but before the response is sent.

## Error Handling

| Scenario | Response |
|----------|----------|
| No payment header | 402 with requirements |
| Invalid/malformed header | 400 Bad Request |
| Payment verification fails | 402 with error |
| Facilitator unreachable | 502 Bad Gateway |
| Settlement fails | 502 or 402 (response discarded) |
| Settlement succeeds | Original handler response sent |

If settlement fails, the buffered response is discarded — no access is granted without payment.

## See Also

- [x402 Specification](https://github.com/coinbase/x402)
- [x402-go Facilitator](../../facilitator)
- [x402-go Facilitator Client](../../facilitator/client)
- [x402-go Resource Client](../client)
