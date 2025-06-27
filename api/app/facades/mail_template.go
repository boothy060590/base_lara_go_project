package facades

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/providers"
	"fmt"
)

// MailTemplate sends an email using a template
func MailTemplate(to []string, templateName string, data core.EmailTemplateData) error {
	// Render the template
	body, err := core.RenderEmailTemplate(templateName, data)
	if err != nil {
		return err
	}

	// Send the email
	return providers.SendMail(to, data.Subject, body)
}

// MailTemplateAsync sends an email using a template asynchronously via queue
func MailTemplateAsync(to []string, templateName string, data core.EmailTemplateData) error {
	// Render the template
	body, err := core.RenderEmailTemplate(templateName, data)
	if err != nil {
		return err
	}

	// Send the email asynchronously
	return providers.SendMailAsync(to, data.Subject, body)
}

// MailTemplateToUser sends a templated email to a specific user
func MailTemplateToUser(user interface{}, templateName string, data core.EmailTemplateData) error {
	// Extract email from user (assuming user has GetEmail method)
	if userWithEmail, ok := user.(interface{ GetEmail() string }); ok {
		to := []string{userWithEmail.GetEmail()}
		return MailTemplate(to, templateName, data)
	}

	// If user doesn't have GetEmail method, return an error
	return fmt.Errorf("user does not implement GetEmail() method")
}

// MailTemplateToUserAsync sends a templated email to a specific user asynchronously
func MailTemplateToUserAsync(user interface{}, templateName string, data core.EmailTemplateData) error {
	// Extract email from user (assuming user has GetEmail method)
	if userWithEmail, ok := user.(interface{ GetEmail() string }); ok {
		to := []string{userWithEmail.GetEmail()}
		return MailTemplateAsync(to, templateName, data)
	}

	// If user doesn't have GetEmail method, return an error
	return fmt.Errorf("user does not implement GetEmail() method")
}
