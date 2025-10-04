package purchase

import (
	"belimang/internal/middleware"
	"belimang/internal/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func PurchaseRoutes(router *gin.Engine, handler *PurchaseHandler, jwtService *jwt.JWTService) {
	nearby := router.Group("/merchants")
	{
		nearby.GET("/nearby/:coords", middleware.RequireUser(jwtService), handler.GetMerchantsNearbyHandler)
	}
	purchase := router.Group("/users")
	purchase.Use(middleware.RequireUser(jwtService))

	{
		purchase.POST("/estimate", handler.Estimate)
		purchase.POST("/orders", handler.CreateOrder)
	}

}
