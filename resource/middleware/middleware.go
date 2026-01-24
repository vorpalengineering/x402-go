package middleware

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/facilitator/client"
	"github.com/vorpalengineering/x402-go/types"
	"github.com/vorpalengineering/x402-go/utils"
)

type X402Middleware struct {
	config      *MiddlewareConfig
	facilitator *client.Client
}

func NewX402Middleware(cfg *MiddlewareConfig) *X402Middleware {
	return &X402Middleware{
		config:      cfg,
		facilitator: client.NewClient(cfg.FacilitatorURL),
	}
}

func (m *X402Middleware) Handler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Serve discovery endpoint if enabled
		if m.config.DiscoveryEnabled && ctx.Request.URL.Path == "/.well-known/x402" {
			m.serveDiscovery(ctx)
			return
		}

		// Check if the current path requires payment
		if !m.isProtectedPath(ctx.Request.URL.Path) {
			ctx.Next()
			return
		}

		// Extract payment header
		headerName := m.config.GetPaymentHeaderName()
		paymentHeader := ctx.GetHeader(headerName)

		// If no payment header is present, return 402 Payment Required
		if paymentHeader == "" {
			m.sendPaymentRequired(ctx, ctx.Request.URL.Path)
			return
		}

		// Decode payment header into PaymentPayload
		paymentPayload, err := utils.DecodePaymentHeader(paymentHeader)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid payment header: " + err.Error(),
			})
			ctx.Abort()
			return
		}

		// Get payment requirements for this route
		requirements := m.getRequirements(ctx.Request.URL.Path)

		// Verify payment with facilitator
		verifyReq := &types.VerifyRequest{
			PaymentPayload:      *paymentPayload,
			PaymentRequirements: requirements,
		}

		verifyResp, err := m.facilitator.Verify(verifyReq)
		if err != nil {
			// Facilitator communication error
			ctx.JSON(http.StatusBadGateway, gin.H{
				"error": "Failed to verify payment: " + err.Error(),
			})
			ctx.Abort()
			return
		}

		// Check if payment is valid
		if !verifyResp.IsValid {
			// Payment is invalid, return 402 with reason
			response := types.PaymentRequired{
				X402Version: 2,
				Accepts:     []types.PaymentRequirements{requirements},
				Error:       verifyResp.InvalidReason,
			}
			setPaymentRequiredHeader(ctx, &response)
			ctx.JSON(http.StatusPaymentRequired, response)
			ctx.Abort()
			return
		}

		// Payment is valid, store payment info in context for downstream handlers
		ctx.Set("x402_payment_verified", true)
		ctx.Set("x402_payment_header", paymentHeader)
		ctx.Set("x402_payment_requirements", requirements)

		// Replace response writer with buffered version to capture response
		buffered := newBufferedWriter(ctx.Writer, m.config.MaxBufferSize)
		ctx.Writer = buffered

		// STEP 2: Fulfill request (handler executes)
		ctx.Next()

		// Check for buffer overflow
		if buffered.overflow {
			log.Printf("Response exceeded max buffer size (%d bytes), aborting", m.config.MaxBufferSize)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "Response too large to process payment",
			})
			ctx.Abort()
			return
		}

		// STEP 3: Settle payment if handler succeeded (2xx status)
		if buffered.Status() >= 200 && buffered.Status() < 300 {
			settleReq := &types.SettleRequest{
				PaymentPayload:      *paymentPayload,
				PaymentRequirements: requirements,
			}

			settleResp, err := m.facilitator.Settle(settleReq)
			if err != nil {
				// Settlement failed, don't send the buffered response
				ctx.JSON(http.StatusBadGateway, gin.H{
					"error": "Failed to settle payment: " + err.Error(),
				})
				ctx.Abort()
				return
			}

			if !settleResp.Success {
				// Settlement unsuccessful
				ctx.JSON(http.StatusPaymentRequired, gin.H{
					"error": "Payment settlement failed: " + settleResp.ErrorReason,
				})
				ctx.Abort()
				return
			}

			// Store settlement info in context
			ctx.Set("x402_settlement_tx", settleResp.Transaction)
			ctx.Set("x402_settlement_network", settleResp.Network)
			ctx.Set("x402_settlement_payer", settleResp.Payer)

			// Set PAYMENT-RESPONSE header with settlement details
			setPaymentResponseHeader(ctx, settleResp)

			log.Printf("Payment settled: tx=%s, network=%s, payer=%s",
				settleResp.Transaction, settleResp.Network, settleResp.Payer)
		}

		// STEP 4: Send response to client (only after successful settlement)
		buffered.flush()
	}
}

func (m *X402Middleware) isProtectedPath(path string) bool {
	for _, pattern := range m.config.ProtectedPaths {
		matched, err := filepath.Match(pattern, path)
		if err != nil {
			// Invalid pattern, skip
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

func (m *X402Middleware) getRequirements(path string) types.PaymentRequirements {
	// Check for exact route match first
	if req, exists := m.config.RouteRequirements[path]; exists {
		return req
	}

	// Check for pattern matches in route requirements
	for pattern, req := range m.config.RouteRequirements {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return req
		}
	}

	return m.config.DefaultRequirements
}

func (m *X402Middleware) sendPaymentRequired(ctx *gin.Context, path string) {
	requirements := m.getRequirements(path)
	headerName := m.config.GetPaymentHeaderName()

	resource := &types.ResourceInfo{
		URL: path,
	}
	if r, exists := m.config.RouteResources[path]; exists {
		resource.Description = r.Description
		resource.MimeType = r.MimeType
	}

	response := types.PaymentRequired{
		X402Version: 2,
		Error:       headerName + " header is required",
		Resource:    resource,
		Accepts:     []types.PaymentRequirements{requirements},
	}
	setPaymentRequiredHeader(ctx, &response)
	ctx.JSON(http.StatusPaymentRequired, response)
	ctx.Abort()
}

// setPaymentRequiredHeader encodes the PaymentRequired response as base64 JSON
// and sets it as the PAYMENT-REQUIRED response header.
func setPaymentRequiredHeader(ctx *gin.Context, response *types.PaymentRequired) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to encode PAYMENT-REQUIRED header: %v", err)
		return
	}
	ctx.Header("PAYMENT-REQUIRED", base64.StdEncoding.EncodeToString(data))
}

// setPaymentResponseHeader encodes the SettleResponse as base64 JSON
// and sets it as the PAYMENT-RESPONSE response header.
func setPaymentResponseHeader(ctx *gin.Context, response *types.SettleResponse) {
	data, err := json.Marshal(response)
	if err != nil {
		log.Printf("Failed to encode PAYMENT-RESPONSE header: %v", err)
		return
	}
	ctx.Header("PAYMENT-RESPONSE", base64.StdEncoding.EncodeToString(data))
}

func (m *X402Middleware) serveDiscovery(ctx *gin.Context) {
	discovery := gin.H{
		"version":   1,
		"resources": m.config.ProtectedPaths,
	}
	if len(m.config.OwnershipProofs) > 0 {
		discovery["ownershipProofs"] = m.config.OwnershipProofs
	}
	if m.config.Instructions != "" {
		discovery["instructions"] = m.config.Instructions
	}
	ctx.JSON(http.StatusOK, discovery)
	ctx.Abort()
}
