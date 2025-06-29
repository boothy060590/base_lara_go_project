package mail_core

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	clients_core "base_lara_go_project/app/core/clients"
)

// MailProviderFactory creates mail providers based on configuration
type MailProviderFactory struct {
	container *app_core.ServiceContainer
}

// NewMailProviderFactory creates a new mail provider factory
func NewMailProviderFactory(container *app_core.ServiceContainer) *MailProviderFactory {
	return &MailProviderFactory{container: container}
}

var mailProviderMap = map[string]func(cfg *clients_core.ClientConfig) app_core.MailProviderServiceInterface{
	"local": func(cfg *clients_core.ClientConfig) app_core.MailProviderServiceInterface {
		return NewLocalMailProvider(NewLocalMailClient(cfg))
	},
	"mailhog": func(cfg *clients_core.ClientConfig) app_core.MailProviderServiceInterface {
		return NewMailHogMailProvider(NewMailHogClient(cfg))
	},
}

// Create creates a mail provider based on the driver
func (f *MailProviderFactory) Create(driver string, cfg *clients_core.ClientConfig) (app_core.MailProviderServiceInterface, error) {
	constructor, ok := mailProviderMap[driver]
	if !ok {
		return nil, fmt.Errorf("unsupported mail driver: %s", driver)
	}
	return constructor(cfg), nil
}

// RegisterFromConfig registers a mail provider from configuration
func (f *MailProviderFactory) RegisterFromConfig(config map[string]interface{}) error {
	// Get default mailer from config
	defaultMailer, ok := config["default"].(string)
	if !ok {
		return fmt.Errorf("default mailer not set in config")
	}

	// Get mailers configuration
	mailers, ok := config["mailers"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("mailers not configured")
	}

	// Get the specific mailer config
	mailerConfig, ok := mailers[defaultMailer].(map[string]interface{})
	if !ok {
		return fmt.Errorf("mailer config for %s not found", defaultMailer)
	}

	// Get from configuration
	fromConfig, ok := config["from"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("from configuration not found")
	}

	// Build client config from mailer config
	clientConfig := f.buildClientConfig(defaultMailer, mailerConfig, fromConfig)

	provider, err := f.Create(defaultMailer, clientConfig)
	if err != nil {
		return err
	}

	f.container.Singleton("mail.provider", provider)
	return nil
}

// buildClientConfig converts mailer config to client config
func (f *MailProviderFactory) buildClientConfig(driver string, mailerConfig map[string]interface{}, fromConfig map[string]interface{}) *clients_core.ClientConfig {
	clientConfig := &clients_core.ClientConfig{
		Driver:  driver,
		Options: mailerConfig,
	}

	// Set common fields
	if host, ok := mailerConfig["host"].(string); ok {
		clientConfig.Host = host
	}
	if username, ok := mailerConfig["username"].(string); ok {
		clientConfig.Username = username
	}
	if password, ok := mailerConfig["password"].(string); ok {
		clientConfig.Password = password
	}

	// Set port based on driver
	switch driver {
	case "mailhog":
		if port, ok := mailerConfig["port"].(int); ok {
			clientConfig.Port = port
		} else {
			clientConfig.Port = 1025
		}
	case "local":
		// Local mail doesn't use host/port
		clientConfig.Host = "localhost"
		clientConfig.Port = 0
	}

	// Add from configuration to options
	clientConfig.Options["from"] = fromConfig["address"]
	clientConfig.Options["from_name"] = fromConfig["name"]

	return clientConfig
}
