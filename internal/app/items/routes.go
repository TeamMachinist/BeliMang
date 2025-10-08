package items

import (
	"belimang/internal/middleware"
	"belimang/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func ItemRoutes(router *gin.Engine, handler *ItemHandler, jwt *jwt.JWTService) {
	items := router.Group("/admin/merchants")
	items.Use(middleware.RequireAdmin(jwt))
	{
		items.POST("/:merchantId/items", handler.CreateItem)
		items.GET("/:merchantId/items", handler.GetItems)
	}
}
