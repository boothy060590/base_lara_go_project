package listeners

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_listeners "base_lara_go_project/app/core/laravel_core/listeners"
	auth_dto "base_lara_go_project/app/data_objects/auth"
	"context"
	"fmt"
)

// SendEmailConfirmation handles sending welcome emails to new users
type SendEmailConfirmation struct {
	laravel_listeners.BaseListener[auth_dto.UserDTO]
}

// Handle processes the user created event and sends a welcome email
func (l *SendEmailConfirmation) Handle(ctx context.Context, event *app_core.Event[auth_dto.UserDTO]) error {
	user := event.Data

	// TODO: Implement email sending using go_core mail system
	fmt.Printf("Sending welcome email to: %s\n", user.Email)

	// Example of how to use the mail system:
	// mailService := ctx.Value("mail_service").(app_core.MailService)
	// return mailService.Send(
	//     []string{user.Email},
	//     "Welcome to our platform!",
	//     "welcome",
	//     map[string]interface{}{
	//         "user": user,
	//     },
	// )

	return nil
}
