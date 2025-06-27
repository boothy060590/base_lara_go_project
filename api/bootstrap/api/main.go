package main

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/providers"
	"base_lara_go_project/config"
	_ "base_lara_go_project/routes/api/v1/auth"

	"github.com/gin-gonic/gin"
)

func main() {
	// register config first
	providers.RegisterConfig()

	// register service providers
	providers.RegisterFormFieldValidators()
	providers.RegisterDatabase()
	providers.RegisterMailer()
	providers.RegisterQueue()
	providers.RegisterJobDispatcher()
	providers.RegisterMessageProcessor()
	providers.RegisterEventDispatcher()

	// Initialize core systems
	core.InitializeRegistry()
	core.InitializeEventDispatcher()

	// Register app-specific events
	providers.RegisterAppEvents()

	// Initialize email template engine
	if err := providers.RegisterMailTemplateEngine(); err != nil {
		panic("Failed to initialize email template engine: " + err.Error())
	}

	// Set up the mail function for event dispatcher
	core.SetSendMailFunc(core.SendMail)

	// Set up facades with concrete implementations
	facades.SetEventDispatcher(core.EventDispatcherServiceInstance)
	facades.SetJobDispatcher(core.JobDispatcherServiceInstance)

	// Register event listeners
	providers.RegisterListeners()

	// Register job processors
	providers.RegisterJobProcessors()

	providers.RunMigrations()

	router := gin.Default()
	providers.RegisterRoutes(router)
	appConfig := config.AppConfig()
	router.Run(":" + appConfig["port"].(string))
}
