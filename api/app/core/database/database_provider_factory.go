package database_core

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	clients_core "base_lara_go_project/app/core/clients"
)

type DatabaseProviderFactory struct {
	container *app_core.ServiceContainer
}

func NewDatabaseProviderFactory(container *app_core.ServiceContainer) *DatabaseProviderFactory {
	return &DatabaseProviderFactory{container: container}
}

var databaseProviderMap = map[string]func(cfg *clients_core.ClientConfig) app_core.DatabaseProviderServiceInterface{
	"mysql": func(cfg *clients_core.ClientConfig) app_core.DatabaseProviderServiceInterface {
		return NewMySQLDatabaseProvider(NewMySQLClient(cfg))
	},
	"postgres": func(cfg *clients_core.ClientConfig) app_core.DatabaseProviderServiceInterface {
		return NewPostgresDatabaseProvider(NewPostgresClient(cfg))
	},
	"sqlite": func(cfg *clients_core.ClientConfig) app_core.DatabaseProviderServiceInterface {
		return NewSQLiteDatabaseProvider(NewSQLiteClient(cfg))
	},
}

func (f *DatabaseProviderFactory) Create(driver string, cfg *clients_core.ClientConfig) (app_core.DatabaseProviderServiceInterface, error) {
	constructor, ok := databaseProviderMap[driver]
	if !ok {
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
	return constructor(cfg), nil
}

func (f *DatabaseProviderFactory) RegisterFromConfig(config map[string]interface{}) error {
	// Get default connection from config
	defaultConnection, ok := config["default"].(string)
	if !ok {
		return fmt.Errorf("default database connection not set in config")
	}

	// Get connections configuration
	connections, ok := config["connections"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("database connections not configured")
	}

	// Get the specific connection config
	connectionConfig, ok := connections[defaultConnection].(map[string]interface{})
	if !ok {
		return fmt.Errorf("connection config for %s not found", defaultConnection)
	}

	// Build client config from connection config
	clientConfig := f.buildClientConfig(defaultConnection, connectionConfig)

	provider, err := f.Create(defaultConnection, clientConfig)
	if err != nil {
		return err
	}

	f.container.Singleton("database.provider", provider)
	return nil
}

// buildClientConfig converts connection config to client config
func (f *DatabaseProviderFactory) buildClientConfig(driver string, config map[string]interface{}) *clients_core.ClientConfig {
	clientConfig := &clients_core.ClientConfig{
		Driver:  driver,
		Options: config,
	}

	// Set common fields
	if host, ok := config["host"].(string); ok {
		clientConfig.Host = host
	}
	if database, ok := config["database"].(string); ok {
		clientConfig.Database = database
	}
	if username, ok := config["username"].(string); ok {
		clientConfig.Username = username
	}
	if password, ok := config["password"].(string); ok {
		clientConfig.Password = password
	}

	// Set port based on driver
	switch driver {
	case "mysql":
		if port, ok := config["port"].(int); ok {
			clientConfig.Port = port
		} else {
			clientConfig.Port = 3306
		}
	case "postgres":
		if port, ok := config["port"].(int); ok {
			clientConfig.Port = port
		} else {
			clientConfig.Port = 5432
		}
	case "sqlite":
		// SQLite doesn't use host/port, but we'll set defaults
		clientConfig.Host = "localhost"
		clientConfig.Port = 0
	}

	return clientConfig
}
