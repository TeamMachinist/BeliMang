package user

import (
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.RouterGroup, handler *UserHandler) {
	users := router.Group("/users")
	{
		users.POST("/register", handler.Register)
		users.POST("/login", handler.Login)
	}
}
