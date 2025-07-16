# Configuration System

This project uses a Laravel-style configuration system with Go config files that provide fallback values and easy developer control.

## Overview

The configuration system consists of:

1. **Go Config Files** - Simple functions returning maps with fallback values
2. **Config Facade** - Laravel-style `config()` helper functionality
3. **Environment Integration** - Automatic environment variable fallbacks
4. **Automatic Discovery** - Config files automatically registered by filename

## Config Files

### Available Config Files

- `api/config/app.go` - Application configuration (name, debug, port, etc.)
- `api/config/database.go` - Database connections and settings
- `api/config/cache.go` - Cache store configuration
- `api/config/http.go` - FastHTTP optimization settings
- `api/config/queue.go` - Queue and worker configuration
- `api/config/mail.go` - Mail configuration
- `api/config/logging.go` - Logging channels and settings
- `api/config/goroutine.go` - Goroutine pool and optimization settings
- `api/config/context.go` - Context optimization settings
- `api/config/custom_allocators.go` - Memory allocation settings
- `api/config/profile_guided.go` - Profile-guided optimization settings
- `api/config/work_stealing.go` - Work stealing pool settings
- `api/config/repository_optimizations.go` - Repository optimization settings
- `api/config/infrastructure_optimizations.go` - Infrastructure optimization settings

### Example Config File

```go
// api/config/app.go
package config

import "base_lara_go_project/app/core/env"

func AppConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":                env.GetEnv("APP_NAME", "Base Laravel Go Project"),
		"debug":               env.GetEnv("APP_DEBUG", "false"),
		"url":                 env.GetEnv("APP_URL", "http://localhost"),
		"env":                 env.GetEnv("APP_ENV", "development"),
		"port":                env.GetEnv("APP_PORT", "8080"),
		"secret":              env.GetEnv("API_SECRET", "changeme"),
		"token_hour_lifespan": env.GetEnv("TOKEN_HOUR_LIFESPAN", "1"),
	}
}

// init automatically registers this config with the global config loader
func init() {
	go_core.RegisterGlobalConfig("app", AppConfig)
}
```

## Usage

### Using the Config Facade

```go
import "base_lara_go_project/app/core/facades"

// Get app name
appName := facades.GetString("app.name")

// Get database host with fallback
dbHost := facades.GetString("database.connections.mysql.host", "localhost")

// Get HTTP optimization settings
maxConnections := facades.GetInt("http.max_connections", 10000)
enableFastHTTP := facades.GetBool("http.enable_fasthttp", true)

// Check if config exists
if facades.Has("mail.from.address") {
    fromAddress := facades.GetString("mail.from.address")
}

// Get boolean values
debugMode := facades.GetBool("app.debug", false)
```

### In Laravel-Style Facades

```go
import facades_core "base_lara_go_project/app/core/laravel_core/facades"

// Using the config facade
config := facades_core.Config()
appName := config.GetString("app.name")

// Or use the global functions
appName := facades_core.GetString("app.name")
debugMode := facades_core.GetBool("app.debug")
maxConnections := facades_core.GetInt("http.max_connections")
```

### In Service Providers

```go
import "base_lara_go_project/config"

func (p *MyServiceProvider) Register(container *app_core.Container) error {
    // Get config values
    myConfig := config.Get("my").(map[string]interface{})
    enabled := myConfig["enabled"].(bool)
    
    // Register services based on config
    if enabled {
        container.Singleton("my.service", func() (any, error) {
            return NewMyService(myConfig), nil
        })
    }
    
    return nil
}
```

## HTTP Configuration

The framework includes comprehensive HTTP optimization settings:

```go
// api/config/http.go
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
			"max_connections":     15000,
			"max_idle_connections": 1500,
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
```

## Database Configuration

Database configuration with connection pooling and optimization settings:

```go
// api/config/database.go
func DatabaseConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.GetEnv("DB_CONNECTION", "mysql"),
		"connections": map[string]interface{}{
			"mysql": map[string]interface{}{
				"driver":   "mysql",
				"host":     env.GetEnv("DB_HOST", "localhost"),
				"port":     env.GetEnv("DB_PORT", "3306"),
				"database": env.GetEnv("DB_DATABASE", "laravel_go"),
				"username": env.GetEnv("DB_USERNAME", "root"),
				"password": env.GetEnv("DB_PASSWORD", ""),
				"charset":  "utf8mb4",
				"collation": "utf8mb4_unicode_ci",
				"prefix":   "",
				"strict":   true,
				"engine":   "",
				// Connection pooling settings
				"max_open_conns": env.GetEnvInt("DB_MAX_OPEN_CONNS", 25),
				"max_idle_conns": env.GetEnvInt("DB_MAX_IDLE_CONNS", 5),
				"conn_max_lifetime": env.GetEnvInt("DB_CONN_MAX_LIFETIME", 300),
			},
		},
	}
}
```

## Cache Configuration

Cache configuration with multiple store options:

```go
// api/config/cache.go
func CacheConfig() map[string]interface{} {
	return map[string]interface{}{
		"default": env.Get("CACHE_STORE", "local"),
		"stores": map[string]interface{}{
			"local": map[string]interface{}{
				"driver": "local",
				"ttl":    env.GetInt("CACHE_TTL", 3600),
			},
			"redis": map[string]interface{}{
				"driver":     "redis",
				"host":       env.Get("REDIS_HOST", "127.0.0.1"),
				"port":       env.Get("REDIS_PORT", "6379"),
				"password":   env.Get("REDIS_PASSWORD", ""),
				"database":   env.GetInt("REDIS_DB", 0),
				"prefix":     env.Get("CACHE_PREFIX", "laravel_cache"),
				"connection": env.Get("REDIS_CONNECTION", "default"),
			},
		},
	}
}
```

## Environment Variables

All config files use the `env` package for environment variable integration:

```go
// Get string with default
value := env.Get("MY_VAR", "default")

// Get integer with default
timeout := env.GetInt("MY_TIMEOUT", 30)

// Get boolean with default
enabled := env.GetBool("MY_ENABLED", true)
```

## Configuration Profiles

The system supports different configuration profiles for different use cases:

- **Web Apps**: Low latency, fast response times (30s timeouts)
- **APIs**: Moderate timeouts, high throughput (60s timeouts)
- **Background Jobs**: Long timeouts, high performance (300s timeouts)
- **Streaming**: Very long timeouts, large buffers (1800s timeouts)
- **Batch Processing**: Long timeouts, large buffers (1800s timeouts)

## Automatic Discovery

Config files are automatically discovered and registered:

```go
// Config files in api/config/ are automatically registered
// The filename becomes the config key
// app.go -> config.Get("app")
// database.go -> config.Get("database")
// http.go -> config.Get("http")

// No manual registration required
func init() {
    go_core.RegisterGlobalConfig("app", AppConfig)
}
```

## Service Provider Integration

Service providers automatically load and apply configuration:

```go
// HTTPOptimizationServiceProvider automatically applies HTTP config
func (p *HTTPOptimizationServiceProvider) Boot(container *app_core.Container) error {
    // Get config facade
    configFacade := facades_core.Config()
    
    // Apply configuration from config service
    if httpConfig := configFacade.Get("http"); httpConfig != nil {
        if configMap, ok := httpConfig.(map[string]interface{}); ok {
            p.applyConfiguration(newConfig, configMap)
        }
    }
    
    return nil
}
```

## Testing Configuration

For testing, you can use different configuration profiles:

```go
// Test-specific configuration
func TestConfig() map[string]interface{} {
    return map[string]interface{}{
        "database": map[string]interface{}{
            "connections": map[string]interface{}{
                "mysql": map[string]interface{}{
                    "database": "test_db",
                    "host":     "localhost",
                    "port":     "3306",
                },
            },
        },
        "cache": map[string]interface{}{
            "default": "local",
        },
    }
}
```

## Best Practices

### 1. Use Environment Variables
Always provide environment variable fallbacks:
```go
"host": env.GetEnv("DB_HOST", "localhost"),
"port": env.GetEnv("DB_PORT", "3306"),
```

### 2. Provide Sensible Defaults
Give reasonable defaults for all configuration values:
```go
"max_connections": env.GetEnvInt("HTTP_MAX_CONNECTIONS", 10000),
"timeout": env.GetEnvInt("HTTP_TIMEOUT", 30),
```

### 3. Use Type-Safe Access
Use the appropriate getter methods:
```go
// ✅ Correct
maxConn := facades.GetInt("http.max_connections", 10000)
enabled := facades.GetBool("http.enable_fasthttp", true)

// ❌ Avoid
maxConn := facades.Get("http.max_connections").(int)
```

### 4. Group Related Settings
Organize related settings together:
```go
"http": map[string]any{
    "enable_fasthttp": true,
    "read_timeout":    30,
    "write_timeout":   30,
},
"http_pool": map[string]any{
    "max_connections": 15000,
    "max_idle_connections": 1500,
},
```

### 5. Document Configuration Options
Include comments for complex configuration options:
```go
"zero_copy_enabled": true, // Enable zero-copy operations for better performance
"pre_allocated_buffers": 100, // Number of pre-allocated buffers for HTTP operations
``` 