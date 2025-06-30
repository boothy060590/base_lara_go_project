package main

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_providers "base_lara_go_project/app/core/laravel_core/providers"
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

	// Create container instance
	container := app_core.NewContainer()

	// Register core services automatically
	coreProvider := &laravel_providers.CoreServiceProvider{}
	if err := coreProvider.Register(container); err != nil {
		log.Fatalf("Failed to register core services: %v", err)
	}

	// Register application providers
	appProviders := []laravel_providers.ServiceProvider{
		&providers.AppServiceProvider{},
		&providers.RepositoryServiceProvider{},
		&providers.ListenerServiceProvider{},
		&providers.RouterServiceProvider{},
		&laravel_providers.ValidationServiceProvider{},
	}

	for _, provider := range appProviders {
		if err := provider.Register(container); err != nil {
			log.Printf("Warning: Failed to register provider %T: %v", provider, err)
		}
	}

	// Boot all providers
	if err := coreProvider.Boot(container); err != nil {
		log.Fatalf("Failed to boot core services: %v", err)
	}

	for _, provider := range appProviders {
		if err := provider.Boot(container); err != nil {
			log.Printf("Warning: Failed to boot provider %T: %v", provider, err)
		}
	}

	// Run database migrations if migration provider exists
	migrationProvider := &laravel_providers.MigrationServiceProvider{}
	if err := migrationProvider.Register(container); err != nil {
		log.Printf("Warning: Failed to register migration provider: %v", err)
	} else {
		if err := migrationProvider.Boot(container); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		}
	}

	log.Println("All service providers registered and booted successfully")

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

	// TODO: Implement proper queue worker with new go_core structure
	log.Printf("Queue worker not yet implemented with new core structure")
	log.Printf("Starting queue worker for %s with %d assigned queues", workerType, len(workerQueues))

	// Keep the process running for now
	select {}
}
