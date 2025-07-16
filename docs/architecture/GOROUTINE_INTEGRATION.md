# Goroutine Integration

## Overview

The framework provides automatic goroutine optimization through canonical constructors and work stealing pools. All core services automatically leverage goroutines for optimal performance without requiring developer intervention.

## Automatic Goroutine Integration

### Canonical Constructors

All core services automatically include goroutine optimization:

```go
import go_core "base_lara_go_project/app/core/go_core"

// Event bus with automatic work stealing pool
eventBus := go_core.NewEventBus()

// Job dispatcher with automatic goroutine pool
dispatcher := go_core.NewJobDispatcher()

// Repository with automatic async operations
repo := go_core.NewRepository[User](db)

// Cache with automatic async operations
cache := go_core.NewCache()
```

### Service Provider Integration

The `AppServiceProvider` automatically sets up all goroutine optimizations:

```go
func (p *AppServiceProvider) Register(container *go_core.Container) error {
    // Register event bus with work stealing pool
    container.Singleton("event_bus", func() (any, error) {
        return go_core.NewEventBus(), nil
    })

    // Register job dispatcher with goroutine pool
    container.Singleton("job_dispatcher", func() (any, error) {
        return go_core.NewJobDispatcher(), nil
    })

    return nil
}
```

### ListenerServiceProvider Integration

Your existing `ListenerServiceProvider` automatically gets goroutine optimization:

```go
func (p *ListenerServiceProvider) Boot(container *go_core.Container) error {
    // Set up automatic goroutine optimization for all listeners
    if err := p.setupGoroutineOptimization(container); err != nil {
        log.Printf("Warning: Failed to setup goroutine optimization: %v", err)
        // Don't fail the boot process if goroutine optimization fails
    }
    return nil
}

func (p *ListenerServiceProvider) setupGoroutineOptimization(container *go_core.Container) error {
    // Get the event bus
    eventBusInstance, err := container.Resolve("event_bus")
    if err != nil {
        return err
    }

    eventBus := eventBusInstance.(go_core.EventBus)

    // Set up automatic goroutine optimization for specific events
    if sendEmailConfirmationInstance, err := container.Resolve("listener.send_email_confirmation"); err == nil {
        sendEmailConfirmation := sendEmailConfirmationInstance.(*listeners.SendEmailConfirmation)
        
        // Register the listener with automatic goroutine optimization
        eventBus.AddListener("user.created", func(event interface{}) error {
            if userData, ok := event.(auth_dto.UserDTO); ok {
                return sendEmailConfirmation.Handle(context.Background(), &go_core.Event[auth_dto.UserDTO]{
                    Data: userData,
                }) // ← Runs in goroutine automatically
            }
            return nil
        })
    }

    return nil
}
```

## Usage Examples

### 1. Event Dispatching with Automatic Goroutines

```go
// Your existing event dispatching code works unchanged
event := &UserCreated{User: user}

// This automatically uses goroutines for all listeners
err := eventBus.Dispatch("user.created", event)

// This definitely uses goroutines
err := eventBus.DispatchAsync("user.created", event)
```

### 2. Job Processing with Automatic Goroutines

```go
// Your existing job dispatching code works unchanged
job := &EmailJob{
    To:      "user@example.com",
    Subject: "Welcome!",
    Body:    "Welcome to our platform!",
}

// This automatically uses goroutines
err := dispatcher.Dispatch(job)

// This definitely uses goroutines
err := dispatcher.DispatchAsync(job)
```

### 3. Repository Operations with Automatic Goroutines

```go
// Create repository with automatic optimization
repo := go_core.NewRepository[User](db)

// Operations automatically use optimal paths
user, err := repo.Find(1)                    // Synchronous, optimized
users, err := repo.Where(conds).Get()        // Prepared statements
err := repo.Complex().BulkCreate(users)      // Advanced operations

// Context-aware operations with timeouts
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
result, err := repo.WithContext(ctx).Find(1)
```

### 4. Manual Goroutine Operations

```go
// Get the goroutine facade
goroutine := facades.Goroutine()

// Async execution
goroutine.Async(func() error {
    return sendEmail()
})

// Parallel execution
errors := goroutine.Parallel(
    func() error { return task1() },
    func() error { return task2() },
    func() error { return task3() },
)

// Retry with backoff
goroutine.Retry(func() error {
    return apiCall()
}, 3, 100*time.Millisecond)

// Batch processing
goroutine.Batch(items, 2, func(batch []interface{}) error {
    return processBatch(batch)
})
```

## Benefits

### 1. Zero Developer Effort
- **No goroutine management**: Developers don't need to think about concurrency
- **Automatic optimization**: Existing code automatically gets goroutine optimization
- **Backward compatible**: All existing code continues to work unchanged

### 2. Seamless SQS Integration
- **Existing SQS system**: Works with your current SQS eventing system
- **Automatic scaling**: Worker pools automatically scale based on load
- **Event processing**: All event listeners automatically run in goroutines

### 3. Performance Improvements
- **Parallel processing**: Multiple listeners run in parallel
- **Async operations**: Database operations can run asynchronously
- **Load balancing**: Automatic worker pool management
- **CPU optimization**: Automatically uses optimal number of workers per CPU

### 4. Developer Experience
- **Laravel-style APIs**: Familiar patterns for Laravel developers
- **Type safety**: Full type safety with generics
- **Error handling**: Comprehensive error handling and retry logic
- **Metrics**: Built-in performance monitoring

## Configuration

### Default Configuration

The system works out of the box with sensible defaults:

```go
// Default worker pool configuration
config := &go_core.GoroutineConfig{
    MinWorkers:    2,                    // Minimum workers
    MaxWorkers:    runtime.NumCPU() * 2, // CPU-aware scaling
    QueueSize:     1000,                 // Job queue size
    IdleTimeout:   30 * time.Second,     // Worker idle timeout
    ShutdownTimeout: 5 * time.Second,    // Graceful shutdown timeout
}
```

### Custom Configuration

You can customize the configuration through environment variables:

```bash
# Goroutine configuration
GOROUTINE_MIN_WORKERS=4
GOROUTINE_MAX_WORKERS=16
GOROUTINE_QUEUE_SIZE=2000
GOROUTINE_IDLE_TIMEOUT=60s
GOROUTINE_SHUTDOWN_TIMEOUT=10s
```

### Work Stealing Pool Configuration

The work stealing pool automatically optimizes for your workload:

```bash
# Work stealing configuration
WORK_STEALING_MIN_WORKERS=2
WORK_STEALING_MAX_WORKERS=8
WORK_STEALING_QUEUE_SIZE=1000
WORK_STEALING_STEAL_INTERVAL=100ms
```

## Performance Monitoring

### Built-in Metrics

The framework provides built-in performance monitoring:

```go
// Get performance stats
stats := eventBus.GetPerformanceStats()
fmt.Printf("Events processed: %d\n", stats["events_processed"])
fmt.Printf("Average processing time: %v\n", stats["avg_processing_time"])

// Get optimization stats
optStats := eventBus.GetOptimizationStats()
fmt.Printf("Work stealing efficiency: %.2f%%\n", optStats["work_stealing_efficiency"])
fmt.Printf("Active workers: %d\n", optStats["active_workers"])
```

### Custom Metrics

You can add custom metrics:

```go
// Track custom metrics
metrics := go_core.NewMetrics()
metrics.Increment("user.registrations")
metrics.Timing("email.send", 150*time.Millisecond)
metrics.Gauge("active.users", 1250)
```

## Best Practices

### 1. Use Canonical Constructors

```go
// ✅ Good: Use canonical constructor
eventBus := go_core.NewEventBus()

// ❌ Avoid: Legacy constructors (removed)
// eventBus := go_core.NewGoroutineAwareEventDispatcher(...)
```

### 2. Leverage Context for Timeouts

```go
// ✅ Good: Context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := eventBus.WithContext(ctx).Dispatch("user.created", user)

// ❌ Avoid: No timeout
// err := eventBus.Dispatch("user.created", user) // Could hang
```

### 3. Handle Errors Gracefully

```go
// ✅ Good: Proper error handling
err := eventBus.Dispatch("user.created", user)
if err != nil {
    log.Printf("Failed to dispatch event: %v", err)
    // Handle error appropriately
}

// ❌ Avoid: Ignoring errors
// eventBus.Dispatch("user.created", user) // Silent failures
```

### 4. Use Appropriate Timeouts

```go
// ✅ Good: Appropriate timeouts for different operations
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Long-running operation
err := dispatcher.WithContext(ctx).Dispatch(&LongRunningJob{})

// Short operation
ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
err := repo.WithContext(ctx).Find(1)
```

## Troubleshooting

### Common Issues

1. **Goroutine Leaks**
   - Ensure all contexts are properly cancelled
   - Use appropriate timeouts for all operations
   - Monitor goroutine count in production

2. **Memory Issues**
   - Use object pools for high-frequency operations
   - Implement proper cleanup in tests
   - Monitor memory usage in production

3. **Performance Issues**
   - Check work stealing pool efficiency
   - Monitor queue sizes and worker utilization
   - Adjust configuration based on workload

### Debug Mode

Enable debug logging to see goroutine usage:

```go
// Debug goroutine operations (temporary, for development)
eventBus := go_core.NewEventBus()
// Debug output will show goroutine usage and performance metrics
```

## Migration Guide

### From Legacy Constructors

If you were using the old wrapper constructors, simply replace them with canonical constructors:

```go
// OLD (removed)
// eventBus := go_core.NewGoroutineAwareEventDispatcher(...)
// dispatcher := go_core.NewGoroutineAwareJobDispatcher(...)
// repo := go_core.NewInfrastructureOptimizedRepository[User](db)

// NEW (canonical)
eventBus := go_core.NewEventBus()
dispatcher := go_core.NewJobDispatcher()
repo := go_core.NewRepository[User](db)
```

The canonical constructors automatically provide all the optimizations that the wrapper constructors provided, with better performance and simpler APIs. 