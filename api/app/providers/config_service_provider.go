package providers

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/config"
)

// RegisterConfig loads all config files and registers them with the config registry
func RegisterConfig() {
	core.LoadConfig(map[string]map[string]interface{}{
		"app":      config.AppConfig(),
		"database": config.DatabaseConfig(),
		"mail":     config.MailConfig(),
		"queue":    config.QueueConfig(),
	})
}
