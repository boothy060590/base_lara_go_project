package v1

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/http/controllers"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterAPIRoutes registers all API v1 routes
func RegisterAPIRoutes(router *gin.Engine, container *app_core.Container) {
	// Create API v1 group
	api := router.Group("/api/v1")
	
	// Health check endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
			"service":   "laravel-go-framework",
			"server":    "fasthttp",
		})
	})

	// Initialize controllers
	authController := controllers.NewAuthController(container)

	// Auth routes
	auth := api.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.GET("/profile", authController.GetProfile) // TODO: Add JWT middleware
	}
}