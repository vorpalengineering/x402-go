package middleware

import (
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
			ctx.JSON(http.StatusPaymentRequired, response)
			ctx.Abort()
			return
		}

		// Payment is valid, store payment info in context for downstream handlers
		ctx.Set("x402_payment_verified", true)
		ctx.Set("x402_payment_header", paymentHeader)
		ctx.Set("x402_payment_requirements", requirements)

		// Replace response writer with buffered version to capture response
		buffered := newBufferedWriter(ctx.Writer)
		ctx.Writer = buffered

		// STEP 2: Fulfill request (handler executes)
		ctx.Next()

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
	response := types.PaymentRequired{
		X402Version: 2,
		Resource: &types.ResourceInfo{
			URL:         path,
			Description: "",
			MimeType:    "",
		},
		Accepts: []types.PaymentRequirements{requirements},
	}
	ctx.JSON(http.StatusPaymentRequired, response)
	ctx.Abort()
}
