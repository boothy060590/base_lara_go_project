# Configuration System

This project uses a JSON-based configuration system that replaces the old Go-based configuration files. This approach provides better flexibility, environment variable substitution, and easier maintenance.

## Overview

The configuration system consists of:

1. **JSON Configuration Files** - Human-readable configuration files with template variables
2. **Config Core Package** - Go package that loads and parses JSON configs with environment variable substitution
3. **Config Wrappers** - Go packages that provide type-safe access to configuration values

## JSON Configuration Files

### Queue Configuration (`api/config/queue.json`)

The queue configuration defines:
- Default connection (sync, sqs)
- Connection settings for each driver
- Worker configurations with queue assignments
- API queue mappings

Example:
```json
{
  "default_connection": "sqs",
  "connections": {
    "sync": {
      "driver": "sync",
      "queues": ["default"]
    },
    "sqs": {
      "driver": "sqs",
      "key": "{{SQS_ACCESS_KEY}}",
      "secret": "{{SQS_SECRET_KEY}}",
      "region": "{{SQS_REGION}}",
      "endpoint": "http://sqs.{{APP_DOMAIN}}:9324",
      "queues": ["mail", "jobs", "events", "default"]
    }
  },
  "workers": {
    "default": {
      "queues": ["mail", "jobs", "events", "default"],
      "max_jobs": 1000,
      "memory_limit": 128,
      "timeout": 60,
      "sleep": 3,
      "tries": 3
    }
  },
  "api_queues": {
    "mail": "mail",
    "jobs": "jobs",
    "events": "events",
    "default": "default"
  }
}
```

### Template Variables

JSON configs support template variables that are replaced with environment values:

- `{{APP_DOMAIN}}` - Application domain
- `{{SQS_ACCESS_KEY}}` - SQS access key
- `{{SQS_SECRET_KEY}}` - SQS secret key
- `{{SQS_REGION}}` - SQS region

## Config Core Package

The `app/core/config` package provides:

### JSONConfig

```go
type JSONConfig struct {
    data map[string]interface{}
}

// LoadJSONConfig loads a JSON file and substitutes template variables
func LoadJSONConfig(filename string, templateVars map[string]string) (*JSONConfig, error)

// GetWorkerConfig returns configuration for a specific worker
func (c *JSONConfig) GetWorkerConfig(workerName string) map[string]interface{}

// GetWorkerQueues returns the queues assigned to a specific worker
func (c *JSONConfig) GetWorkerQueues(workerName string) []string

// GetConnectionConfig returns configuration for a specific connection
func (c *JSONConfig) GetConnectionConfig(connectionName string) map[string]interface{}

// GetAPIQueues returns the queue mapping for the API
func (c *JSONConfig) GetAPIQueues() map[string]string
```

## Configuration Wrappers

### Queue Config (`api/config/queue.go`)

Provides a clean interface to queue configuration:

```go
// QueueConfig returns the complete queue configuration
func QueueConfig() map[string]interface{}

// GetWorkerConfig returns configuration for a specific worker
func GetWorkerConfig(workerName string) map[string]interface{}

// GetWorkerQueues returns the queues assigned to a specific worker
func GetWorkerQueues(workerName string) []string

// GetDefaultWorkerQueues returns the queues for the default worker
func GetDefaultWorkerQueues() []string

// GetConnectionConfig returns configuration for a specific connection
func GetConnectionConfig(connectionName string) map[string]interface{}

// GetAPIQueues returns the queue mapping for the API
func GetAPIQueues() map[string]string

// GetDefaultConnection returns the default connection name
func GetDefaultConnection() string
```

## Worker Configuration

### Dynamic Worker Generation

The `setup/generate-workers.sh` script allows you to:

1. **Configure Multiple Workers** - Define worker instances with specific queue assignments
2. **Generate Environment Files** - Create worker-specific environment files
3. **Generate Docker Compose Services** - Create Docker Compose services for each worker
4. **Update Queue Configuration** - Automatically update `queue.json` with worker configurations

### Usage

```bash
# Run the worker configuration script
./setup/generate-workers.sh

# Start workers
docker-compose -f docker-compose.workers.yaml up -d

# Start API with workers
docker-compose -f docker-compose.yaml -f docker-compose.workers.yaml up -d
```

### Worker Environment Files

Each worker gets its own environment file (`api/env/.env.{worker_name}`) with:

- Worker-specific configuration
- Queue assignments
- Database and cache settings
- Logging configuration

## Migration from Go Configs

The old Go-based configuration files have been removed:

- ❌ `api/config/queue.go` (old version)
- ❌ `api/scripts/run-workers.sh` (replaced by Docker Compose)

The new system provides:

- ✅ **Better Maintainability** - JSON is easier to read and modify
- ✅ **Environment Variable Support** - Template substitution for dynamic values
- ✅ **Type Safety** - Go wrappers provide compile-time safety
- ✅ **Docker Integration** - Native Docker Compose support
- ✅ **Flexibility** - Easy to add new workers and configurations

## Adding New Configuration Types

To add a new configuration type:

1. **Create JSON File** - Add `api/config/{type}.json`
2. **Add Template Variables** - Use `{{VARIABLE_NAME}}` syntax
3. **Create Config Wrapper** - Add `api/config/{type}.go`
4. **Update Loader** - Add loading logic to the wrapper

Example:
```go
// api/config/cache.go
func CacheConfig() map[string]interface{} {
    return loadCacheConfig().ToMap()
}

func GetCacheStore() string {
    return loadCacheConfig().GetString("default_store")
}
```

## Best Practices

1. **Use Template Variables** - Always use `{{VARIABLE_NAME}}` for environment-dependent values
2. **Provide Defaults** - Include sensible defaults in JSON configs
3. **Type Safety** - Use Go wrappers for type-safe access
4. **Documentation** - Document new configuration options
5. **Testing** - Test configuration loading and template substitution
