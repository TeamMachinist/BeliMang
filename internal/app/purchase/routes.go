package purchase

import (
	"github.com/gin-gonic/gin"
)

func PurchaseRoutes(router *gin.Engine, handler *PurchaseHandler) {
	purchase := router.Group("/users")
	{
		purchase.POST("/estimate", handler.Estimate)
	}
}
