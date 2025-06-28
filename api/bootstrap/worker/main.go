package main

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/providers"
	"base_lara_go_project/config"
	"log"
)

func main() {
	log.Println("Starting worker...")

	// Register config first
	providers.RegisterConfig()

	// Register all service providers
	providers.RegisterFormFieldValidators()
	providers.RegisterDatabase()
	providers.RegisterCache()
	providers.RegisterMailer()
	providers.RegisterQueue()
	providers.RegisterJobDispatcher()
	providers.RegisterMessageProcessor()
	providers.RegisterEventDispatcher()
	providers.RegisterRepository()
	providers.RegisterServices() // Register service provider

	// Initialize core systems
	core.InitializeRegistry()
	core.InitializeEventDispatcher()

	// Register app-specific events
	providers.RegisterAppEvents()

	// Initialize email template engine
	if err := providers.RegisterMailTemplateEngine(); err != nil {
		log.Fatalf("Failed to initialize email template engine: %v", err)
	}

	// Set up the mail function for event dispatcher
	core.SetSendMailFunc(core.SendMail)

	// Set up facades with concrete implementations
	facades.SetEventDispatcher(core.EventDispatcherServiceInstance)
	facades.SetJobDispatcher(core.JobDispatcherServiceInstance)
	facades.SetCache(core.CacheInstance)

	// Register event listeners
	providers.RegisterListeners()

	// Register job processors
	providers.RegisterJobProcessors()

	providers.RunMigrations()

	log.Println("All service providers registered successfully")

	// Start a worker for all enabled queues
	queueConfig := config.QueueConfig()
	enabledQueues := queueConfig["enabled_queues"].([]string)
	worker := core.NewQueueWorker(enabledQueues)

	log.Printf("Starting queue worker with %d enabled queues", len(enabledQueues))
	worker.Start()
}
