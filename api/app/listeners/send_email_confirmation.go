package listeners

import (
	"base_lara_go_project/app/core"
	authEvents "base_lara_go_project/app/events/auth"
	"fmt"
)

// MailService defines the interface for sending emails
type MailService interface {
	SendMail(to []string, subject, body string) error
}

type SendEmailConfirmation struct {
	BaseListener
	Event authEvents.UserCreated
}

func (l *SendEmailConfirmation) Handle(mailService interface{}) error {
	user := l.Event.GetUser()

	// Prepare template data
	templateData := core.EmailTemplateData{
		Subject:        "Welcome to Base Laravel Go Project!",
		AppName:        "Base Laravel Go Project",
		RecipientEmail: user.Email,
		User:           user,
		LoginURL:       "https://app.baselaragoproject.test/login", // You can make this configurable
	}

	// Render email template
	body, err := core.RenderEmailTemplate("auth/welcome", templateData)
	if err != nil {
		return fmt.Errorf("failed to render email template: %v", err)
	}

	// Cast the mailService to our interface and send the email
	if mailSvc, ok := mailService.(MailService); ok {
		err := mailSvc.SendMail([]string{user.Email}, templateData.Subject, body)
		if err != nil {
			return fmt.Errorf("failed to send welcome email: %v", err)
		}
		fmt.Printf("Welcome email sent successfully to %s\n", user.Email)
	} else {
		// Fallback to console output if mail service is not available
		fmt.Printf("Sending email to %s: %s\nBody: %s\n", user.Email, templateData.Subject, body)
	}

	return nil
}
