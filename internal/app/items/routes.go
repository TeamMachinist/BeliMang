package items

import (
	"github.com/gin-gonic/gin"
)

func ItemRoutes(router *gin.Engine, handler *ItemHandler) {
	items := router.Group("/merchant")
	{
		items.POST("/:merchantId/items", handler.CreateItem)
		items.GET("/:merchantId/items", handler.GetItems)
	}
}
