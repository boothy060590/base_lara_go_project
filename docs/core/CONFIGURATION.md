# Configuration Reference

## Overview

This document provides a complete reference for all configuration options in the Laravel-Inspired Go Framework. The framework uses environment variables and configuration files to automatically optimize your application based on your use case.

## Environment Variables

### **Core Configuration**

```bash
# Application
APP_NAME="My Application"
APP_ENV=production
APP_DEBUG=false
APP_URL=http://localhost:8080

# Database
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=myapp
DB_USERNAME=root
DB_PASSWORD=

# Cache
CACHE_DRIVER=redis
REDIS_HOST=127.0.0.1
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Queue
QUEUE_CONNECTION=redis
QUEUE_DRIVER=redis

# Mail
MAIL_DRIVER=smtp
MAIL_HOST=smtp.mailtrap.io
MAIL_PORT=2525
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS=noreply@example.com
MAIL_FROM_NAME="${APP_NAME}"
```

### **Goroutine Optimization**

```bash
# Enable/disable goroutine optimization
GOROUTINE_ENABLED=true

# Worker pool configuration
GOROUTINE_MIN_WORKERS=2
GOROUTINE_MAX_WORKERS=10
GOROUTINE_QUEUE_SIZE=1000
GOROUTINE_IDLE_TIMEOUT=60
GOROUTINE_SHUTDOWN_TIMEOUT=30

# High-performance pool settings
GOROUTINE_HP_MIN_WORKERS=5
GOROUTINE_HP_MAX_WORKERS=50
GOROUTINE_HP_QUEUE_SIZE=5000

# Low-latency pool settings
GOROUTINE_LL_MIN_WORKERS=10
GOROUTINE_LL_MAX_WORKERS=100
GOROUTINE_LL_QUEUE_SIZE=10000

# Operation-specific settings
GOROUTINE_REPO_ASYNC=true
GOROUTINE_REPO_PARALLEL_LIMIT=10
GOROUTINE_REPO_TIMEOUT=30

GOROUTINE_JOBS_ASYNC=true
GOROUTINE_JOBS_PARALLEL_LIMIT=20
GOROUTINE_JOBS_TIMEOUT=60

GOROUTINE_EVENTS_ASYNC=true
GOROUTINE_EVENTS_PARALLEL_LIMIT=15
GOROUTINE_EVENTS_TIMEOUT=30
```

### **Context Management**

```bash
# Enable/disable context optimization
CONTEXT_ENABLED=true

# Default timeouts
CONTEXT_DEFAULT_TIMEOUT=30
CONTEXT_MAX_TIMEOUT=300
CONTEXT_ENABLE_DEADLINE=true
CONTEXT_ENABLE_CANCELLATION=true
CONTEXT_PROPAGATE_VALUES=true

# Operation-specific timeouts
CONTEXT_REPO_TIMEOUT=30
CONTEXT_EVENTS_TIMEOUT=30
CONTEXT_JOBS_TIMEOUT=60
CONTEXT_CACHE_TIMEOUT=5
CONTEXT_MAIL_TIMEOUT=30
CONTEXT_HTTP_TIMEOUT=30

# Profile-based timeouts
CONTEXT_PROFILE_WEB_TIMEOUT=30
CONTEXT_PROFILE_API_TIMEOUT=60
CONTEXT_PROFILE_BACKGROUND_TIMEOUT=300
CONTEXT_PROFILE_STREAMING_TIMEOUT=1800
CONTEXT_PROFILE_BATCH_TIMEOUT=1800
```

### **Memory Optimization**

```bash
# Enable/disable memory optimization
MEMORY_OPTIMIZATION_ENABLED=true

# Object pool settings
MEMORY_POOL_SIZE=100
MEMORY_POOL_MAX_SIZE=1000
MEMORY_POOL_CLEANUP_INTERVAL=300

# Custom allocator settings
MEMORY_ALLOCATOR_ENABLED=true
MEMORY_SLAB_SIZE=1024
MEMORY_SLAB_MAX_SIZE=1048576
MEMORY_ALLOCATION_STRATEGY=pool

# JSON encoding/decoding pools
MEMORY_JSON_ENCODER_POOL_SIZE=50
MEMORY_JSON_DECODER_POOL_SIZE=50
MEMORY_JSON_POOL_CLEANUP_INTERVAL=600
```

### **Work Stealing Pool**

```bash
# Enable/disable work stealing
WORK_STEALING_ENABLED=true

# Pool configuration
WORK_STEALING_NUM_WORKERS=10
WORK_STEALING_QUEUE_SIZE=1000
WORK_STEALING_STEAL_THRESHOLD=5
WORK_STEALING_STEAL_BATCH_SIZE=10
WORK_STEALING_IDLE_TIMEOUT=60

# Optimization settings
WORK_STEALING_ENABLE_METRICS=true
WORK_STEALING_ENABLE_PROFILING=true
WORK_STEALING_ENABLE_AUTO_SCALING=true
```

### **Profile-Guided Optimization**

```bash
# Enable/disable profile-guided optimization
PROFILE_GUIDED_ENABLED=true

# Sampling configuration
PROFILE_GUIDED_SAMPLING_INTERVAL=5
PROFILE_GUIDED_MIN_SAMPLES=100
PROFILE_GUIDED_MAX_SAMPLES=10000

# Optimization configuration
PROFILE_GUIDED_OPTIMIZATION_INTERVAL=60
PROFILE_GUIDED_MAX_OPTIMIZATIONS=10
PROFILE_GUIDED_AUTO_TUNING=true

# Metrics and profiling
PROFILE_GUIDED_ENABLE_METRICS=true
PROFILE_GUIDED_ENABLE_PROFILING=true
```

### **Advanced Channel Patterns**

```bash
# Enable/disable advanced channel patterns
GO_CHANNELS_ENABLED=true

# Default settings
GO_CHANNELS_BUFFER_SIZE=1000
GO_CHANNELS_TIMEOUT=30
GO_CHANNELS_MAX_WORKERS=10

# Pattern-specific settings
GO_CHANNELS_FAN_OUT_ENABLED=true
GO_CHANNELS_FAN_OUT_MAX_WORKERS=20
GO_CHANNELS_FAN_OUT_BUFFER_SIZE=500

GO_CHANNELS_PIPELINE_ENABLED=true
GO_CHANNELS_PIPELINE_MAX_STAGES=10
GO_CHANNELS_PIPELINE_BUFFER_SIZE=500

# Profile-based settings
GO_CHANNELS_HT_BUFFER_SIZE=10000
GO_CHANNELS_HT_MAX_WORKERS=50
GO_CHANNELS_LL_BUFFER_SIZE=100
GO_CHANNELS_LL_MAX_WORKERS=20
```

## Configuration Files

### **1. Goroutine Configuration** (`api/config/goroutine.go`)

```go
package config

import "time"

func GoroutineConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled": getEnvBool("GOROUTINE_ENABLED", true),
        "pools": map[string]interface{}{
            "default": map[string]interface{}{
                "min_workers":     getEnvInt("GOROUTINE_MIN_WORKERS", 2),
                "max_workers":     getEnvInt("GOROUTINE_MAX_WORKERS", 10),
                "queue_size":      getEnvInt("GOROUTINE_QUEUE_SIZE", 1000),
                "idle_timeout":    getEnvDuration("GOROUTINE_IDLE_TIMEOUT", 60*time.Second),
                "shutdown_timeout": getEnvDuration("GOROUTINE_SHUTDOWN_TIMEOUT", 30*time.Second),
            },
            "high_performance": map[string]interface{}{
                "min_workers": getEnvInt("GOROUTINE_HP_MIN_WORKERS", 5),
                "max_workers": getEnvInt("GOROUTINE_HP_MAX_WORKERS", 50),
                "queue_size":  getEnvInt("GOROUTINE_HP_QUEUE_SIZE", 5000),
            },
            "low_latency": map[string]interface{}{
                "min_workers": getEnvInt("GOROUTINE_LL_MIN_WORKERS", 10),
                "max_workers": getEnvInt("GOROUTINE_LL_MAX_WORKERS", 100),
                "queue_size":  getEnvInt("GOROUTINE_LL_QUEUE_SIZE", 10000),
            },
        },
        "operations": map[string]interface{}{
            "repository": map[string]interface{}{
                "async":           getEnvBool("GOROUTINE_REPO_ASYNC", true),
                "parallel_limit":  getEnvInt("GOROUTINE_REPO_PARALLEL_LIMIT", 10),
                "timeout":         getEnvDuration("GOROUTINE_REPO_TIMEOUT", 30*time.Second),
            },
            "jobs": map[string]interface{}{
                "async":           getEnvBool("GOROUTINE_JOBS_ASYNC", true),
                "parallel_limit":  getEnvInt("GOROUTINE_JOBS_PARALLEL_LIMIT", 20),
                "timeout":         getEnvDuration("GOROUTINE_JOBS_TIMEOUT", 60*time.Second),
            },
            "events": map[string]interface{}{
                "async":           getEnvBool("GOROUTINE_EVENTS_ASYNC", true),
                "parallel_limit":  getEnvInt("GOROUTINE_EVENTS_PARALLEL_LIMIT", 15),
                "timeout":         getEnvDuration("GOROUTINE_EVENTS_TIMEOUT", 30*time.Second),
            },
        },
    }
}
```

### **2. Context Configuration** (`api/config/context.go`)

```go
package config

import "time"

func ContextConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled": getEnvBool("CONTEXT_ENABLED", true),
        "defaults": map[string]interface{}{
            "timeout":     getEnvDuration("CONTEXT_DEFAULT_TIMEOUT", 30*time.Second),
            "max_timeout": getEnvDuration("CONTEXT_MAX_TIMEOUT", 5*time.Minute),
        },
        "features": map[string]interface{}{
            "enable_deadline":      getEnvBool("CONTEXT_ENABLE_DEADLINE", true),
            "enable_cancellation":  getEnvBool("CONTEXT_ENABLE_CANCELLATION", true),
            "propagate_values":     getEnvBool("CONTEXT_PROPAGATE_VALUES", true),
        },
        "operations": map[string]interface{}{
            "repository": map[string]interface{}{
                "timeout": getEnvDuration("CONTEXT_REPO_TIMEOUT", 30*time.Second),
            },
            "events": map[string]interface{}{
                "timeout": getEnvDuration("CONTEXT_EVENTS_TIMEOUT", 30*time.Second),
            },
            "jobs": map[string]interface{}{
                "timeout": getEnvDuration("CONTEXT_JOBS_TIMEOUT", 60*time.Second),
            },
            "cache": map[string]interface{}{
                "timeout": getEnvDuration("CONTEXT_CACHE_TIMEOUT", 5*time.Second),
            },
            "mail": map[string]interface{}{
                "timeout": getEnvDuration("CONTEXT_MAIL_TIMEOUT", 30*time.Second),
            },
            "http": map[string]interface{}{
                "timeout": getEnvDuration("CONTEXT_HTTP_TIMEOUT", 30*time.Second),
            },
        },
        "profiles": map[string]interface{}{
            "web": map[string]interface{}{
                "timeout":     getEnvDuration("CONTEXT_PROFILE_WEB_TIMEOUT", 30*time.Second),
                "max_timeout": getEnvDuration("CONTEXT_PROFILE_WEB_MAX_TIMEOUT", 60*time.Second),
            },
            "api": map[string]interface{}{
                "timeout":     getEnvDuration("CONTEXT_PROFILE_API_TIMEOUT", 60*time.Second),
                "max_timeout": getEnvDuration("CONTEXT_PROFILE_API_MAX_TIMEOUT", 300*time.Second),
            },
            "background": map[string]interface{}{
                "timeout":     getEnvDuration("CONTEXT_PROFILE_BACKGROUND_TIMEOUT", 300*time.Second),
                "max_timeout": getEnvDuration("CONTEXT_PROFILE_BACKGROUND_MAX_TIMEOUT", 1800*time.Second),
            },
            "streaming": map[string]interface{}{
                "timeout":     getEnvDuration("CONTEXT_PROFILE_STREAMING_TIMEOUT", 1800*time.Second),
                "max_timeout": getEnvDuration("CONTEXT_PROFILE_STREAMING_MAX_TIMEOUT", 3600*time.Second),
            },
            "batch": map[string]interface{}{
                "timeout":     getEnvDuration("CONTEXT_PROFILE_BATCH_TIMEOUT", 1800*time.Second),
                "max_timeout": getEnvDuration("CONTEXT_PROFILE_BATCH_MAX_TIMEOUT", 7200*time.Second),
            },
        },
    }
}
```

### **3. Work Stealing Configuration** (`api/config/work_stealing.go`)

```go
package config

import "time"

func WorkStealingConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled": getEnvBool("WORK_STEALING_ENABLED", true),
        "workers": map[string]interface{}{
            "num_workers":      getEnvInt("WORK_STEALING_NUM_WORKERS", 10),
            "queue_size":       getEnvInt("WORK_STEALING_QUEUE_SIZE", 1000),
            "steal_threshold":  getEnvInt("WORK_STEALING_STEAL_THRESHOLD", 5),
            "steal_batch_size": getEnvInt("WORK_STEALING_STEAL_BATCH_SIZE", 10),
            "idle_timeout":     getEnvDuration("WORK_STEALING_IDLE_TIMEOUT", 60*time.Second),
        },
        "optimizations": map[string]interface{}{
            "enable_metrics":   getEnvBool("WORK_STEALING_ENABLE_METRICS", true),
            "enable_profiling": getEnvBool("WORK_STEALING_ENABLE_PROFILING", true),
            "enable_auto_scaling": getEnvBool("WORK_STEALING_ENABLE_AUTO_SCALING", true),
        },
    }
}
```

### **4. Profile-Guided Configuration** (`api/config/profile_guided.go`)

```go
package config

import "time"

func ProfileGuidedConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled": getEnvBool("PROFILE_GUIDED_ENABLED", true),
        "sampling": map[string]interface{}{
            "interval":   getEnvDuration("PROFILE_GUIDED_SAMPLING_INTERVAL", 5*time.Second),
            "min_samples": getEnvInt("PROFILE_GUIDED_MIN_SAMPLES", 100),
            "max_samples": getEnvInt("PROFILE_GUIDED_MAX_SAMPLES", 10000),
        },
        "optimization": map[string]interface{}{
            "interval":        getEnvDuration("PROFILE_GUIDED_OPTIMIZATION_INTERVAL", 60*time.Second),
            "max_optimizations": getEnvInt("PROFILE_GUIDED_MAX_OPTIMIZATIONS", 10),
            "auto_tuning":     getEnvBool("PROFILE_GUIDED_AUTO_TUNING", true),
        },
        "optimizations": map[string]interface{}{
            "enable_metrics":   getEnvBool("PROFILE_GUIDED_ENABLE_METRICS", true),
            "enable_profiling": getEnvBool("PROFILE_GUIDED_ENABLE_PROFILING", true),
        },
    }
}
```

### **5. Custom Allocators Configuration** (`api/config/custom_allocators.go`)

```go
package config

import "time"

func CustomAllocatorsConfig() map[string]interface{} {
    return map[string]interface{}{
        "enabled": getEnvBool("MEMORY_OPTIMIZATION_ENABLED", true),
        "pools": map[string]interface{}{
            "size":               getEnvInt("MEMORY_POOL_SIZE", 100),
            "max_size":           getEnvInt("MEMORY_POOL_MAX_SIZE", 1000),
            "cleanup_interval":   getEnvDuration("MEMORY_POOL_CLEANUP_INTERVAL", 300*time.Second),
        },
        "strategies": map[string]interface{}{
            "default": getEnvString("MEMORY_ALLOCATION_STRATEGY", "pool"),
        },
        "allocators": map[string]interface{}{
            "enabled":        getEnvBool("MEMORY_ALLOCATOR_ENABLED", true),
            "slab_size":      getEnvInt("MEMORY_SLAB_SIZE", 1024),
            "slab_max_size":  getEnvInt("MEMORY_SLAB_MAX_SIZE", 1048576),
        },
        "json_pools": map[string]interface{}{
            "encoder_pool_size":      getEnvInt("MEMORY_JSON_ENCODER_POOL_SIZE", 50),
            "decoder_pool_size":      getEnvInt("MEMORY_JSON_DECODER_POOL_SIZE", 50),
            "cleanup_interval":       getEnvDuration("MEMORY_JSON_POOL_CLEANUP_INTERVAL", 600*time.Second),
        },
        "optimizations": map[string]interface{}{
            "enable_metrics":   getEnvBool("MEMORY_ENABLE_METRICS", true),
            "enable_profiling": getEnvBool("MEMORY_ENABLE_PROFILING", true),
        },
    }
}
```

## Profile-Based Configuration

### **Web Application Profile**

```bash
# Optimized for low latency, fast response times
CONTEXT_PROFILE_WEB_TIMEOUT=30
CONTEXT_PROFILE_WEB_MAX_TIMEOUT=60
GOROUTINE_LL_MIN_WORKERS=10
GOROUTINE_LL_MAX_WORKERS=100
GO_CHANNELS_LL_BUFFER_SIZE=100
GO_CHANNELS_LL_MAX_WORKERS=20
MEMORY_POOL_SIZE=50
MEMORY_ALLOCATION_STRATEGY=conservative
```

### **API Service Profile**

```bash
# Optimized for moderate timeouts, high throughput
CONTEXT_PROFILE_API_TIMEOUT=60
CONTEXT_PROFILE_API_MAX_TIMEOUT=300
GOROUTINE_DEFAULT_MIN_WORKERS=2
GOROUTINE_DEFAULT_MAX_WORKERS=10
GO_CHANNELS_BUFFER_SIZE=1000
GO_CHANNELS_MAX_WORKERS=10
MEMORY_POOL_SIZE=100
MEMORY_ALLOCATION_STRATEGY=moderate
```

### **Background Processing Profile**

```bash
# Optimized for long timeouts, high performance
CONTEXT_PROFILE_BACKGROUND_TIMEOUT=300
CONTEXT_PROFILE_BACKGROUND_MAX_TIMEOUT=1800
GOROUTINE_HP_MIN_WORKERS=5
GOROUTINE_HP_MAX_WORKERS=50
GO_CHANNELS_HT_BUFFER_SIZE=10000
GO_CHANNELS_HT_MAX_WORKERS=50
MEMORY_POOL_SIZE=200
MEMORY_ALLOCATION_STRATEGY=aggressive
```

### **Streaming Application Profile**

```bash
# Optimized for very long timeouts, large buffers
CONTEXT_PROFILE_STREAMING_TIMEOUT=1800
CONTEXT_PROFILE_STREAMING_MAX_TIMEOUT=3600
GOROUTINE_HP_MIN_WORKERS=5
GOROUTINE_HP_MAX_WORKERS=50
GO_CHANNELS_HT_BUFFER_SIZE=10000
GO_CHANNELS_HT_MAX_WORKERS=50
MEMORY_POOL_SIZE=500
MEMORY_ALLOCATION_STRATEGY=aggressive
```

### **Batch Processing Profile**

```bash
# Optimized for long timeouts, large buffers
CONTEXT_PROFILE_BATCH_TIMEOUT=1800
CONTEXT_PROFILE_BATCH_MAX_TIMEOUT=7200
GOROUTINE_HP_MIN_WORKERS=5
GOROUTINE_HP_MAX_WORKERS=50
GO_CHANNELS_HT_BUFFER_SIZE=10000
GO_CHANNELS_HT_MAX_WORKERS=50
MEMORY_POOL_SIZE=500
MEMORY_ALLOCATION_STRATEGY=aggressive
```

## Configuration Access

### **Using Facades**

```go
// Get configuration through facades
config := facades.Config()

// Get specific configuration
goroutineConfig := config.Get("goroutine")
contextConfig := config.Get("context")
workStealingConfig := config.Get("work_stealing")

// Get specific values
minWorkers := config.GetInt("goroutine.pools.default.min_workers", 2)
maxWorkers := config.GetInt("goroutine.pools.default.max_workers", 10)
timeout := config.GetDuration("context.defaults.timeout", 30*time.Second)
```

### **Using Service Container**

```go
// Get configuration from container
container := facades.App().Container()

goroutineConfig, err := container.Resolve("config.goroutine")
if err != nil {
    // Handle error
}

contextConfig, err := container.Resolve("config.context")
if err != nil {
    // Handle error
}
```

### **Environment Variable Helpers**

```go
// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
    if value := os.Getenv(key); value != "" {
        if boolValue, err := strconv.ParseBool(value); err == nil {
            return boolValue
        }
    }
    return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

## Configuration Validation

### **Validation Rules**

```go
// Validate configuration
func ValidateConfig(config map[string]interface{}) error {
    // Validate goroutine configuration
    if goroutine, ok := config["goroutine"].(map[string]interface{}); ok {
        if err := validateGoroutineConfig(goroutine); err != nil {
            return fmt.Errorf("goroutine config: %w", err)
        }
    }
    
    // Validate context configuration
    if context, ok := config["context"].(map[string]interface{}); ok {
        if err := validateContextConfig(context); err != nil {
            return fmt.Errorf("context config: %w", err)
        }
    }
    
    return nil
}

func validateGoroutineConfig(config map[string]interface{}) error {
    if pools, ok := config["pools"].(map[string]interface{}); ok {
        if defaultPool, ok := pools["default"].(map[string]interface{}); ok {
            minWorkers := defaultPool["min_workers"].(int)
            maxWorkers := defaultPool["max_workers"].(int)
            
            if minWorkers > maxWorkers {
                return fmt.Errorf("min_workers cannot be greater than max_workers")
            }
        }
    }
    return nil
}
```

## Dynamic Configuration

### **Runtime Configuration Updates**

```go
// Update configuration at runtime
func UpdateConfig(key string, value interface{}) error {
    config := facades.Config()
    
    // Update configuration
    if err := config.Set(key, value); err != nil {
        return err
    }
    
    // Notify services of configuration change
    facades.Event().Dispatch(&ConfigUpdated{Key: key, Value: value})
    
    return nil
}

// Listen for configuration changes
facades.Event().Listen("config.updated", func(event *ConfigUpdated) error {
    // Handle configuration update
    log.Printf("Configuration updated: %s = %v", event.Key, event.Value)
    return nil
})
```

## Next Steps

- [Examples and Tutorials](./EXAMPLES.md) - Real-world examples
- [Developer Guide](./DEVELOPER_GUIDE.md) - How to use the framework
- [Performance Optimizations](./PERFORMANCE_OPTIMIZATIONS.md) - How optimizations work
- [Core Architecture](./CORE_ARCHITECTURE.md) - Detailed architecture 