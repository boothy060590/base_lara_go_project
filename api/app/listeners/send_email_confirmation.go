package listeners

import (
	"base_lara_go_project/app/core"
	authEvents "base_lara_go_project/app/events/auth"
	"base_lara_go_project/app/facades"
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

// RegisterSelf registers this listener with the event system
func RegisterSelf() {
	core.RegisterEvent("UserCreated", func(e core.EventInterface) core.ListenerInterface {
		listener := &SendEmailConfirmation{}
		if userCreated, ok := e.(*authEvents.UserCreated); ok {
			listener.Event = *userCreated
		}
		return listener
	})
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

	// Send email asynchronously via mail queue
	err = facades.MailAsync([]string{user.Email}, templateData.Subject, body)
	if err != nil {
		return fmt.Errorf("failed to queue welcome email: %v", err)
	}

	return nil
}
