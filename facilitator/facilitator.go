package facilitator

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vorpalengineering/x402-go/types"
)

func RegisterRoutes(router *gin.Engine) {
	router.POST("/verify", handleVerify)
	router.POST("/settle", handleSettle)
}

func handleVerify(ctx *gin.Context) {
	var req types.VerifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	res := types.VerifyResponse{
		Valid:   true,
		Message: "Payment verification stubbed - not yet implemented",
	}

	ctx.JSON(http.StatusOK, res)
}

func handleSettle(ctx *gin.Context) {
	var req types.SettleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	res := types.SettleResponse{
		Status:  "pending",
		Message: "Payment settlement stubbed - not yet implemented",
	}

	ctx.JSON(http.StatusOK, res)
}
