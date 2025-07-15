package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	v1 "base_lara_go_project/routes/api/v1"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RouterServiceProvider provides routing with FastHTTP server integration
type RouterServiceProvider struct {
	laravel_providers.BaseServiceProvider
	httpOptimizer *app_core.HTTPOptimizer
}

// Register registers the router service provider with FastHTTP integration
func (p *RouterServiceProvider) Register(container *app_core.Container) error {
	// Register router as singleton with FastHTTP support
	container.Singleton("router", func() (any, error) {
		router := gin.Default()

		// Add performance monitoring middleware
		router.Use(p.performanceMiddleware(container))

		return router, nil
	})

	// Register FastHTTP server (no fallback) - avoid circular dependency
	container.Singleton("http.server", func() (any, error) {
		// Get HTTP optimizer
		optimizerInstance, err := container.Resolve("http.optimizer")
		if err != nil {
			return nil, fmt.Errorf("HTTP optimizer not found: %w", err)
		}

		optimizer := optimizerInstance.(*app_core.HTTPOptimizer)
		p.httpOptimizer = optimizer

		// Note: We'll set the router handler during boot, not registration
		// This prevents circular dependency during container resolution

		log.Printf("FastHTTP server registered successfully")
		return optimizer, nil
	})

	log.Printf("Router service provider registered successfully with FastHTTP integration")
	return nil
}

// Boot loads application routes for FastHTTP server
func (p *RouterServiceProvider) Boot(container *app_core.Container) error {
	log.Printf("Starting RouterServiceProvider boot process...")

	// Get router from container
	routerInstance, err := container.Resolve("router")
	if err != nil {
		log.Printf("Router not found in container: %v", err)
		return err
	}

	router := routerInstance.(*gin.Engine)
	log.Printf("Router resolved successfully")

	// Load routes from routes directory
	log.Printf("Loading routes...")
	if err := p.loadRoutes(router, container); err != nil {
		log.Printf("Failed to load routes: %v", err)
		return err
	}
	log.Printf("Routes loaded successfully")

	// Add metrics routes
	log.Printf("Adding metrics routes...")
	if err := p.addMetricsRoutes(router, container); err != nil {
		log.Printf("Failed to add metrics routes: %v", err)
		return err
	}
	log.Printf("Metrics routes added successfully")

	// Set the router as the handler for the HTTP optimizer
	if p.httpOptimizer != nil {
		log.Printf("Setting router as handler for FastHTTP optimizer...")
		p.httpOptimizer.SetHandler(router)
		log.Printf("Router handler set successfully")
	}

	log.Printf("Routes loaded successfully for FastHTTP server")
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
	// Load API v1 routes
	log.Printf("Loading API v1 routes...")
	if err := p.loadAPIV1Routes(router, container); err != nil {
		log.Printf("Failed to load API v1 routes: %v", err)
		return err
	}

	return nil
}

// loadAPIV1Routes loads API v1 routes from the routes package
func (p *RouterServiceProvider) loadAPIV1Routes(router *gin.Engine, container *app_core.Container) error {
	v1.RegisterAPIRoutes(router, container)
	log.Printf("API v1 routes loaded successfully")
	return nil
}

// addMetricsRoutes adds metrics endpoint routes
func (p *RouterServiceProvider) addMetricsRoutes(router *gin.Engine, container *app_core.Container) error {
	// Add basic metrics endpoint
	router.GET("/metrics", func(c *gin.Context) {
		metrics := map[string]any{
			"server":    "fasthttp",
			"service":   "laravel-go-framework",
			"status":    "healthy",
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		}
		c.JSON(200, metrics)
	})

	log.Printf("Metrics routes added successfully")
	return nil
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
