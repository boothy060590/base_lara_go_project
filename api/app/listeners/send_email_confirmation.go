package listeners

import (
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

	subject := "Welcome to Base Laravel Go Project!"
	body := fmt.Sprintf(`
		<h1>Welcome %s!</h1>
		<p>Thank you for registering with Base Laravel Go Project.</p>
		<p>Your account has been created successfully.</p>
		<p>Email: %s</p>
	`, user.FirstName, user.Email)

	// Cast the mailService to our interface and send the email
	if mailSvc, ok := mailService.(MailService); ok {
		err := mailSvc.SendMail([]string{user.Email}, subject, body)
		if err != nil {
			return fmt.Errorf("failed to send welcome email: %v", err)
		}
		fmt.Printf("Welcome email sent successfully to %s\n", user.Email)
	} else {
		// Fallback to console output if mail service is not available
		fmt.Printf("Sending email to %s: %s\nBody: %s\n", user.Email, subject, body)
	}

	return nil
}
