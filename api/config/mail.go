package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// MailConfig returns the mail configuration with environment variable fallbacks
// This config defines mail drivers, SMTP settings, and mail server parameters
func MailConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("MAIL_MAILER", "smtp"),
		"mailers": map[string]interface{}{
			"smtp": map[string]interface{}{
				"transport":  "smtp",
				"host":       env.Get("MAIL_HOST", "smtp.mailgun.org"),
				"port":       env.Get("MAIL_PORT", "587"),
				"encryption": env.Get("MAIL_ENCRYPTION", "tls"),
				"username":   env.Get("MAIL_USERNAME", ""),
				"password":   env.Get("MAIL_PASSWORD", ""),
			},
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the mail config is available via config.Get("mail") and dot notation
func init() {
	go_core.RegisterGlobalConfig("mail", MailConfig)
}
