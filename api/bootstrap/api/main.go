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

	// Try to get optimized HTTP server first, fallback to standard router
	if httpServerInstance, err := container.Resolve("http.server"); err == nil {
		// Use optimized HTTP server with fasthttp
		httpServer := httpServerInstance.(*app_core.HTTPOptimizer)
		appConfig := config.AppConfig()
		port := appConfig["port"].(string)
		
		log.Printf("Starting optimized HTTP server on port %s with fasthttp", port)
		if err := httpServer.ListenAndServe(":" + port); err != nil {
			panic(err)
		}
	} else {
		// Fallback to standard Gin router
		log.Printf("HTTP optimizer not available, falling back to standard Gin router: %v", err)
		
		routerInstance, err := container.Resolve("router")
		if err != nil {
			panic(err)
		}

		router := routerInstance.(*gin.Engine)
		appConfig := config.AppConfig()
		router.Run(":" + appConfig["port"].(string))
	}
}
