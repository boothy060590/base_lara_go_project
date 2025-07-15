package config

import (
	go_core "base_lara_go_project/app/core/go_core"
	"os"
)

// HTTPConfig returns HTTP optimization configuration
func HTTPConfig() map[string]any {
	return map[string]any{
		// HTTP server settings
		"http": map[string]any{
			"enable_fasthttp":        true,
			"read_timeout":          30,    // seconds
			"write_timeout":         30,    // seconds  
			"idle_timeout":          120,   // seconds
			"max_request_body_size": 10485760, // 10MB
		},
		
		// Connection pooling settings
		"http_pool": map[string]any{
			"max_connections":     15000, // Test value - increased from 10000
			"max_idle_connections": 1500, // Test value - increased from 1000  
			"connection_timeout":   5, // seconds
		},
		
		// Performance optimizations
		"http_performance": map[string]any{
			"enable_compression":    true,
			"enable_keep_alive":     true,
			"pre_allocated_buffers": 100,
			"zero_copy_enabled":     true,
		},
		
		// CORS settings
		"http_cors": map[string]any{
			"allowed_origins":    []string{getAppOrigin()},
			"allowed_methods":    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			"allowed_headers":    []string{"Origin", "Content-Type", "Accept", "Authorization"},
			"exposed_headers":    []string{"Content-Length"},
			"allow_credentials":  true,
			"max_age":           86400, // 24 hours
		},
		
		// Monitoring settings
		"http_monitoring": map[string]any{
			"enable_metrics": true,
			"metrics_path":   "/metrics",
		},
	}
}

// getAppOrigin returns the app origin URL based on APP_DOMAIN environment variable
func getAppOrigin() string {
	appDomain := os.Getenv("APP_DOMAIN")
	if appDomain == "" {
		appDomain = "baselaragoproject.test" // fallback
	}
	return "https://app." + appDomain
}

// init registers the HTTP configuration with the global config loader
func init() {
	go_core.RegisterGlobalConfig("http", HTTPConfig)
}