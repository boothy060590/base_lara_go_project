package providers

import "base_lara_go_project/app/core"

// RegisterMailTemplateEngine initializes the email template engine
func RegisterMailTemplateEngine() error {
	return core.InitializeEmailTemplateEngine()
}
