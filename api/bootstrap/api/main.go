package main

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
	"base_lara_go_project/app/providers"
	"base_lara_go_project/config"

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

	// Get router from container
	routerInstance, err := container.Resolve("router")
	if err != nil {
		panic(err)
	}

	router := routerInstance.(*gin.Engine)

	// Start the server
	appConfig := config.AppConfig()
	router.Run(":" + appConfig["port"].(string))
}
