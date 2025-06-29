package config

import "base_lara_go_project/app/core/env"

// MailConfig returns the mail configuration with fallback values
func MailConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.GetEnv("MAIL_MAILER", "local"),
		"mailers": map[string]interface{}{
			"smtp": map[string]interface{}{
				"host":     env.GetEnv("MAIL_HOST", "localhost"),
				"port":     env.GetEnv("MAIL_PORT", "1025"),
				"username": env.GetEnv("MAIL_USERNAME", ""),
				"password": env.GetEnv("MAIL_PASSWORD", ""),
			},
			"local": map[string]interface{}{
				"driver": "local",
				"path":   env.GetEnv("MAIL_PATH", "storage/logs/mail.log"),
			},
			"mailhog": map[string]interface{}{
				"driver": "mailhog",
				"host":   env.GetEnv("MAILHOG_HOST", "localhost"),
				"port":   env.GetEnv("MAILHOG_PORT", "1025"),
			},
		},
		"from": map[string]interface{}{
			"address": env.GetEnv("MAIL_FROM_ADDRESS", "noreply@example.com"),
			"name":    env.GetEnv("MAIL_FROM_NAME", "Base Laravel Go Project"),
		},
	}
}
