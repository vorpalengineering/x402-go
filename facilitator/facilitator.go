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

	// TODO: verify request

	// Craft response
	res := types.VerifyResponse{
		IsValid: true,
		// InvalidReason: "Payment verification stubbed - not yet implemented",
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
