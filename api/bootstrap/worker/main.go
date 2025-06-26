package main

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/providers"
	"log"
	"os"
	"time"
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

	// Set up the mail function for event dispatcher
	core.SetSendMailFunc(providers.SendMail)

	// Set up facades with concrete implementations
	facades.SetEventDispatcher(providers.NewEventDispatcherProvider())
	facades.SetJobDispatcher(providers.NewJobDispatcherProvider())

	// Register event listeners
	providers.RegisterListeners()

	providers.RunMigrations()

	// Create queue if it doesn't exist
	queueName := os.Getenv("QUEUE_NAME")
	if queueName == "" {
		queueName = "default"
	}

	// Start the worker
	log.Printf("Starting worker for queue: %s", queueName)
	for {
		messageProcessor := providers.NewMessageProcessorProvider()
		if err := messageProcessor.ProcessMessages(); err != nil {
			log.Printf("Error processing messages: %v", err)
		}
		time.Sleep(1 * time.Second)
	}
}
