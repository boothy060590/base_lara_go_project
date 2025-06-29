package main

import (
	app_core "base_lara_go_project/app/core/app"
	cache_core "base_lara_go_project/app/core/cache"
	events_core "base_lara_go_project/app/core/events"
	facades_core "base_lara_go_project/app/core/facades"
	job_core "base_lara_go_project/app/core/jobs"
	mail_core "base_lara_go_project/app/core/mail"
	"base_lara_go_project/app/providers"
	"base_lara_go_project/config"
	_ "base_lara_go_project/routes/api/v1/auth"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	godotenv.Load()
	// register service providers
	providers.RegisterFormFieldValidators()
	providers.RegisterDatabase()
	providers.RegisterCache(app_core.App)
	providers.RegisterMailer()
	providers.RegisterQueue()
	providers.RegisterJobDispatcher()
	providers.RegisterMessageProcessor()
	providers.RegisterEventDispatcher()
	providers.RegisterRepository()
	providers.RegisterServices()
	providers.RegisterLogging()

	// Initialize core systems
	app_core.InitializeRegistry()
	events_core.InitializeEventDispatcher()

	// Register app-specific events
	providers.RegisterAppEvents()

	// Initialize email template engine
	if err := providers.RegisterMailTemplateEngine(); err != nil {
		panic("Failed to initialize email template engine: " + err.Error())
	}

	// Set up the mail function for event dispatcher
	events_core.SetSendMailFunc(mail_core.SendMail)

	// Set up facades with concrete implementations
	facades_core.SetEventDispatcher(events_core.EventDispatcherServiceInstance)
	facades_core.SetJobDispatcher(job_core.JobDispatcherServiceInstance)
	facades_core.SetCache(cache_core.CacheInstance)

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
