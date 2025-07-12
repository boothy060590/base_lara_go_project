# Dynamic Configuration Guide

## Overview

The Laravel-inspired Go framework supports **dynamic configuration changes at runtime**. This means you can modify configuration values while your application is running, without restarting the application.

## Core Concepts

### 1. Config Registration
Configs are registered at startup using functions that return configuration maps:

```go
import (
    go_core "base_lara_go_project/app/core/go_core"
    config_core "base_lara_go_project/app/core/laravel_core/config"
)

// Register a config function
go_core.RegisterGlobalConfig("app", func() map[string]interface{} {
    return map[string]interface{}{
        "name": "MyApp",
        "debug": false,
        "port": 8080,
        "database": map[string]interface{}{
            "host": "localhost",
            "port": 5432,
        },
    }
})
```

### 2. Config Access
Use dot notation to access config values:

```go
// Get values
appName := config_core.GetString("app.name")
debugMode := config_core.GetBool("app.debug")
port := config_core.GetInt("app.port")
dbHost := config_core.GetString("app.database.host")
```

## Dynamic Configuration Operations

### Setting Values at Runtime

Use `Set()` to modify configuration values:

```go
// Set simple values
config_core.Set("app.name", "UpdatedApp")
config_core.Set("app.debug", true)
config_core.Set("app.port", 9000)

// Set nested values
config_core.Set("app.database.host", "newhost.com")
config_core.Set("app.database.port", 3306)

// Add new nested values
config_core.Set("app.database.ssl", true)
config_core.Set("app.database.connection_pool", 20)
```

**Important:** Changes are immediately visible to all subsequent `Get()` calls.

### Reloading Configs

Use `Reload()` to revert to the original values from the config function:

```go
// Reload a specific config
err := config_core.Reload("app")
if err != nil {
    // Handle error
}

// After reload, values revert to original
appName := config_core.GetString("app.name") // Returns "MyApp" (original)
debugMode := config_core.GetBool("app.debug") // Returns false (original)
```

**Note:** Dynamically added values (not in the original config function) will be removed after reload.

### Clearing Cache

Use `ClearCache()` to clear all cached configs:

```go
// Clear all config caches
config_core.ClearCache()

// This forces all configs to be reloaded from their functions on next access
```

**Use Cases:**
- Testing (clear between tests)
- Global config resets
- Force reload of all configs

## Thread Safety

All configuration operations are **thread-safe**:

```go
// Safe to call from multiple goroutines
go func() {
    config_core.Set("app.concurrent_value", "value1")
}()

go func() {
    config_core.Set("app.concurrent_value", "value2")
}()

// Both goroutines can safely read/write configs
value := config_core.GetString("app.concurrent_value")
```

## Best Practices

### 1. Config Organization

```go
// Good: Organize configs by feature
go_core.RegisterGlobalConfig("database", func() map[string]interface{} {
    return map[string]interface{}{
        "host": "localhost",
        "port": 5432,
        "ssl": false,
    }
})

go_core.RegisterGlobalConfig("cache", func() map[string]interface{} {
    return map[string]interface{}{
        "driver": "redis",
        "host": "localhost",
        "port": 6379,
    }
})
```

### 2. Dynamic Updates

```go
// Update config based on user input
func updateDatabaseConfig(newHost string, newPort int) {
    config_core.Set("database.host", newHost)
    config_core.Set("database.port", newPort)
    
    // Optionally reload database connections
    // This depends on your application logic
}

// Update config based on environment
func updateConfigForEnvironment(env string) {
    if env == "production" {
        config_core.Set("app.debug", false)
        config_core.Set("database.ssl", true)
    } else {
        config_core.Set("app.debug", true)
        config_core.Set("database.ssl", false)
    }
}
```

### 3. Hot-Reloading

```go
// Example: Reload configs when files change
func watchConfigFiles() {
    // Watch for config file changes
    for {
        select {
        case <-fileChangeEvent:
            // Reload specific configs
            config_core.Reload("app")
            config_core.Reload("database")
            
            // Or reload all configs
            config_core.ClearCache()
        }
    }
}
```

### 4. Testing

```go
func TestConfigChanges(t *testing.T) {
    // Always clear configs and cache in tests
    go_core.ClearGlobalConfigs()
    config_core.ClearCache()
    
    // Register test config
    go_core.RegisterGlobalConfig("test", func() map[string]interface{} {
        return map[string]interface{}{
            "value": "initial",
        }
    })
    
    // Test initial value
    require.Equal(t, "initial", config_core.GetString("test.value"))
    
    // Test dynamic change
    config_core.Set("test.value", "updated")
    require.Equal(t, "updated", config_core.GetString("test.value"))
    
    // Test reload
    config_core.Reload("test")
    require.Equal(t, "initial", config_core.GetString("test.value"))
}
```

## Performance Considerations

### Caching
- Configs are cached after first access for performance
- `Set()` operations modify the cached values
- `Reload()` clears the cache and reloads from the function
- `ClearCache()` clears all caches

### Memory Usage
- Each config is cached in memory
- Large configs consume more memory
- Use `ClearCache()` to free memory if needed

### Concurrency
- All operations are thread-safe
- No performance penalty for concurrent access
- Use mutexes internally for safety

## Error Handling

```go
// Reload can return errors
err := config_core.Reload("nonexistent_config")
if err != nil {
    // Handle error (config doesn't exist)
}

// Get operations return defaults for missing values
value := config_core.GetString("missing.key", "default")
```

## Migration from Static Configs

If you're migrating from static configs:

```go
// Before: Static config
type Config struct {
    AppName string
    Debug   bool
}

// After: Dynamic config
go_core.RegisterGlobalConfig("app", func() map[string]interface{} {
    return map[string]interface{}{
        "name": "MyApp",
        "debug": false,
    }
})

// Access with dot notation
appName := config_core.GetString("app.name")
debugMode := config_core.GetBool("app.debug")
```

## Troubleshooting

### Common Issues

1. **Values not updating after Set()**
   - Ensure the config is loaded/cached first
   - Check that you're using the correct key path

2. **Reload() not working**
   - Verify the config name exists
   - Check that the config function is properly registered

3. **Thread safety issues**
   - All operations are thread-safe by default
   - No additional synchronization needed

4. **Memory leaks**
   - Use `ClearCache()` periodically if needed
   - Monitor memory usage with large configs

### Debug Output

Enable debug output to troubleshoot:

```go
// Debug output is automatically enabled in tests
// Shows cache hits/misses and value types
```

## Summary

Dynamic configuration provides:

- ✅ **Runtime flexibility**: Change configs without restarts
- ✅ **Thread safety**: Safe concurrent access
- ✅ **Performance**: Cached access with fast updates
- ✅ **Reliability**: Fallback to original values with reload
- ✅ **Developer-friendly**: Simple dot notation access

Use `Set()` for runtime changes, `Reload()` to revert, and `ClearCache()` for global resets. 