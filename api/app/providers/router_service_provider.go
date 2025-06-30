package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// RouterServiceProvider loads and registers application routes
type RouterServiceProvider struct {
	laravel_providers.BaseServiceProvider
}

// Register registers the router service provider
func (p *RouterServiceProvider) Register(container *app_core.Container) error {
	// Register router as singleton
	container.Singleton("router", func() (any, error) {
		router := gin.Default()

		// Add CORS middleware
		router.Use(p.corsMiddleware())

		return router, nil
	})

	log.Printf("Router service provider registered successfully")
	return nil
}

// Boot loads application routes
func (p *RouterServiceProvider) Boot(container *app_core.Container) error {
	// Get router from container
	routerInstance, err := container.Resolve("router")
	if err != nil {
		log.Printf("Router not found in container: %v", err)
		return err
	}

	router := routerInstance.(*gin.Engine)

	// Load routes from files
	if err := p.loadRoutes(router, container); err != nil {
		return err
	}

	log.Printf("Routes loaded successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *RouterServiceProvider) Provides() []string {
	return []string{"router"}
}

// When returns the conditions when this provider should be loaded
func (p *RouterServiceProvider) When() []string {
	return []string{}
}

// loadRoutes loads routes from the routes directory
func (p *RouterServiceProvider) loadRoutes(router *gin.Engine, container *app_core.Container) error {
	routesDir := "routes"

	// Check if routes directory exists
	if _, err := os.Stat(routesDir); os.IsNotExist(err) {
		log.Printf("Routes directory not found, creating default routes")
		return p.createDefaultRoutes(router, container)
	}

	// Load route files
	return filepath.Walk(routesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		// Load route file
		log.Printf("Loading routes from: %s", path)
		return p.loadRouteFile(path, router, container)
	})
}

// loadRouteFile loads routes from a specific file
func (p *RouterServiceProvider) loadRouteFile(path string, router *gin.Engine, container *app_core.Container) error {
	// TODO: Implement route file loading
	// This would parse the route file and register routes
	// For now, we'll just log that we're loading the file
	log.Printf("Loading route file: %s", path)
	return nil
}

// createDefaultRoutes creates default routes if no route files exist
func (p *RouterServiceProvider) createDefaultRoutes(router *gin.Engine, container *app_core.Container) error {
	// Create API routes group
	api := router.Group("/api/v1")
	{
		// Auth routes
		_ = api.Group("/auth")
		{
			// TODO: Get controllers from container
			// authController := getAuthController(container)
			// auth.POST("/register", authController.Register)
			// auth.POST("/login", authController.Login)
			// auth.GET("/profile", authController.GetProfile)
		}
	}

	return nil
}

// corsMiddleware returns CORS middleware configuration
func (p *RouterServiceProvider) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
