# Dynamic Configuration System

This directory contains the dynamic configuration system for the Laravel-inspired Go framework. The system provides zero-configuration, automatic loading of configuration files with Laravel-style dot notation access.

## Overview

The configuration system automatically discovers and loads all Go files in this directory. Each file should export a function that returns a `map[string]interface{}` and register itself with the global config loader using `init()` functions.

## How It Works

1. **Automatic Discovery**: All `.go` files in this directory are automatically loaded
2. **Zero Configuration**: No manual registration or mapping required
3. **Laravel-Style Access**: Use dot notation to access config values
4. **Type-Safe Access**: Helper methods for different data types
5. **Environment Integration**: Built-in environment variable fallbacks

## File Structure

Each config file should follow this pattern:

```go
package config

import (
    "base_lara_go_project/app/core/go_core"
    "base_lara_go_project/app/core/laravel_core/env"
)

// MyConfig returns the configuration for my feature
func MyConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled": env.GetBool("MY_FEATURE_ENABLED", true),
        "timeout": env.GetInt("MY_FEATURE_TIMEOUT", 30),
        "options": map[string]interface{}{
            "option1": env.Get("MY_OPTION_1", "default"),
            "option2": env.GetInt("MY_OPTION_2", 100),
        },
    }
}

// init automatically registers this config with the global config loader
func init() {
    go_core.RegisterGlobalConfig("my", MyConfig)
}
```

## Usage

### In Application Code

```go
import "base_lara_go_project/config"

// Get entire config
appConfig := config.Get("app")

// Get specific values with dot notation
appName := config.GetString("app.name")
debugMode := config.GetBool("app.debug")
port := config.GetInt("app.port")

// Get with defaults
timeout := config.GetInt("my.timeout", 30)
enabled := config.GetBool("my.enabled", true)

// Check if config exists
if config.Has("my.feature") {
    // Use the config
}
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

## Available Config Files

- **app.go**: Application configuration (name, debug, port, etc.)
- **cache.go**: Cache configuration (stores, TTL, etc.)
- **context.go**: Context optimization settings
- **database.go**: Database connection settings
- **goroutine.go**: Goroutine pool and optimization settings
- **logging.go**: Logging channels and handlers
- **mail.go**: Mail configuration
- **queue.go**: Queue system settings

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

## Performance Features

- **Automatic Caching**: Config values are cached after first load
- **Lazy Loading**: Configs are only loaded when accessed
- **Thread-Safe**: All operations are concurrency-safe
- **Memory Efficient**: Minimal memory footprint with smart caching

## Adding New Config Files

1. Create a new `.go` file in this directory
2. Define a function that returns `map[string]interface{}`
3. Add an `init()` function that calls `go_core.RegisterGlobalConfig()`
4. Use the `env` package for environment variable integration
5. The config will be automatically available via dot notation

## Best Practices

1. **Use Descriptive Names**: Config function names should be clear and descriptive
2. **Provide Sensible Defaults**: Always provide fallback values for environment variables
3. **Group Related Settings**: Use nested maps to group related configuration
4. **Document Environment Variables**: Include comments for required environment variables
5. **Use Type-Safe Access**: Prefer `GetString()`, `GetInt()`, `GetBool()` over `Get()`

## Example: Adding a New Feature Config

```go
package config

import (
    "base_lara_go_project/app/core/go_core"
    "base_lara_go_project/app/core/laravel_core/env"
)

// FeatureConfig returns the feature configuration
func FeatureConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled": env.GetBool("FEATURE_ENABLED", true),
        "api": map[string]interface{}{
            "endpoint": env.Get("FEATURE_API_ENDPOINT", "https://api.example.com"),
            "timeout":  env.GetInt("FEATURE_API_TIMEOUT", 30),
            "retries":  env.GetInt("FEATURE_API_RETRIES", 3),
        },
        "cache": map[string]interface{}{
            "enabled": env.GetBool("FEATURE_CACHE_ENABLED", true),
            "ttl":     env.GetInt("FEATURE_CACHE_TTL", 3600),
        },
    }
}

// init automatically registers this config with the global config loader
func init() {
    go_core.RegisterGlobalConfig("feature", FeatureConfig)
}
```

Access it in your code:

```go
// Check if feature is enabled
if config.GetBool("feature.enabled") {
    // Use the feature
    endpoint := config.GetString("feature.api.endpoint")
    timeout := config.GetInt("feature.api.timeout")
}
``` 