package providers

import (
	"fmt"

	app_core "base_lara_go_project/app/core/app"
	mail_core "base_lara_go_project/app/core/mail"
	"base_lara_go_project/config"
)

// MailServiceProvider registers mail services with the container
type MailServiceProvider struct{}

// NewMailServiceProvider creates a new mail service provider
func NewMailServiceProvider() *MailServiceProvider {
	return &MailServiceProvider{}
}

// Register registers mail services with the container
func (p *MailServiceProvider) Register() error {
	// Create mail provider factory
	factory := mail_core.NewMailProviderFactory(app_core.App)

	// Get mail configuration
	mailConfig := config.MailConfig()

	// Register mail provider from config
	if err := factory.RegisterFromConfig(mailConfig); err != nil {
		return fmt.Errorf("failed to register mail provider: %w", err)
	}

	return nil
}

// RegisterMailer registers the mail service provider
func RegisterMailer() error {
	provider := NewMailServiceProvider()
	return provider.Register()
}
