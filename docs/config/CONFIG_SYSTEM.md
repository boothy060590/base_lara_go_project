# Configuration System

This project uses a Laravel-style configuration system with Go config files that provide fallback values and easy developer control.

## Overview

The configuration system consists of:

1. **Go Config Files** - Simple functions returning maps with fallback values
2. **Config Facade** - Laravel-style `config()` helper functionality
3. **Environment Integration** - Automatic environment variable fallbacks

## Config Files

### Available Config Files

- `api/config/app.go` - Application configuration
- `api/config/database.go` - Database connections and settings
- `api/config/queue.go` - Queue and worker configuration
- `api/config/cache.go` - Cache store configuration
- `api/config/mail.go` - Mail configuration
- `api/config/logging.go` - Logging channels and settings

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
```

## Usage

### Using the Config Facade

```go
import "base_lara_go_project/app/core/facades"

// Get app name
appName := facades.GetString("app.name")

// Get database host with fallback
dbHost := facades.GetString("database.connections.mysql.host", "localhost")

// Get queue worker settings
maxJobs := facades.GetInt("queue.workers.default.max_jobs", 1000)

// Check if config exists
if facades.Has("mail.from.address") {
    fromAddress := facades.GetString("mail.from.address")
}

// Get boolean values
debug := facades.GetBool("app.debug", false)
```

### Direct Config Access

```go
import "base_lara_go_project/config"

// Get entire config
appConfig := config.AppConfig()
databaseConfig := config.DatabaseConfig()

// Access specific values
appName := appConfig["name"].(string)
dbHost := databaseConfig["connections"].(map[string]interface{})["mysql"].(map[string]interface{})["host"].(string)
```

## Environment Variables

All config values support environment variable fallbacks:

```go
// This will use APP_NAME environment variable, or fallback to "Base Laravel Go Project"
appName := env.GetEnv("APP_NAME", "Base Laravel Go Project")
```

### Common Environment Variables

#### App
- `APP_NAME` - Application name
- `APP_DEBUG` - Debug mode (true/false)
- `APP_URL` - Application URL
- `APP_ENV` - Environment (development/staging/production)
- `APP_PORT` - Application port
- `API_SECRET` - API secret key
- `TOKEN_HOUR_LIFESPAN` - Token expiration hours

#### Database
- `DB_CONNECTION` - Default connection (mysql/postgres/sqlite)
- `DB_HOST` - Database host
- `DB_PORT` - Database port
- `DB_NAME` - Database name
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_CHARSET` - Database charset
- `DB_PREFIX` - Table prefix

#### Queue
- `QUEUE_CONNECTION` - Default queue connection (sync/sqs)
- `SQS_ACCESS_KEY` - SQS access key
- `SQS_SECRET_KEY` - SQS secret key
- `SQS_REGION` - SQS region
- `SQS_ENDPOINT` - SQS endpoint URL
- `WORKER_MAX_JOBS` - Maximum jobs per worker
- `WORKER_MEMORY_LIMIT` - Memory limit in MB
- `WORKER_TIMEOUT` - Worker timeout in seconds

#### Cache
- `CACHE_STORE` - Cache store (local/redis)
- `CACHE_PREFIX` - Cache key prefix
- `CACHE_TTL` - Cache TTL in seconds
- `REDIS_HOST` - Redis host
- `REDIS_PORT` - Redis port
- `REDIS_PASSWORD` - Redis password

#### Mail
- `MAIL_MAILER` - Mail driver (smtp/local/mailhog)
- `MAIL_HOST` - SMTP host
- `MAIL_PORT` - SMTP port
- `MAIL_USERNAME` - SMTP username
- `MAIL_PASSWORD` - SMTP password
- `MAIL_FROM_ADDRESS` - From email address
- `MAIL_FROM_NAME` - From name

#### Logging
- `LOG_CHANNEL` - Default log channel
- `LOG_LEVEL` - Log level (debug/info/warning/error)
- `LOG_PATH` - Log file path
- `SENTRY_DSN` - Sentry DSN for error tracking

## Benefits

- ✅ **Developer Control** - Easy to read and modify config files
- ✅ **Environment Fallbacks** - Automatic fallback values for all settings
- ✅ **Type Safety** - Compile-time checking with Go
- ✅ **Laravel Familiarity** - Same pattern as Laravel's PHP configs
- ✅ **No Dependencies** - Pure Go, no external dependencies
- ✅ **Dot Notation** - Easy access with `config.key.subkey` syntax

## Adding New Config Files

1. Create a new Go file in `api/config/`
2. Define a function that returns `map[string]interface{}`
3. Use `env.GetEnv()` for environment variable fallbacks
4. Add the config to the ConfigFacade in `api/app/core/config/facade.go`

Example:
```go
// api/config/custom.go
package config

import "base_lara_go_project/app/core/env"

func CustomConfig() map[string]interface{} {
	return map[string]interface{}{
		"setting1": env.GetEnv("CUSTOM_SETTING1", "default1"),
		"setting2": env.GetEnv("CUSTOM_SETTING2", "default2"),
	}
}
```
