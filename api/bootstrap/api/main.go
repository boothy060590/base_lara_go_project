package main

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"base_lara_go_project/app/providers"
	"base_lara_go_project/config"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Initialize the global service container
	container := app_core.NewContainer()

	// Create provider manager
	providerManager := laravel_providers.NewProviderManager(container)

	// Register the main AppServiceProvider which handles all core and app providers
	appProvider := &providers.AppServiceProvider{}
	if err := providerManager.Register(appProvider); err != nil {
		panic(err)
	}

	// Boot all providers
	if err := providerManager.Boot(); err != nil {
		panic(err)
	}

	log.Printf("All providers booted successfully, starting HTTP server...")

	// Create FastHTTP server directly to avoid container resolution issues
	log.Printf("Creating FastHTTP server directly...")
	
	// Get HTTP optimizer directly
	optimizerInstance, err := container.Resolve("http.optimizer")
	if err != nil {
		log.Printf("HTTP optimizer not available: %v", err)
		panic(err)
	}
	httpOptimizer := optimizerInstance.(*app_core.HTTPOptimizer)
	log.Printf("HTTP optimizer resolved successfully")

	// Get router directly
	routerInstance, err := container.Resolve("router")
	if err != nil {
		log.Printf("Router not available: %v", err)
		panic(err)
	}
	router := routerInstance.(*gin.Engine)
	log.Printf("Router resolved successfully")

	// Set router as handler
	log.Printf("Setting router as handler for FastHTTP optimizer...")
	httpOptimizer.SetHandler(router)
	log.Printf("Router handler set successfully")

	// Start server
	appConfig := config.AppConfig()
	port := appConfig["port"].(string)
	
	log.Printf("Starting FastHTTP server on port %s", port)
	log.Printf("Server will be available at: http://localhost:%s", port)
	log.Printf("Health check: http://localhost:%s/api/v1/health", port)
	
	if err := httpOptimizer.ListenAndServe(":" + port); err != nil {
		log.Printf("FastHTTP server failed to start: %v", err)
		panic(err)
	}
}
