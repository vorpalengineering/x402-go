# X402 Middleware for Gin

Standalone Gin middleware for x402 payment verification. This package allows you to add payment-gated access to any Gin-based API.

## Features

- Drop-in Gin middleware for payment verification and settlement
- Complete payment flow: verify → fulfill → settle → respond
- Buffered response ensures payment before access
- Flexible path protection using glob patterns
- Route-specific payment requirements
- Automatic 402 Payment Required responses
- Integration with x402 facilitator services

## Installation

```bash
go get github.com/vorpalengineering/x402-go/middleware
```

## Quick Start

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/vorpalengineering/x402-go/middleware"
    "github.com/vorpalengineering/x402-go/types"
)

func main() {
    router := gin.Default()

    // Configure x402 middleware
    x402 := middleware.NewX402Middleware(&middleware.Config{
        FacilitatorURL: "http://localhost:8080",
        DefaultRequirements: types.PaymentRequirements{
            Scheme:            "exact",
            Network:           "base",
            MaxAmountRequired: "1000000", // 1 USDC (6 decimals)
            PayTo:             "0x123...", // Your seller address here
            Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC on Base
            Description:       "API access",
            MaxTimeoutSeconds: 120,
        },
        ProtectedPaths: []string{"/api/*"},
    })

    // Apply middleware globally
    router.Use(x402.Handler())

    // Your routes
    router.GET("/api/data", func(c *gin.Context) {
        c.JSON(200, gin.H{"data": "protected content"})
    })

    router.Run(":8080")
}
```

## Configuration

### Config Structure

```go
type Config struct {
    // FacilitatorURL is the base URL of the x402 facilitator service
    FacilitatorURL string

    // DefaultRequirements specifies default payment requirements
    DefaultRequirements types.PaymentRequirements

    // ProtectedPaths is a list of path patterns requiring payment
    // Supports glob patterns like "/api/*" or exact paths like "/data"
    ProtectedPaths []string

    // RouteRequirements maps specific routes to custom payment requirements
    RouteRequirements map[string]types.PaymentRequirements

    // PaymentHeaderName is the HTTP header containing payment
    // Defaults to "X-Payment"
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
        Scheme:            "exact",
        Network:           "base",
        MaxAmountRequired: "5000000", // 5 USDC for premium
        PayTo:             "0x123...", // Your seller address here
        Asset:             "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913", // USDC on Base
        Description:       "Premium API access",
        MaxTimeoutSeconds: 120,
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

The middleware implements a complete payment flow with verification and settlement:

1. **Request without payment** → Returns 402 with payment requirements
2. **Request with payment** → Verifies payment with facilitator
3. **Invalid payment** → Returns 402 with error details
4. **Valid payment** → Handler executes (response is buffered)
5. **Handler succeeds (2xx)** → Settles payment on-chain via facilitator
6. **Settlement succeeds** → Sends buffered response to client
7. **Settlement fails** → Returns error (response is not sent)

**Important:** The response is only sent to the client AFTER successful payment settlement. This ensures payment is collected before granting access to protected resources.

## Response Formats

### 402 Payment Required (No Payment)

```json
{
  "x402Version": 1,
  "accepts": [
    {
      "scheme": "exact",
      "network": "base",
      "maxAmountRequired": "1000000",
      "resource": "/api/data",
      "description": "API access payment",
      "mimeType": "application/json",
      "payTo": "0x123...",
      "maxTimeoutSeconds": 120,
      "asset": "0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913"
    }
  ]
}
```

### 402 Payment Required (Invalid Payment)

```json
{
  "x402Version": 1,
  "accepts": [...],
  "error": "insufficient amount: got 500000, required 1000000"
}
```

## Context Values

The middleware sets these values in the Gin context:

### During Handler Execution (After Verification)

```go
// Check if payment was verified
verified, _ := c.Get("x402_payment_verified")  // bool

// Get the payment header
paymentHeader, _ := c.Get("x402_payment_header")  // string

// Get payment requirements
requirements, _ := c.Get("x402_payment_requirements")  // types.PaymentRequirements
```

### After Settlement (Available in Response)

```go
// Settlement transaction hash
txHash, _ := c.Get("x402_settlement_tx")  // string

// Network where payment was settled
network, _ := c.Get("x402_settlement_network")  // string

// Payer address
payer, _ := c.Get("x402_settlement_payer")  // string
```

**Note:** Settlement context values are set AFTER the handler completes but BEFORE the response is sent to the client.

## Error Handling

The middleware handles errors automatically:

- **No payment header** → 402 with payment requirements
- **Invalid payment** → 402 with error reason
- **Verification failure** → 502 Bad Gateway
- **Valid payment** → Handler executes (buffered)
- **Settlement failure** → 502 Bad Gateway or 402 (buffered response is discarded)
- **Settlement success** → Response sent to client

**Critical:** If settlement fails, the buffered response from the handler is NOT sent to the client, ensuring no access is granted without payment.

## See Also

- [x402 Specification](https://github.com/coinbase/x402)
- [x402-go Client Library](../client)
- [x402-go Facilitator](../facilitator)
- [x402 Reverse Proxy](../proxy)
