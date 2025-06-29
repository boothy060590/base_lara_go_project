package providers

import mail_core "base_lara_go_project/app/core/mail"

// RegisterMailTemplateEngine initializes the email template engine
func RegisterMailTemplateEngine() error {
	return mail_core.InitializeEmailTemplateEngine()
}
