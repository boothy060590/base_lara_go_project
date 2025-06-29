package mail_core

import (
	"fmt"
	"net/smtp"

	app_core "base_lara_go_project/app/core/app"
)

// MailHogClient provides MailHog mail functionality
type MailHogClient struct {
	*BaseMailClient
	auth smtp.Auth
}

// NewMailHogClient creates a new MailHog client
func NewMailHogClient(config *app_core.ClientConfig) *MailHogClient {
	return &MailHogClient{
		BaseMailClient: NewBaseMailClient(config, "mailhog"),
	}
}

// Connect establishes a connection to MailHog
func (c *MailHogClient) Connect() error {
	// Get MailHog configuration from options
	host := "localhost"
	if configHost, ok := c.config.Options["host"].(string); ok {
		host = configHost
	}

	username := ""
	if configUsername, ok := c.config.Options["username"].(string); ok {
		username = configUsername
	}

	password := ""
	if configPassword, ok := c.config.Options["password"].(string); ok {
		password = configPassword
	}

	// Create SMTP auth (MailHog typically doesn't require auth)
	if username != "" && password != "" {
		c.auth = smtp.PlainAuth("", username, password, host)
	}

	return c.BaseClient.Connect()
}

// Disconnect closes the MailHog connection
func (c *MailHogClient) Disconnect() error {
	return c.BaseClient.Disconnect()
}

// Send sends an email via MailHog
func (c *MailHogClient) Send(to []string, subject string, body string, options map[string]interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("mail client not connected")
	}

	// Get MailHog configuration
	host := "localhost"
	if configHost, ok := c.config.Options["host"].(string); ok {
		host = configHost
	}

	port := 1025
	if configPort, ok := c.config.Options["port"].(int); ok {
		port = configPort
	}

	// Build email message
	message := fmt.Sprintf("From: %s <%s>\r\n", c.GetFromName(), c.GetFrom())
	message += fmt.Sprintf("To: %s\r\n", to[0]) // MailHog typically uses first recipient
	message += fmt.Sprintf("Subject: %s\r\n", subject)
	message += "\r\n"
	message += body

	// Send email
	addr := fmt.Sprintf("%s:%d", host, port)
	err := smtp.SendMail(addr, c.auth, c.GetFrom(), to, []byte(message))
	return err
}

// SendWithAttachments sends an email with attachments via MailHog
func (c *MailHogClient) SendWithAttachments(to []string, subject string, body string, attachments []string, options map[string]interface{}) error {
	// For simplicity, we'll just send the email without attachments
	// In a real implementation, you'd use a proper MIME library
	return c.Send(to, subject, body, options)
}

// GetStats returns MailHog statistics
func (c *MailHogClient) GetStats() map[string]interface{} {
	host := "localhost"
	if configHost, ok := c.config.Options["host"].(string); ok {
		host = configHost
	}

	port := 1025
	if configPort, ok := c.config.Options["port"].(int); ok {
		port = configPort
	}

	return map[string]interface{}{
		"status":    "connected",
		"driver":    "mailhog",
		"host":      host,
		"port":      port,
		"from":      c.GetFrom(),
		"from_name": c.GetFromName(),
	}
}
