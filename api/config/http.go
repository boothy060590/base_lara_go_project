package config

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
			"max_connections":     10000,
			"max_idle_connections": 1000,
			"connection_timeout":   5, // seconds
		},
		
		// Performance optimizations
		"http_performance": map[string]any{
			"enable_compression":    true,
			"enable_keep_alive":     true,
			"pre_allocated_buffers": 100,
			"zero_copy_enabled":     true,
		},
		
		// Monitoring settings
		"http_monitoring": map[string]any{
			"enable_metrics": true,
			"metrics_path":   "/metrics",
		},
	}
}