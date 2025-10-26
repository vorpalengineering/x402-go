package middleware

import (
	"bytes"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/client"
	"github.com/vorpalengineering/x402-go/types"
)

// bufferedWriter captures the response so we can settle payment before sending to client
type bufferedWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
	header http.Header
}

func newBufferedWriter(w gin.ResponseWriter) *bufferedWriter {
	return &bufferedWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		status:         200,
		header:         make(http.Header),
	}
}

func (w *bufferedWriter) Write(data []byte) (int, error) {
	return w.body.Write(data)
}

func (w *bufferedWriter) WriteHeader(status int) {
	w.status = status
}

func (w *bufferedWriter) Header() http.Header {
	return w.header
}

func (w *bufferedWriter) Status() int {
	return w.status
}

func (w *bufferedWriter) flush() error {
	// Copy buffered headers to real response
	for k, v := range w.header {
		for _, val := range v {
			w.ResponseWriter.Header().Add(k, val)
		}
	}
	// Write status and body
	w.ResponseWriter.WriteHeader(w.status)
	_, err := w.ResponseWriter.Write(w.body.Bytes())
	return err
}

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
			log.Println("Not a protected path. Ignoring...")
			ctx.Next()
			return
		}

		log.Println("Handling x402 request...")

		// Extract payment header
		headerName := m.config.GetPaymentHeaderName()
		paymentHeader := ctx.GetHeader(headerName)

		// If no payment header is present, return 402 Payment Required
		if paymentHeader == "" {
			log.Println("returned PaymentRequired response...")
			m.sendPaymentRequired(ctx, ctx.Request.URL.Path)
			return
		}

		// Get payment requirements for this route
		requirements := m.getRequirements(ctx.Request.URL.Path)

		// Verify payment with facilitator
		verifyReq := &types.VerifyRequest{
			X402Version:         1,
			PaymentHeader:       paymentHeader,
			PaymentRequirements: requirements,
		}

		log.Println("Verifying payment...")

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
			response := types.PaymentRequiredResponse{
				X402Version: 1,
				Accepts:     []types.PaymentRequirements{requirements},
				Error:       verifyResp.InvalidReason,
			}
			ctx.JSON(http.StatusPaymentRequired, response)
			ctx.Abort()
			return
		}

		log.Println("Payment verified...")

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
			log.Println("API Request complete. Settling payment...")

			settleReq := &types.SettleRequest{
				X402Version:         1,
				PaymentHeader:       paymentHeader,
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

			log.Println("Payment Settled")

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

	// Use default requirements with resource set to the path
	requirements := m.config.DefaultRequirements
	if requirements.Resource == "" {
		requirements.Resource = path
	}
	return requirements
}

func (m *X402Middleware) sendPaymentRequired(ctx *gin.Context, path string) {
	requirements := m.getRequirements(path)
	response := types.PaymentRequiredResponse{
		X402Version: 1,
		Accepts:     []types.PaymentRequirements{requirements},
	}
	ctx.JSON(http.StatusPaymentRequired, response)
	ctx.Abort()
}
