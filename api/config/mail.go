package config

import "base_lara_go_project/app/core/laravel_core/env"

// MailConfig returns the mail configuration with fallback values
func MailConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("MAIL_MAILER", "local"),
		"mailers": map[string]interface{}{
			"smtp": map[string]interface{}{
				"host":     env.Get("MAIL_HOST", "localhost"),
				"port":     env.Get("MAIL_PORT", "1025"),
				"username": env.Get("MAIL_USERNAME", ""),
				"password": env.Get("MAIL_PASSWORD", ""),
			},
			"local": map[string]interface{}{
				"driver": "local",
				"path":   env.Get("MAIL_PATH", "storage/logs/mail.log"),
			},
			"mailhog": map[string]interface{}{
				"driver": "mailhog",
				"host":   env.Get("MAILHOG_HOST", "localhost"),
				"port":   env.Get("MAILHOG_PORT", "1025"),
			},
		},
		"from": map[string]interface{}{
			"address": env.Get("MAIL_FROM_ADDRESS", "noreply@example.com"),
			"name":    env.Get("MAIL_FROM_NAME", "Base Laravel Go Project"),
		},
	}
}
