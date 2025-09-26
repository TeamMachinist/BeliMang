package items

import (
	"github.com/gin-gonic/gin"
)

func ItemRoutes(router *gin.RouterGroup, handler *ItemHandler) {
	items := router.Group("/merchant")
	{
		items.POST("/:merchantId/items", handler.CreateItem)
	}
}
