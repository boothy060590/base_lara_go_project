package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// RouterServiceProvider provides optimized routing with HTTP performance enhancements
type RouterServiceProvider struct {
	laravel_providers.BaseServiceProvider
	httpOptimizer *app_core.HTTPOptimizer
}

// Register registers the router service provider with HTTP optimization
func (p *RouterServiceProvider) Register(container *app_core.Container) error {
	// Register router as singleton with HTTP optimization support
	container.Singleton("router", func() (any, error) {
		router := gin.Default()

		// Add CORS middleware
		router.Use(p.corsMiddleware())

		// Add performance monitoring middleware
		router.Use(p.performanceMiddleware(container))

		return router, nil
	})

	// Register optimized HTTP server
	container.Singleton("http.server", func() (any, error) {
		// Get HTTP optimizer
		optimizerInstance, err := container.Resolve("http.optimizer")
		if err != nil {
			log.Printf("HTTP optimizer not found, falling back to standard server: %v", err)
			return nil, err
		}

		optimizer := optimizerInstance.(*app_core.HTTPOptimizer)
		p.httpOptimizer = optimizer

		// Get router
		routerInstance, err := container.Resolve("router")
		if err != nil {
			return nil, err
		}

		router := routerInstance.(*gin.Engine)

		// Set the router as the handler for the optimizer
		optimizer.SetHandler(router)

		log.Printf("Optimized HTTP server registered with fasthttp support")
		return optimizer, nil
	})

	log.Printf("Router service provider registered successfully with HTTP optimization")
	return nil
}

// Boot loads application routes with optimization
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

	// Add HTTP optimization routes
	if err := p.addOptimizationRoutes(router, container); err != nil {
		return err
	}

	log.Printf("Routes loaded successfully with HTTP optimization")
	return nil
}

// Provides returns the services this provider provides
func (p *RouterServiceProvider) Provides() []string {
	return []string{"router", "http.server"}
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
		// Health check endpoint
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"timestamp": gin.H{"now": "2024-01-01T00:00:00Z"},
				"service":   "laravel-go-framework",
			})
		})

		// Performance test endpoint
		api.GET("/performance", func(c *gin.Context) {
			if p.httpOptimizer != nil {
				metrics := p.httpOptimizer.GetMetrics()
				c.JSON(http.StatusOK, gin.H{
					"performance": metrics,
				})
			} else {
				c.JSON(http.StatusOK, gin.H{
					"performance": "HTTP optimizer not available",
				})
			}
		})

		// Auth routes placeholder
		auth := api.Group("/auth")
		{
			auth.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Register endpoint - TODO: implement"})
			})
			auth.POST("/login", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Login endpoint - TODO: implement"})
			})
			auth.GET("/profile", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Profile endpoint - TODO: implement"})
			})
		}
	}

	return nil
}

// addOptimizationRoutes adds HTTP optimization specific routes
func (p *RouterServiceProvider) addOptimizationRoutes(router *gin.Engine, container *app_core.Container) error {
	// Get HTTP optimization config
	configInstance, err := container.Resolve("http.optimization.config")
	if err != nil {
		return nil // Skip if not available
	}

	config := configInstance.(*app_core.HTTPOptimizationConfig)

	// Add metrics endpoint if enabled
	if config.EnableMetrics && config.MetricsPath != "" {
		router.GET(config.MetricsPath, func(c *gin.Context) {
			if p.httpOptimizer != nil {
				metrics := p.httpOptimizer.GetMetrics()
				c.JSON(http.StatusOK, metrics)
			} else {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "HTTP optimizer not available",
				})
			}
		})
		log.Printf("HTTP metrics endpoint available at: %s", config.MetricsPath)
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

// performanceMiddleware adds performance monitoring
func (p *RouterServiceProvider) performanceMiddleware(container *app_core.Container) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// Custom log format with performance metrics
		return fmt.Sprintf("[HTTP] %v | %3d | %13v | %15s | %-7s %#v\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	})
}