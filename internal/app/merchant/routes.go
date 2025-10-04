package merchant

import (
	"belimang/internal/pkg/jwt"
	"belimang/internal/middleware"

	"github.com/gin-gonic/gin"
)

func MerchantRoutes(router *gin.Engine, handler *MerchantHandler, jwtService *jwt.JWTService) {
	merchants := router.Group("/admin/merchants")
	merchants.Use(middleware.RequireAdmin(jwtService))
	{
		merchants.POST("", handler.CreateMerchantHandler)
		merchants.GET("", handler.SearchMerchantsHandler)
	}
}
