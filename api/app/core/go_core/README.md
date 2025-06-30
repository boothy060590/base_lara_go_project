# Go Core Package

A high-performance, type-safe core package for building Laravel-inspired applications in Go.

## Overview

This package provides optimized, generic-based implementations of core framework components that serve as the foundation for Laravel-style applications. Built with Go's performance characteristics and type safety in mind.

## Architecture

```
go_core/
├── cache.go          # Generic cache interface and implementations
├── container.go      # Service container with dependency injection
├── events.go         # Event system with type-safe event handling
├── mail.go          # Mail service with template support
├── queue.go         # Queue system for background job processing
├── repository.go    # Generic repository pattern with GORM integration
└── README.md        # This file
```

## Key Features

### 🚀 Performance Optimized
- Generic-based implementations for zero-cost abstractions
- Type-safe interfaces prevent runtime errors
- Efficient memory usage with proper resource management

### 🔧 Type Safety
- Compile-time type checking for all operations
- Generic constraints ensure correct data types
- Interface-based design for easy testing and mocking

### 🏗️ Laravel-Inspired Patterns
- Repository pattern for data access
- Service container for dependency injection
- Event-driven architecture
- Queue system for background processing

## Usage Examples

### Repository Pattern
```go
import "your-project/api/app/core/go_core"

// Define your model
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"not null"`
}

// Create repository
userRepo := go_core.NewRepository[User](db)

// Type-safe operations
user, err := userRepo.FindByID(1)
users, err := userRepo.FindAll()
err = userRepo.Create(&User{Name: "John"})
```

### Service Container
```go
container := go_core.NewContainer()

// Register services
container.Singleton("cache", func() go_core.Cache[any] {
    return go_core.NewMemoryCache[any]()
})

// Resolve services
cache := container.Make("cache").(go_core.Cache[any])
```

### Event System
```go
dispatcher := go_core.NewEventDispatcher()

// Register listeners
dispatcher.Listen("user.created", func(event go_core.Event) {
    // Handle user created event
})

// Dispatch events
dispatcher.Dispatch("user.created", go_core.NewEvent("user.created", userData))
```

### Queue System
```go
queue := go_core.NewQueue[Job]()

// Push jobs
queue.Push(&EmailJob{To: "user@example.com"})

// Process jobs
queue.Process(func(job Job) error {
    return job.Execute()
})
```

## Integration with Laravel Core

This package works seamlessly with the `laravel_core` package, which provides Laravel-style utilities and patterns:

- **go_core**: High-performance, type-safe foundation
- **laravel_core**: Laravel-inspired application layer

## Benefits

1. **Performance**: Go-native implementations with minimal overhead
2. **Type Safety**: Compile-time guarantees prevent runtime errors
3. **Maintainability**: Clean, generic-based code that's easy to understand
4. **Extensibility**: Interface-based design allows easy customization
5. **Testing**: Mockable interfaces simplify unit testing

## Next Steps

- [ ] Add more specialized repository methods
- [ ] Implement additional cache backends (Redis, etc.)
- [ ] Add queue backends (Redis, RabbitMQ, etc.)
- [ ] Create middleware system
- [ ] Add validation framework
- [ ] Implement authentication system

## Contributing

When adding new features:
1. Use generics for type safety
2. Implement interfaces for testability
3. Follow Go idioms and best practices
4. Add comprehensive tests
5. Update documentation 