# Laravel Core Package

Laravel-inspired utilities and patterns for building applications with familiar Laravel developer experience.

## Overview

This package provides Laravel-style patterns and utilities that work on top of the high-performance `go_core` package. It offers familiar Laravel developer experience while leveraging Go's performance characteristics.

## Architecture

```
laravel_core/
├── config/           # Configuration management
├── facades/          # Laravel-style facades
├── logging/          # Comprehensive logging system
├── exceptions/       # Exception handling
├── env/             # Environment variable management
├── models/          # Base model system
├── dtos/            # Data transfer objects
├── clients/         # Client interfaces
├── observers/       # Model observers
└── README.md        # This file
```

## Key Features

### 🏗️ Laravel-Style Patterns
- **Facades**: Static-like access to services
- **Service Providers**: Dependency injection and service registration
- **Configuration**: Hierarchical configuration management
- **Logging**: Multi-level logging with context support

### 🔧 Developer Experience
- **Familiar APIs**: Laravel-like method names and patterns
- **Intuitive Interfaces**: Easy to understand and use
- **Comprehensive Documentation**: Clear examples and guides
- **IDE Support**: Full autocomplete and type hints

### 🚀 Performance
- **Go-Native**: Built on top of optimized Go core
- **Efficient**: Minimal overhead over pure Go
- **Scalable**: Designed for high-performance applications

## Usage Examples

### Configuration Management
```go
import "your-project/api/app/core/laravel_core/config"

// Load configuration
config.Load("config/app.go")

// Access configuration values
dbHost := config.Get("database.host", "localhost")
debug := config.GetBool("app.debug", false)
```

### Facades
```go
import "your-project/api/app/core/laravel_core/facades"

// Use facades for easy access
user := facades.DB().Table("users").Where("id", 1).First()
facades.Cache().Set("user:1", user, 30*time.Minute)
facades.Log().Info("User retrieved", map[string]interface{}{"user_id": 1})
```

### Logging
```go
import "your-project/api/app/core/laravel_core/logging"

// Create logger
logger := logging.NewLogger()

// Log with context
logger.Info("User logged in", map[string]interface{}{
    "user_id": 123,
    "ip": "192.168.1.1",
})
```

### Exception Handling
```go
import "your-project/api/app/core/laravel_core/exceptions"

// Throw exceptions
if user == nil {
    exceptions.Throw("User not found", 404)
}

// Handle exceptions
defer func() {
    if r := recover(); r != nil {
        exceptions.Handle(r)
    }
}()
```

### Environment Management
```go
import "your-project/api/app/core/laravel_core/env"

// Load environment
env.Load(".env")

// Access environment variables
dbHost := env.Get("DB_HOST", "localhost")
debug := env.GetBool("APP_DEBUG", false)
```

## Integration with Go Core

This package is designed to work seamlessly with the `go_core` package:

```go
// Go Core provides the foundation
userRepo := go_core.NewRepository[User](db)
cache := go_core.NewMemoryCache[User]()

// Laravel Core provides the developer experience
facades.DB().Table("users").Where("active", true).Get()
facades.Cache().Remember("users:active", 30*time.Minute, func() interface{} {
    return userRepo.FindAll()
})
```

## Benefits

1. **Familiarity**: Laravel developers feel at home
2. **Productivity**: Rapid development with familiar patterns
3. **Performance**: Go-native performance under the hood
4. **Maintainability**: Clean, well-structured code
5. **Extensibility**: Easy to customize and extend

## Design Philosophy

### Laravel Patterns, Go Performance
- Use Laravel-style APIs for developer experience
- Implement with Go-optimized code for performance
- Provide type safety where possible
- Maintain Go idioms under the hood

### Separation of Concerns
- **go_core**: High-performance foundation
- **laravel_core**: Developer experience layer
- **Application**: Business logic and features

## Next Steps

- [ ] Implement Laravel-style facades for all go_core services
- [ ] Add service provider system
- [ ] Create middleware framework
- [ ] Build routing system
- [ ] Add validation framework
- [ ] Implement authentication system
- [ ] Create admin panel scaffolding

## Contributing

When adding new features:
1. Follow Laravel patterns and naming conventions
2. Ensure Go-native performance under the hood
3. Provide comprehensive documentation
4. Include usage examples
5. Add tests for all functionality 