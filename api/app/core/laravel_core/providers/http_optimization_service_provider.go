package providers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"log"
	"time"
)

// HTTPOptimizationServiceProvider provides HTTP optimization services
type HTTPOptimizationServiceProvider struct {
	BaseServiceProvider
}

// Register registers the HTTP optimization service provider
func (p *HTTPOptimizationServiceProvider) Register(container *app_core.Container) error {
	// Register HTTP optimizer as singleton
	container.Singleton("http.optimizer", func() (any, error) {
		// Get configuration
		config := app_core.DefaultHTTPOptimizationConfig()
		
		// Override with custom configuration if available
		if configInstance, err := container.Resolve("config"); err == nil {
			if configMap, ok := configInstance.(map[string]any); ok {
				p.applyConfiguration(config, configMap)
			}
		}

		optimizer := app_core.NewHTTPOptimizer(config)
		
		log.Printf("HTTP Optimizer registered with fasthttp enabled: %v", config.EnableFastHTTP)
		return optimizer, nil
	})

	// Register HTTP optimization configuration
	container.Singleton("http.optimization.config", func() (any, error) {
		config := app_core.DefaultHTTPOptimizationConfig()
		
		// Override with environment-specific settings
		if configInstance, err := container.Resolve("config"); err == nil {
			if configMap, ok := configInstance.(map[string]any); ok {
				p.applyConfiguration(config, configMap)
			}
		}
		
		return config, nil
	})

	log.Printf("HTTP Optimization service provider registered successfully")
	return nil
}

// Boot boots the HTTP optimization service provider
func (p *HTTPOptimizationServiceProvider) Boot(container *app_core.Container) error {
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

	// Monitoring settings
	if monConfig, ok := configMap["http_monitoring"].(map[string]any); ok {
		if enableMetrics, ok := monConfig["enable_metrics"].(bool); ok {
			config.EnableMetrics = enableMetrics
		}
		if metricsPath, ok := monConfig["metrics_path"].(string); ok {
			config.MetricsPath = metricsPath
		}
	}

	log.Printf("Applied HTTP optimization configuration: FastHTTP=%v, MaxConn=%d", 
		config.EnableFastHTTP, config.MaxConnections)
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