package mail_core

import (
	app_core "base_lara_go_project/app/core/app"
)

// LocalMailProvider provides local mail services
type LocalMailProvider struct {
	client *LocalMailClient
}

// NewLocalMailProvider creates a new local mail provider
func NewLocalMailProvider(client *LocalMailClient) *LocalMailProvider {
	return &LocalMailProvider{
		client: client,
	}
}

// Connect establishes a connection to the mail service
func (p *LocalMailProvider) Connect() error {
	return p.client.Connect()
}

// Disconnect closes the mail connection
func (p *LocalMailProvider) Disconnect() error {
	return p.client.Disconnect()
}

// Send sends an email
func (p *LocalMailProvider) Send(to []string, subject string, body string, options map[string]interface{}) error {
	return p.client.Send(to, subject, body, options)
}

// SendWithAttachments sends an email with attachments
func (p *LocalMailProvider) SendWithAttachments(to []string, subject string, body string, attachments []string, options map[string]interface{}) error {
	return p.client.SendWithAttachments(to, subject, body, attachments, options)
}

// GetStats returns mail statistics
func (p *LocalMailProvider) GetStats() map[string]interface{} {
	return p.client.GetStats()
}

// GetClient returns the underlying mail client
func (p *LocalMailProvider) GetClient() app_core.MailClientInterface {
	return p.client
}
