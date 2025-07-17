package config

import (
	"base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/core/laravel_core/env"
)

// JSONOptimizationsConfig returns the JSON optimization configuration with environment variable fallbacks
// This config defines JSON processing settings, buffer pools, response pools, and schema definitions
func JSONOptimizationsConfig() map[string]interface{} {
	return map[string]interface{}{
		// Global enable/disable for JSON optimizations
		"enabled": env.GetBool("JSON_OPTIMIZATIONS_ENABLED", true),

		// Buffer pool configuration
		"buffer_pool": map[string]interface{}{
			"max_buffer_size": env.GetInt("JSON_BUFFER_POOL_MAX_SIZE", 64*1024), // 64KB
			"initial_size":    env.GetInt("JSON_BUFFER_POOL_INITIAL_SIZE", 4*1024), // 4KB
		},

		// Response pool configuration
		"response_pool": map[string]interface{}{
			"pool_size": env.GetInt("JSON_RESPONSE_POOL_SIZE", 500),
		},

		// Metrics configuration
		"metrics": map[string]interface{}{
			"enabled":             env.GetBool("JSON_METRICS_ENABLED", true),
			"collection_interval": env.GetInt("JSON_METRICS_COLLECTION_INTERVAL", 300), // 5 minutes in seconds
			"report_interval":     env.GetInt("JSON_METRICS_REPORT_INTERVAL", 900),      // 15 minutes in seconds
		},

		// Schema definitions - developers can add custom schemas here
		"schemas": map[string]interface{}{
			// Standard API Success Response
			"api_success": map[string]interface{}{
				"name":        "API Success Response",
				"description": "Standard successful API response",
				"type":        "response",
				"template":    `{"success":true,"message":"{{.message}}","data":{{.data}}}`,
				"fields": map[string]interface{}{
					"success": map[string]interface{}{
						"type":          "boolean",
						"required":      true,
						"default_value": true,
					},
					"message": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"data": map[string]interface{}{
						"type":     "object",
						"required": false,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_API_SUCCESS_ENABLED", true),
			},

			// Standard API Error Response
			"api_error": map[string]interface{}{
				"name":        "API Error Response",
				"description": "Standard error API response",
				"type":        "response",
				"template":    `{"success":false,"message":"{{.message}}","errors":{{.errors}}}`,
				"fields": map[string]interface{}{
					"success": map[string]interface{}{
						"type":          "boolean",
						"required":      true,
						"default_value": false,
					},
					"message": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"errors": map[string]interface{}{
						"type":     "object",
						"required": false,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_API_ERROR_ENABLED", true),
			},

			// Validation Error Response
			"validation_error": map[string]interface{}{
				"name":        "Validation Error Response",
				"description": "Validation error response",
				"type":        "response",
				"template":    `{"message":"Validation failed","errors":{{.errors}}}`,
				"fields": map[string]interface{}{
					"message": map[string]interface{}{
						"type":          "string",
						"required":      true,
						"default_value": "Validation failed",
					},
					"errors": map[string]interface{}{
						"type":     "object",
						"required": true,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_VALIDATION_ERROR_ENABLED", true),
			},

			// Health Check Response
			"health_check": map[string]interface{}{
				"name":        "Health Check Response",
				"description": "Health check endpoint response",
				"type":        "response",
				"template":    `{"status":"{{.status}}","timestamp":{{.timestamp}},"service":"{{.service}}","server":"{{.server}}"}`,
				"fields": map[string]interface{}{
					"status": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"timestamp": map[string]interface{}{
						"type":     "number",
						"required": true,
						"format":   "timestamp",
					},
					"service": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"server": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_HEALTH_CHECK_ENABLED", true),
			},

			// Login Success Response
			"login_success": map[string]interface{}{
				"name":        "Login Success Response",
				"description": "Successful login response",
				"type":        "response",
				"template":    `{"message":"{{.message}}","user":{{.user}},"token":"{{.token}}"}`,
				"fields": map[string]interface{}{
					"message": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"user": map[string]interface{}{
						"type":     "object",
						"required": true,
					},
					"token": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_LOGIN_SUCCESS_ENABLED", true),
			},

			// User Response
			"user_response": map[string]interface{}{
				"name":        "User Response",
				"description": "User-related response",
				"type":        "response",
				"template":    `{"message":"{{.message}}","user":{{.user}}}`,
				"fields": map[string]interface{}{
					"message": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"user": map[string]interface{}{
						"type":     "object",
						"required": true,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_USER_RESPONSE_ENABLED", true),
			},

			// Created Response
			"created_response": map[string]interface{}{
				"name":        "Created Response",
				"description": "Resource created response",
				"type":        "response",
				"template":    `{"success":true,"message":"Resource created successfully","data":{{.data}}}`,
				"fields": map[string]interface{}{
					"success": map[string]interface{}{
						"type":          "boolean",
						"required":      true,
						"default_value": true,
					},
					"message": map[string]interface{}{
						"type":          "string",
						"required":      true,
						"default_value": "Resource created successfully",
					},
					"data": map[string]interface{}{
						"type":     "object",
						"required": true,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_CREATED_RESPONSE_ENABLED", true),
			},

			// Updated Response
			"updated_response": map[string]interface{}{
				"name":        "Updated Response",
				"description": "Resource updated response",
				"type":        "response",
				"template":    `{"success":true,"message":"Resource updated successfully","data":{{.data}}}`,
				"fields": map[string]interface{}{
					"success": map[string]interface{}{
						"type":          "boolean",
						"required":      true,
						"default_value": true,
					},
					"message": map[string]interface{}{
						"type":          "string",
						"required":      true,
						"default_value": "Resource updated successfully",
					},
					"data": map[string]interface{}{
						"type":     "object",
						"required": true,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_UPDATED_RESPONSE_ENABLED", true),
			},

			// Deleted Response
			"deleted_response": map[string]interface{}{
				"name":        "Deleted Response",
				"description": "Resource deleted response",
				"type":        "response",
				"template":    `{"success":true,"message":"Resource deleted successfully"}`,
				"fields": map[string]interface{}{
					"success": map[string]interface{}{
						"type":          "boolean",
						"required":      true,
						"default_value": true,
					},
					"message": map[string]interface{}{
						"type":          "string",
						"required":      true,
						"default_value": "Resource deleted successfully",
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_DELETED_RESPONSE_ENABLED", true),
			},

			// Paginated Response
			"paginated_response": map[string]interface{}{
				"name":        "Paginated Response",
				"description": "Paginated collection response",
				"type":        "response",
				"template":    `{"success":true,"message":"{{.message}}","data":{{.data}},"pagination":{{.pagination}}}`,
				"fields": map[string]interface{}{
					"success": map[string]interface{}{
						"type":          "boolean",
						"required":      true,
						"default_value": true,
					},
					"message": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"data": map[string]interface{}{
						"type":     "array",
						"required": true,
					},
					"pagination": map[string]interface{}{
						"type":     "object",
						"required": true,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_PAGINATED_RESPONSE_ENABLED", true),
			},

			// Example Custom Schema (developers can add more like this)
			"custom_notification": map[string]interface{}{
				"name":        "Custom Notification Response",
				"description": "Custom notification response example",
				"type":        "response",
				"template":    `{"type":"{{.type}}","title":"{{.title}}","message":"{{.message}}","read":{{.read}}}`,
				"fields": map[string]interface{}{
					"type": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"title": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"message": map[string]interface{}{
						"type":     "string",
						"required": true,
					},
					"read": map[string]interface{}{
						"type":          "boolean",
						"required":      true,
						"default_value": false,
					},
				},
				"enabled": env.GetBool("JSON_SCHEMA_CUSTOM_NOTIFICATION_ENABLED", false), // Disabled by default
			},
		},
	}
}

// init automatically registers this config with the global config loader
// This ensures the config is available via config.Get("json_optimizations") and dot notation
func init() {
	go_core.RegisterGlobalConfig("json_optimizations", JSONOptimizationsConfig)
}