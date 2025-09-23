package user

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers user routes with the router
func UserRoutes(router *gin.RouterGroup, handler *UserHandler) {
	users := router.Group("/users")
	{
		users.POST("/", handler.Create)
		users.GET("/", handler.GetAll)
		users.GET("/:id", handler.GetByID)
		users.PUT("/:id", handler.Update)
		users.DELETE("/:id", handler.Delete)
	}
}
