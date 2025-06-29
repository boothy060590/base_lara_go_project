package mail_core

import (
	"fmt"
	"log"
	"time"

	app_core "base_lara_go_project/app/core/app"
)

// LocalMailClient provides local mail functionality (logs emails)
type LocalMailClient struct {
	*BaseMailClient
}

// NewLocalMailClient creates a new local mail client
func NewLocalMailClient(config *app_core.ClientConfig) *LocalMailClient {
	return &LocalMailClient{
		BaseMailClient: NewBaseMailClient(config, "local"),
	}
}

// Connect establishes the mail connection (no-op for local mail)
func (c *LocalMailClient) Connect() error {
	return c.BaseClient.Connect()
}

// Disconnect closes the mail connection (no-op for local mail)
func (c *LocalMailClient) Disconnect() error {
	return c.BaseClient.Disconnect()
}

// Send sends an email (logs it locally)
func (c *LocalMailClient) Send(to []string, subject string, body string, options map[string]interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("mail client not connected")
	}

	// Log the email
	log.Printf("[MAIL] From: %s <%s>", c.GetFromName(), c.GetFrom())
	log.Printf("[MAIL] To: %v", to)
	log.Printf("[MAIL] Subject: %s", subject)
	log.Printf("[MAIL] Body: %s", body)
	if len(options) > 0 {
		log.Printf("[MAIL] Options: %+v", options)
	}
	log.Printf("[MAIL] Timestamp: %s", time.Now().Format(time.RFC3339))
	log.Printf("[MAIL] ---")

	return nil
}

// SendWithAttachments sends an email with attachments (logs it locally)
func (c *LocalMailClient) SendWithAttachments(to []string, subject string, body string, attachments []string, options map[string]interface{}) error {
	if !c.IsConnected() {
		return fmt.Errorf("mail client not connected")
	}

	// Log the email with attachments
	log.Printf("[MAIL] From: %s <%s>", c.GetFromName(), c.GetFrom())
	log.Printf("[MAIL] To: %v", to)
	log.Printf("[MAIL] Subject: %s", subject)
	log.Printf("[MAIL] Body: %s", body)
	log.Printf("[MAIL] Attachments: %v", attachments)
	if len(options) > 0 {
		log.Printf("[MAIL] Options: %+v", options)
	}
	log.Printf("[MAIL] Timestamp: %s", time.Now().Format(time.RFC3339))
	log.Printf("[MAIL] ---")

	return nil
}

// GetStats returns mail statistics
func (c *LocalMailClient) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"status":    "connected",
		"driver":    "local",
		"from":      c.GetFrom(),
		"from_name": c.GetFromName(),
	}
}
