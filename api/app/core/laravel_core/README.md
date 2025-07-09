# Laravel Core Package

Laravel-inspired utilities and patterns for building applications with familiar Laravel developer experience and **automatic goroutine optimization**.

## Overview

This package provides Laravel-style patterns and utilities that work on top of the high-performance `go_core` package. It offers familiar Laravel developer experience while leveraging Go's performance characteristics and **automatic goroutine optimization**.

## Architecture

```
laravel_core/
├── config/           # Configuration management
├── facades/          # Laravel-style facades with goroutine optimization
├── logging/          # Comprehensive logging system
├── exceptions/       # Exception handling
├── env/             # Environment variable management
├── models/          # Base model system
├── dtos/            # Data transfer objects
├── clients/         # Client interfaces
├── observers/       # Model observers
├── providers/       # Service providers including GoroutineServiceProvider
└── README.md        # This file
```

## Key Features

### 🏗️ Laravel-Style Patterns
- **Facades**: Static-like access to services with automatic goroutine optimization
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
- **Automatic Goroutine Optimization**: No need to think about concurrency

### 🆕 Automatic Goroutine Optimization
- **Zero Developer Effort**: No need to think about goroutines
- **Event/Listener Integration**: Works seamlessly with your existing SQS eventing system
- **Service Provider Integration**: Automatically optimizes listeners and jobs
- **Worker Pool Management**: Automatic scaling based on load

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

### Facades with Automatic Goroutine Optimization
```go
import "your-project/api/app/core/laravel_core/facades"

// Use facades for easy access with automatic goroutine optimization
user := facades.DB().Table("users").Where("id", 1).First()
facades.Cache().Set("user:1", user, 30*time.Minute)
facades.Log().Info("User retrieved", map[string]interface{}{"user_id": 1})

// Automatic goroutine optimization
goroutine := facades.Goroutine()
goroutine.Async(func() error {
    // This runs in a goroutine automatically
    return sendEmail()
})
```

### Event System with Automatic Goroutine Optimization
```go
import "your-project/api/app/core/laravel_core/facades"

// Events automatically use goroutines for optimal performance
facades.Event(&UserCreatedEvent{User: user}) // Synchronous
facades.EventAsync(&UserCreatedEvent{User: user}) // Asynchronous with goroutines

// Listeners automatically run in goroutines
// No additional code needed - it's handled by the GoroutineServiceProvider
```

### Service Providers with Goroutine Optimization
```go
// Your existing ListenerServiceProvider automatically gets goroutine optimization
type ListenerServiceProvider struct {
    laravel_providers.BaseServiceProvider
}

func (p *ListenerServiceProvider) Register(container *app_core.Container) error {
    // Register listeners as usual
    container.Singleton("listener.send_email_confirmation", func() (any, error) {
        return &listeners.SendEmailConfirmation{}, nil
    })
    return nil
}

func (p *ListenerServiceProvider) Boot(container *app_core.Container) error {
    // Goroutine optimization is automatically set up
    // Your listeners will run in goroutines without any additional code
    return nil
}
```

### Repository Operations with Automatic Goroutine Optimization
```go
import "your-project/api/app/core/laravel_core/facades"

// Create goroutine-aware repository
userRepo := facades.NewGoroutineAwareRepository[User](db)

// Operations automatically use goroutines for optimal performance
user, err := userRepo.Find(1)                    // Synchronous
userChan := userRepo.FindAsync(1)                // Asynchronous with goroutines
usersChan := userRepo.FindManyAsync([]uint{1,2,3}) // Parallel processing

// Wait for async results
select {
case result := <-userChan:
    if result.Error != nil {
        // Handle error
    }
    user = result.Data
}
```

## Integration with Go Core

This package is designed to work seamlessly with the `go_core` package:

```go
// Go Core provides the foundation with automatic goroutine optimization
userRepo := go_core.NewGoroutineAwareRepository[User](go_core.NewRepository[User](db))
cache := go_core.NewMemoryCache[User]()

// Laravel Core provides the developer experience
facades.DB().Table("users").Where("active", true).Get()
facades.Cache().Remember("users:active", 30*time.Minute, func() interface{} {
    return userRepo.FindAll()
})

// Automatic goroutine optimization throughout
goroutine := facades.Goroutine()
goroutine.Parallel(
    func() error { return task1() },
    func() error { return task2() },
    func() error { return task3() },
)
```

## Goroutine Optimization Features

### Automatic Integration with Existing Systems
- **SQS Eventing**: Works seamlessly with your existing SQS event/listener system
- **Service Providers**: Automatically optimizes listeners registered in service providers
- **Event Dispatching**: Events automatically use goroutines for optimal performance
- **Job Processing**: Background jobs automatically use goroutines

### Worker Pool Management
- **Auto-scaling**: Worker pools scale up/down based on load
- **CPU-aware**: Automatically uses optimal number of workers per CPU
- **Queue Management**: Intelligent job queuing and processing
- **Metrics**: Built-in performance monitoring

### Developer Experience
- **Zero Configuration**: Works out of the box with sensible defaults
- **Backward Compatible**: Existing code continues to work without changes
- **Laravel-Style APIs**: Familiar patterns for Laravel developers
- **Type Safety**: Full type safety with generics

## Benefits

1. **Familiarity**: Laravel developers feel at home
2. **Productivity**: Rapid development with familiar patterns
3. **Performance**: Go-native performance with automatic goroutine optimization
4. **Maintainability**: Clean, well-structured code
5. **Extensibility**: Easy to customize and extend
6. **Zero Effort Optimization**: No need to think about goroutines - it's automatic
7. **Seamless Integration**: Works with your existing SQS eventing system

## Design Philosophy

### Laravel Patterns, Go Performance, Automatic Optimization
- Use Laravel-style APIs for developer experience
- Implement with Go-optimized code for performance
- Provide automatic goroutine optimization without developer effort
- Maintain Go idioms under the hood

### Separation of Concerns
- **go_core**: High-performance foundation with automatic goroutine optimization
- **laravel_core**: Developer experience layer with seamless integration
- **Application**: Business logic and features

## Next Steps

- [x] Implement automatic goroutine optimization
- [x] Integrate with existing event/listener system
- [x] Create GoroutineServiceProvider
- [x] Add Laravel-style facades with goroutine optimization
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
3. Provide automatic goroutine optimization where appropriate
4. Provide comprehensive documentation
5. Include usage examples
6. Add tests for all functionality 