package merchant

import (
	"github.com/gin-gonic/gin"
)

func MerchantRoutes(router *gin.RouterGroup, handler *MerchantHandler) {
	merchants := router.Group("/admin/merchants")
	{
		merchants.POST("", handler.CreateMerchantHandler)
		// merchants.GET("", handler.Login)
	}
}
