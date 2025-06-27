package config

import (
	"os"
)

func AppConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":                getEnv("APP_NAME", "Base Laravel Go Project"),
		"env":                 getEnv("APP_ENV", "development"),
		"debug":               getEnv("APP_DEBUG", "false"),
		"url":                 getEnv("APP_URL", "http://localhost"),
		"port":                getEnv("APP_PORT", "8080"),
		"secret":              getEnv("API_SECRET", "changeme"),
		"token_hour_lifespan": getEnv("TOKEN_HOUR_LIFESPAN", "1"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
