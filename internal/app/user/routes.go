package user

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.Engine, handler *UserHandler) {
	users := router.Group("/users")
	{
		users.POST("/register", handler.RegisterUser)
		users.POST("/login", handler.LoginUser)
	}

	admin := router.Group("/admin")
	{
		admin.POST("/register", handler.RegisterAdmin)
		admin.POST("/login", handler.LoginAdmin)
	}
}
