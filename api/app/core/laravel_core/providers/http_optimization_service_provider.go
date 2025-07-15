package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"fmt"
	"log"
	"time"
)

// HTTPOptimizationServiceProvider provides HTTP optimization services
type HTTPOptimizationServiceProvider struct {
	BaseServiceProvider
}

// Register registers the HTTP optimization service provider
func (p *HTTPOptimizationServiceProvider) Register(container *app_core.Container) error {
	// Register HTTP optimizer as singleton with default configuration
	container.Singleton("http.optimizer", func() (any, error) {
		config := app_core.DefaultHTTPOptimizationConfig()
		optimizer := app_core.NewHTTPOptimizer(config)
		
		log.Printf("HTTP Optimizer registered with fasthttp enabled: %v", config.EnableFastHTTP)
		return optimizer, nil
	})

	// Register HTTP optimization configuration  
	container.Singleton("http.optimization.config", func() (any, error) {
		config := app_core.DefaultHTTPOptimizationConfig()
		return config, nil
	})

	log.Printf("HTTP Optimization service provider registered successfully")
	return nil
}

// Boot boots the HTTP optimization service provider
func (p *HTTPOptimizationServiceProvider) Boot(container *app_core.Container) error {
	// Resolve the HTTP optimizer
	optimizerInstance, err := container.Resolve("http.optimizer")
	if err != nil {
		return fmt.Errorf("failed to resolve http.optimizer: %w", err)
	}
	
	optimizer, ok := optimizerInstance.(*app_core.HTTPOptimizer)
	if !ok {
		return fmt.Errorf("http.optimizer is not of type *HTTPOptimizer")
	}

	// Try to resolve and apply custom configuration
	if configInstance, err := container.Resolve("config"); err == nil {
		if configFacade, ok := configInstance.(interface{
			Get(key string, defaultValue ...interface{}) interface{}
		}); ok {
			// Get the current config
			currentConfig := optimizer.GetConfig()
			
			// Create a new config based on current defaults
			newConfig := &app_core.HTTPOptimizationConfig{
				EnableFastHTTP:         currentConfig.EnableFastHTTP,
				ReadTimeout:           currentConfig.ReadTimeout,
				WriteTimeout:          currentConfig.WriteTimeout,
				IdleTimeout:           currentConfig.IdleTimeout,
				MaxRequestBodySize:    currentConfig.MaxRequestBodySize,
				MaxConnections:        currentConfig.MaxConnections,
				MaxIdleConnections:    currentConfig.MaxIdleConnections,
				ConnectionTimeout:     currentConfig.ConnectionTimeout,
				EnableCompression:     currentConfig.EnableCompression,
				EnableKeepAlive:      currentConfig.EnableKeepAlive,
				PreAllocatedBuffers:  currentConfig.PreAllocatedBuffers,
				ZeroCopyEnabled:      currentConfig.ZeroCopyEnabled,
				// CORS settings
				CORSAllowedOrigins:   currentConfig.CORSAllowedOrigins,
				CORSAllowedMethods:   currentConfig.CORSAllowedMethods,
				CORSAllowedHeaders:   currentConfig.CORSAllowedHeaders,
				CORSExposedHeaders:   currentConfig.CORSExposedHeaders,
				CORSAllowCredentials: currentConfig.CORSAllowCredentials,
				CORSMaxAge:          currentConfig.CORSMaxAge,
				EnableMetrics:        currentConfig.EnableMetrics,
				MetricsPath:          currentConfig.MetricsPath,
			}
			
			// Apply configuration from config service
			// The http config is registered as a top-level config named "http"
			if httpConfig := configFacade.Get("http"); httpConfig != nil {
				if configMap, ok := httpConfig.(map[string]interface{}); ok {
					p.applyConfiguration(newConfig, configMap)
					log.Printf("Applied custom HTTP configuration from config service")
				}
			}
			
			// Update the optimizer with the new configuration
			if err := optimizer.UpdateConfig(newConfig); err != nil {
				log.Printf("Warning: Failed to update HTTP optimizer config: %v", err)
			} else {
				log.Printf("HTTP Optimizer configuration updated successfully")
			}
		}
	}

	log.Printf("HTTP Optimization service provider booted successfully")
	return nil
}

// Provides returns the services this provider provides
func (p *HTTPOptimizationServiceProvider) Provides() []string {
	return []string{"http.optimizer", "http.optimization.config"}
}

// When returns the conditions when this provider should be loaded
func (p *HTTPOptimizationServiceProvider) When() []string {
	return []string{} // Always load
}

// applyConfiguration applies configuration from config map to HTTP optimization config
func (p *HTTPOptimizationServiceProvider) applyConfiguration(config *app_core.HTTPOptimizationConfig, configMap map[string]any) {
	// HTTP server settings
	if httpConfig, ok := configMap["http"].(map[string]any); ok {
		if enableFastHTTP, ok := httpConfig["enable_fasthttp"].(bool); ok {
			config.EnableFastHTTP = enableFastHTTP
		}
		if readTimeout, ok := httpConfig["read_timeout"].(int); ok {
			config.ReadTimeout = getDuration(readTimeout, "seconds")
		}
		if writeTimeout, ok := httpConfig["write_timeout"].(int); ok {
			config.WriteTimeout = getDuration(writeTimeout, "seconds")
		}
		if idleTimeout, ok := httpConfig["idle_timeout"].(int); ok {
			config.IdleTimeout = getDuration(idleTimeout, "seconds")
		}
		if maxBodySize, ok := httpConfig["max_request_body_size"].(int); ok {
			config.MaxRequestBodySize = maxBodySize
		}
	}

	// Connection pooling settings
	if poolConfig, ok := configMap["http_pool"].(map[string]any); ok {
		if maxConn, ok := poolConfig["max_connections"].(int); ok {
			config.MaxConnections = maxConn
		}
		if maxIdle, ok := poolConfig["max_idle_connections"].(int); ok {
			config.MaxIdleConnections = maxIdle
		}
		if connTimeout, ok := poolConfig["connection_timeout"].(int); ok {
			config.ConnectionTimeout = getDuration(connTimeout, "seconds")
		}
	}

	// Performance optimizations
	if perfConfig, ok := configMap["http_performance"].(map[string]any); ok {
		if enableCompression, ok := perfConfig["enable_compression"].(bool); ok {
			config.EnableCompression = enableCompression
		}
		if enableKeepAlive, ok := perfConfig["enable_keep_alive"].(bool); ok {
			config.EnableKeepAlive = enableKeepAlive
		}
		if preAllocBuffers, ok := perfConfig["pre_allocated_buffers"].(int); ok {
			config.PreAllocatedBuffers = preAllocBuffers
		}
		if zeroCopy, ok := perfConfig["zero_copy_enabled"].(bool); ok {
			config.ZeroCopyEnabled = zeroCopy
		}
	}

	// CORS settings
	if corsConfig, ok := configMap["http_cors"].(map[string]any); ok {
		if allowedOrigins, ok := corsConfig["allowed_origins"].([]string); ok {
			config.CORSAllowedOrigins = allowedOrigins
		} else if allowedOrigins, ok := corsConfig["allowed_origins"].([]interface{}); ok {
			// Handle []interface{} conversion
			origins := make([]string, len(allowedOrigins))
			for i, origin := range allowedOrigins {
				if str, ok := origin.(string); ok {
					origins[i] = str
				}
			}
			config.CORSAllowedOrigins = origins
		}
		
		if allowedMethods, ok := corsConfig["allowed_methods"].([]string); ok {
			config.CORSAllowedMethods = allowedMethods
		} else if allowedMethods, ok := corsConfig["allowed_methods"].([]interface{}); ok {
			// Handle []interface{} conversion
			methods := make([]string, len(allowedMethods))
			for i, method := range allowedMethods {
				if str, ok := method.(string); ok {
					methods[i] = str
				}
			}
			config.CORSAllowedMethods = methods
		}
		
		if allowedHeaders, ok := corsConfig["allowed_headers"].([]string); ok {
			config.CORSAllowedHeaders = allowedHeaders
		} else if allowedHeaders, ok := corsConfig["allowed_headers"].([]interface{}); ok {
			// Handle []interface{} conversion
			headers := make([]string, len(allowedHeaders))
			for i, header := range allowedHeaders {
				if str, ok := header.(string); ok {
					headers[i] = str
				}
			}
			config.CORSAllowedHeaders = headers
		}
		
		if exposedHeaders, ok := corsConfig["exposed_headers"].([]string); ok {
			config.CORSExposedHeaders = exposedHeaders
		} else if exposedHeaders, ok := corsConfig["exposed_headers"].([]interface{}); ok {
			// Handle []interface{} conversion
			headers := make([]string, len(exposedHeaders))
			for i, header := range exposedHeaders {
				if str, ok := header.(string); ok {
					headers[i] = str
				}
			}
			config.CORSExposedHeaders = headers
		}
		
		if allowCredentials, ok := corsConfig["allow_credentials"].(bool); ok {
			config.CORSAllowCredentials = allowCredentials
		}
		
		if maxAge, ok := corsConfig["max_age"].(int); ok {
			config.CORSMaxAge = maxAge
		}
	}

	// Monitoring settings
	if monConfig, ok := configMap["http_monitoring"].(map[string]any); ok {
		if enableMetrics, ok := monConfig["enable_metrics"].(bool); ok {
			config.EnableMetrics = enableMetrics
		}
		if metricsPath, ok := monConfig["metrics_path"].(string); ok {
			config.MetricsPath = metricsPath
		}
	}

	log.Printf("Applied HTTP optimization configuration: FastHTTP=%v, MaxConn=%d, CORS Origins=%v", 
		config.EnableFastHTTP, config.MaxConnections, config.CORSAllowedOrigins)
}

// getDuration converts an integer value to time.Duration with the specified unit
func getDuration(value int, unit string) time.Duration {
	duration := time.Duration(value)
	switch unit {
	case "seconds":
		return duration * time.Second
	case "minutes":
		return duration * time.Minute
	case "hours":
		return duration * time.Hour
	case "milliseconds":
		return duration * time.Millisecond
	default:
		return duration * time.Second // Default to seconds
	}
}