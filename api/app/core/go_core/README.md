# Go Core Package

A high-performance, type-safe core package for building Laravel-inspired applications in Go with **automatic goroutine optimization**.

## Overview

This package provides optimized, generic-based implementations of core framework components that serve as the foundation for Laravel-style applications. Built with Go's performance characteristics and type safety in mind, with **automatic goroutine optimization** that handles concurrency for you.

## Architecture

```
go_core/
├── cache.go                    # Generic cache interface and implementations
├── container.go                # Service container with dependency injection
├── events.go                   # Event system with type-safe event handling
├── goroutine_core.go           # 🆕 Automatic goroutine optimization system
├── mail.go                     # Mail service with template support
├── performance_features.go     # 🆕 Go-specific performance optimizations
├── queue.go                    # Queue system for background job processing
├── repository.go               # Generic repository pattern with GORM integration
└── README.md                   # This file
```

## Key Features

### 🚀 Performance Optimized
- **Automatic Goroutine Optimization**: Framework automatically uses goroutines for optimal performance
- **Go-Specific Performance Features**: Channel-based pipelines, atomic operations, memory pools
- **Runtime Optimization**: Dynamic optimization based on type analysis
- **Performance Profiling**: Built-in performance monitoring and metrics
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

### 🆕 Automatic Goroutine Optimization
- **Zero Developer Effort**: No need to think about goroutines
- **Worker Pool Management**: Automatic scaling based on load
- **Parallel Processing**: Repository operations, events, and jobs automatically parallelized
- **Metrics & Monitoring**: Built-in performance tracking

### 🆕 Go-Specific Performance Features
- **Channel-Based Pipelines**: Efficient data processing using Go channels
- **Atomic Operations**: Lock-free counters and operations for maximum performance
- **Memory Pools**: Object reuse to reduce garbage collection pressure
- **Dynamic Optimization**: Runtime type analysis for automatic optimization
- **Performance Profiling**: Built-in metrics and monitoring
- **Interface-Based Polymorphism**: Runtime optimization strategies

## Usage Examples

### Repository Pattern with Automatic Goroutine Optimization
```go
import "your-project/api/app/core/go_core"

// Define your model
type User struct {
    ID   uint   `gorm:"primaryKey"`
    Name string `gorm:"not null"`
}

// Create repository with automatic goroutine optimization
userRepo := go_core.NewGoroutineAwareRepository[User](go_core.NewRepository[User](db))

// Type-safe operations - now automatically optimized with goroutines
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

### Event System with Automatic Goroutine Optimization
```go
dispatcher := go_core.NewGoroutineAwareEventDispatcher[User](go_core.NewEventBus[User]())

// Register listeners
dispatcher.Listen("user.created", func(event go_core.Event[User]) {
    // Handle user created event - automatically runs in goroutine
})

// Dispatch events - automatically optimized
dispatcher.Dispatch(&go_core.Event[User]{Name: "user.created", Data: user})
dispatcher.DispatchAsync(&go_core.Event[User]{Name: "user.created", Data: user}) // Parallel
```

### Queue System with Automatic Goroutine Optimization
```go
queue := go_core.NewQueue[Job]()
dispatcher := go_core.NewGoroutineAwareJobDispatcher[Job](go_core.NewJobDispatcher(queue))

// Push jobs - automatically optimized with goroutines
dispatcher.Dispatch(&EmailJob{To: "user@example.com"})
dispatcher.DispatchAsync(&EmailJob{To: "user@example.com"}) // Parallel processing

// Process jobs
queue.Process(func(job Job) error {
    return job.Execute()
})
```

### Laravel-Style Facade with Automatic Goroutine Optimization
```go
import "your-project/api/app/core/laravel_core/facades"

// Automatic goroutine optimization through facades
goroutine := facades.Goroutine()

// Async operations
goroutine.Async(func() error {
    // This runs in a goroutine automatically
    return sendEmail()
})

// Parallel processing
errors := goroutine.Parallel(
    func() error { return task1() },
    func() error { return task2() },
    func() error { return task3() },
)

// Batch processing with goroutines
goroutine.Batch(items, 100, func(batch []interface{}) error {
    // Process batch in parallel
    return processBatch(batch)
})

// Map operations with goroutines
results, err := goroutine.Map(items, func(item interface{}) (interface{}, error) {
    // Transform item - runs in parallel
    return transform(item)
})

// Repository with automatic goroutine optimization
userRepo := facades.NewGoroutineAwareRepository[User](db)
userChan := userRepo.FindAsync(1) // Automatic goroutine optimization
```

### Go-Specific Performance Features
```go
import "your-project/api/app/core/go_core"

// Performance facade for easy access to all optimizations
perf := go_core.NewPerformanceFacade()

// Track performance of operations
perf.Track("user_processing", func() error {
    return processUsers(users)
})

// Channel-based pipeline processing
pipeline := go_core.NewPipeline[User]().
    AddStage(go_core.FilterStage[User](func(user User) bool {
        return user.Active
    })).
    AddStage(go_core.TransformStage[User](func(user User) User {
        user.Name = strings.ToUpper(user.Name)
        return user
    }))

results := pipeline.Execute(users)
for result := range results {
    // Process each result
}

// Atomic operations for lock-free performance
counter := go_core.NewAtomicCounter()
counter.Increment() // Thread-safe

// Memory pool for object reuse
pool := go_core.NewObjectPool[User](100, 
    func() User { return User{} },
    func(user User) User { return User{} },
)

user := pool.Get()
defer pool.Put(user)

// Get performance statistics
stats := perf.GetStats()
fmt.Printf("Goroutines: %v\n", stats["num_goroutines"])
```

## Integration with Laravel Core

This package works seamlessly with the `laravel_core` package, which provides Laravel-style utilities and patterns:

- **go_core**: High-performance, type-safe foundation with automatic goroutine optimization
- **laravel_core**: Laravel-inspired application layer with developer-friendly facades

## Benefits

1. **Performance**: Go-native implementations with automatic goroutine optimization
2. **Type Safety**: Compile-time guarantees prevent runtime errors
3. **Developer Experience**: No need to think about goroutines - it's automatic
4. **Maintainability**: Clean, generic-based code that's easy to understand
5. **Extensibility**: Interface-based design allows easy customization
6. **Testing**: Mockable interfaces simplify unit testing
7. **Automatic Scaling**: Worker pools scale based on load automatically

## Goroutine Optimization Features

### Automatic Worker Pool Management
- **Auto-scaling**: Worker pools scale up/down based on load
- **CPU-aware**: Automatically uses optimal number of workers per CPU
- **Queue Management**: Intelligent job queuing and processing
- **Metrics**: Built-in performance monitoring

### Parallel Processing
- **Repository Operations**: Find, Create, Update, Delete operations automatically parallelized
- **Event Handling**: Event listeners run in parallel
- **Job Processing**: Background jobs automatically use goroutines
- **Batch Operations**: Large datasets processed in parallel

### Developer Experience
- **Zero Configuration**: Works out of the box with sensible defaults
- **Laravel-Style APIs**: Familiar patterns for Laravel developers
- **Type Safety**: Full type safety with generics
- **Error Handling**: Comprehensive error handling and recovery

## Next Steps

- [x] Add automatic goroutine optimization
- [x] Implement worker pool management
- [x] Add parallel processing for repositories
- [x] Create Laravel-style facades
- [x] Add Go-specific performance features
- [x] Implement channel-based pipelines
- [x] Add atomic operations and memory pools
- [x] Create performance profiling system
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
6. Ensure automatic goroutine optimization where appropriate 