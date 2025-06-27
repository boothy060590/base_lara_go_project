package main

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/providers"
	_ "base_lara_go_project/routes/api/v1/auth"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// register service providers
	providers.RegisterFormFieldValidators()
	providers.RegisterDatabase()
	providers.RegisterMailer()
	providers.RegisterQueue()

	// Initialize core systems
	core.InitializeRegistry()
	core.InitializeEventDispatcher()

	// Initialize email template engine
	if err := providers.RegisterMailTemplateEngine(); err != nil {
		panic("Failed to initialize email template engine: " + err.Error())
	}

	// Set up the mail function for event dispatcher
	core.SetSendMailFunc(providers.SendMail)

	// Set up facades with concrete implementations
	facades.SetEventDispatcher(providers.NewEventDispatcherProvider())
	facades.SetJobDispatcher(providers.NewJobDispatcherProvider())

	// Register event listeners
	providers.RegisterListeners()

	providers.RunMigrations()

	router := gin.Default()
	providers.RegisterRoutes(router)
	router.Run(":" + os.Getenv("APP_PORT"))
}
