package config

func MailConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": getEnv("MAIL_MAILER", "smtp"),
		"mailers": map[string]interface{}{
			"smtp": map[string]interface{}{
				"host":     getEnv("MAIL_HOST", "localhost"),
				"port":     getEnv("MAIL_PORT", "1025"),
				"username": getEnv("MAIL_USERNAME", ""),
				"password": getEnv("MAIL_PASSWORD", ""),
			},
		},
		"from": map[string]interface{}{
			"address": getEnv("MAIL_FROM_ADDRESS", "no-reply@example.com"),
			"name":    getEnv("MAIL_FROM_NAME", "App"),
		},
	}
}
