package facilitator

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/types"
)

var cfg *FacilitatorConfig

func RegisterRoutes(router *gin.Engine, config *FacilitatorConfig) {
	cfg = config
	router.POST("/verify", handleVerify)
	router.POST("/settle", handleSettle)
	router.GET("/supported", handleSupported)
}

func handleVerify(ctx *gin.Context) {
	// Decode request
	var req types.VerifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// TODO: Check if scheme-network combination is supported
	// if !cfg.IsSupported(req.PaymentRequirements.Scheme, req.PaymentRequirements.Network) {
	// 	res := types.VerifyResponse{
	// 		IsValid:       false,
	// 		InvalidReason: fmt.Sprintf("unsupported scheme-network: %s-%s", req.PaymentRequirements.Scheme, req.PaymentRequirements.Network),
	// 	}
	// 	ctx.JSON(http.StatusOK, res)
	// 	return
	// }

	// Verify request
	isValid, invalidReason := VerifyPayment(&req)

	// Craft response
	res := types.VerifyResponse{
		IsValid:       isValid,
		InvalidReason: invalidReason,
	}

	ctx.JSON(http.StatusOK, res)
}

func handleSettle(ctx *gin.Context) {
	// Decode request
	var req types.SettleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	//TODO: settle request

	res := types.SettleResponse{
		Success:     true,
		Transaction: "0x123",
		Network:     "41",
		Payer:       "0xabc",
	}

	ctx.JSON(http.StatusOK, res)
}

func handleSupported(ctx *gin.Context) {
	res := types.SupportedResponse{
		Kinds: cfg.Supported,
	}

	ctx.JSON(http.StatusOK, res)
}
