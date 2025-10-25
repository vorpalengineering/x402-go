package facilitator

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/types"
)

func RegisterRoutes(router *gin.Engine) {
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
	// TODO: implement actual supported SchemeNetwork pairs
	res := types.SupportedResponse{
		Kinds: []types.SchemeNetworkPair{},
	}

	ctx.JSON(http.StatusOK, res)
}
