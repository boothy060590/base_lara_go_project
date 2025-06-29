package main

import (
	app_core "base_lara_go_project/app/core/app"
	events_core "base_lara_go_project/app/core/events"
	facades_core "base_lara_go_project/app/core/facades"
	jobs_core "base_lara_go_project/app/core/jobs"
	mail_core "base_lara_go_project/app/core/mail"
	queue_core "base_lara_go_project/app/core/queue"
	"base_lara_go_project/app/providers"
	"base_lara_go_project/config"
	"flag"
	"log"
)

func main() {
	// Parse command line arguments for worker type
	var workerType string
	flag.StringVar(&workerType, "worker", "default", "Worker type (default, jobs, mail, events)")
	flag.Parse()

	log.Printf("Starting %s worker...", workerType)

	// Register all service providers
	providers.RegisterFormFieldValidators()
	providers.RegisterDatabase()
	providers.RegisterCache(app_core.App)
	providers.RegisterMailer()
	providers.RegisterQueue()
	providers.RegisterJobDispatcher()
	providers.RegisterMessageProcessor()
	providers.RegisterEventDispatcher()
	providers.RegisterRepository()
	providers.RegisterServices() // Register service provider

	// Initialize core systems
	app_core.InitializeRegistry()
	events_core.InitializeEventDispatcher()

	// Register app-specific events
	providers.RegisterAppEvents()

	// Initialize email template engine
	if err := providers.RegisterMailTemplateEngine(); err != nil {
		log.Fatalf("Failed to initialize email template engine: %v", err)
	}

	// Set up the mail function for event dispatcher
	events_core.SetSendMailFunc(mail_core.SendMail)

	// Set up facades with concrete implementations
	facades_core.SetEventDispatcher(events_core.EventDispatcherServiceInstance)
	facades_core.SetJobDispatcher(jobs_core.JobDispatcherServiceInstance)
	facades_core.SetCache(facades_core.CacheInstance)

	// Register event listeners
	providers.RegisterListeners()

	// Register job processors
	providers.RegisterJobProcessors()

	providers.RunMigrations()

	log.Println("All service providers registered successfully")

	// Get worker configuration
	queueConfig := config.QueueConfig()
	workers := queueConfig["workers"].(map[string]interface{})
	workerConfig, ok := workers[workerType].(map[string]interface{})
	if !ok {
		workerConfig = workers["default"].(map[string]interface{})
	}
	workerQueuesIface := workerConfig["queues"].([]interface{})
	workerQueues := make([]string, len(workerQueuesIface))
	for i, q := range workerQueuesIface {
		workerQueues[i] = q.(string)
	}

	log.Printf("Worker type: %s", workerType)
	log.Printf("Assigned queues: %v", workerQueues)
	log.Printf("Max jobs: %v", workerConfig["max_jobs"])
	log.Printf("Memory limit: %v MB", workerConfig["memory_limit"])
	log.Printf("Timeout: %v seconds", workerConfig["timeout"])
	log.Printf("Sleep: %v seconds", workerConfig["sleep"])
	log.Printf("Tries: %v", workerConfig["tries"])

	// Start the worker with assigned queues
	worker := queue_core.NewQueueWorker(
		workerQueues,
		queue_core.QueueServiceInstance,
		jobs_core.JobDispatcherServiceInstance,
		app_core.App.Get("message_processor").(app_core.MessageProcessorService),
		workerConfig,
	)

	log.Printf("Starting queue worker for %s with %d assigned queues", workerType, len(workerQueues))
	worker.Start()
}
