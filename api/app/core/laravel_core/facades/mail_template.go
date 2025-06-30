package facades_core

import (
	"base_lara_go_project/config"
	"fmt"
)

// EmailTemplateData defines the structure for email template data
type EmailTemplateData struct {
	Subject string
	Data    map[string]interface{}
}

// MailTemplate sends an email using a template
func MailTemplate(to []string, templateName string, data EmailTemplateData) error {
	// TODO: Implement template rendering and email sending
	_ = fmt.Sprintf("Template: %s, Subject: %s", templateName, data.Subject)

	return nil
}

// MailTemplateAsync sends an email using a template asynchronously via queue
func MailTemplateAsync(to []string, templateName string, data EmailTemplateData) error {
	// TODO: Implement template rendering and async email sending
	_ = fmt.Sprintf("Template: %s, Subject: %s", templateName, data.Subject)

	// TODO: Implement async email sending via queue
	queueConfig := config.QueueConfig()
	queues := queueConfig["queues"].(map[string]interface{})
	queueName := queues["mail"].(string)
	_ = queueName // Use queueName to avoid unused variable error

	return nil
}

// MailTemplateToUser sends a templated email to a specific user
func MailTemplateToUser(user interface{}, templateName string, data EmailTemplateData) error {
	// Extract email from user (assuming user has GetEmail method)
	if userWithEmail, ok := user.(interface{ GetEmail() string }); ok {
		to := []string{userWithEmail.GetEmail()}
		return MailTemplate(to, templateName, data)
	}

	// If user doesn't have GetEmail method, return an error
	return fmt.Errorf("user does not implement GetEmail() method")
}

// MailTemplateToUserAsync sends a templated email to a specific user asynchronously
func MailTemplateToUserAsync(user interface{}, templateName string, data EmailTemplateData) error {
	// Extract email from user (assuming user has GetEmail method)
	if userWithEmail, ok := user.(interface{ GetEmail() string }); ok {
		to := []string{userWithEmail.GetEmail()}
		return MailTemplateAsync(to, templateName, data)
	}

	// If user doesn't have GetEmail method, return an error
	return fmt.Errorf("user does not implement GetEmail() method")
}
