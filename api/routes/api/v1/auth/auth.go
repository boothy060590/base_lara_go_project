package auth

import (
	"base_lara_go_project/app/http/controllers"
	"base_lara_go_project/app/http/middlewares"
	"base_lara_go_project/app/providers"

	"github.com/gin-gonic/gin"
)

func Routes(router *gin.Engine) {
	public := router.Group("/v1/auth")

	public.POST("/register", controllers.Register)
	public.POST("/login", controllers.Login)
	public.Use(middlewares.JwtAuthMiddleware()).GET("/user", controllers.CurrentUser)
}

func init() {
	providers.RegisterRouteGroup(Routes)
}
