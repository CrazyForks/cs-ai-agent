package middleware

import (
	"cs-agent/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/web"
)

func AuthMiddleware(ctx *gin.Context) {
	if !authenticateRequest(ctx) {
		return
	}
	ctx.Next()
}

func authenticateRequest(ctx *gin.Context) bool {
	if _, err := services.AuthService.Authenticate(ctx); err != nil {
		ctx.JSON(200, web.JsonError(err))
		ctx.Abort()
		return false
	}
	return true
}
