package mail_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// MailHogMailProvider provides MailHog mail services
type MailHogMailProvider struct {
	client *MailHogClient
}

// NewMailHogMailProvider creates a new MailHog mail provider
func NewMailHogMailProvider(client *MailHogClient) *MailHogMailProvider {
	return &MailHogMailProvider{
		client: client,
	}
}

// Connect establishes a connection to the mail service
func (p *MailHogMailProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the mail connection
func (p *MailHogMailProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Send sends an email
func (p *MailHogMailProvider) Send(to []string, subject string, body string, options map[string]interface{}) error {
	return p.client.Send(to, subject, body, options)
}

// SendWithAttachments sends an email with attachments
func (p *MailHogMailProvider) SendWithAttachments(to []string, subject string, body string, attachments []string, options map[string]interface{}) error {
	return p.client.SendWithAttachments(to, subject, body, attachments, options)
}

// GetStats returns mail statistics
func (p *MailHogMailProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying mail client
func (p *MailHogMailProvider) GetClient() app_core.MailClientInterface {
	return p.client
}
