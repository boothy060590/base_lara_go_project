# Goroutine Integration with SQS Eventing System

## Overview

This document explains how the automatic goroutine optimization system integrates seamlessly with your existing SQS eventing system, providing **zero-effort concurrency** without requiring developers to think about goroutines.

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    SQS Eventing System                          │
├─────────────────────────────────────────────────────────────────┤
│  SQS Queue  │  SQS Worker  │  Event Processor  │  Event Bus    │
│             │              │                   │               │
│ ┌─────────┐ │ ┌──────────┐ │ ┌──────────────┐ │ ┌──────────┐  │
│ │Message 1│ │ │Worker 1  │ │ │Event Creator│ │ │Event Bus │  │
│ │Message 2│ │ │Worker 2  │ │ │Event Router │ │ │Listener 1│  │
│ │Message 3│ │ │Worker N  │ │ │Event Store  │ │ │Listener 2│  │
│ └─────────┘ │ └──────────┘ │ └──────────────┘ │ └──────────┘  │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│              Goroutine Optimization Layer                       │
├─────────────────────────────────────────────────────────────────┤
│  GoroutineServiceProvider  │  GoroutineManager  │  Worker Pool │
│                            │                    │              │
│ ┌────────────────────────┐ │ ┌────────────────┐ │ ┌──────────┐ │
│ │• Auto-registers        │ │ │• Manages       │ │ │• Auto-   │ │
│ │  goroutine services    │ │ │  worker pools  │ │ │  scaling │ │
│ │• Integrates with       │ │ │• CPU-aware     │ │ │• Load    │ │
│ │  existing listeners    │ │ │  scaling       │ │ │  balancing│ │
│ │• Zero-config setup     │ │ │• Metrics       │ │ │• Queue   │ │
│ └────────────────────────┘ │ └────────────────┘ │ └──────────┘ │
└─────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Application Layer                            │
├─────────────────────────────────────────────────────────────────┤
│  Listeners  │  Jobs  │  Repositories  │  Services  │  Controllers│
│             │        │                │            │             │
│ ┌─────────┐ │ ┌────┐ │ ┌────────────┐ │ ┌────────┐ │ ┌─────────┐ │
│ │Email    │ │ │Job1│ │ │UserRepo    │ │ │UserSvc │ │ │AuthCtrl │ │
│ │Listener │ │ │Job2│ │ │OrderRepo   │ │ │OrderSvc│ │ │UserCtrl │ │
│ │SMS      │ │ │Job3│ │ │ProductRepo │ │ │MailSvc │ │ │AdminCtrl│ │
│ │Listener │ │ └────┘ │ └────────────┘ │ └────────┘ │ └─────────┘ │
│ └─────────┘ │        │                │            │             │
└─────────────────────────────────────────────────────────────────┘
```

## How It Works

### 1. Automatic Integration

The `GoroutineServiceProvider` automatically integrates with your existing event system:

```go
// In your AppServiceProvider
appProviders := []laravel_providers.ServiceProvider{
    &laravel_providers.GoroutineServiceProvider{}, // ← This enables everything
    &ListenerServiceProvider{},
    &RepositoryServiceProvider{},
    &RouterServiceProvider{},
}
```

### 2. Zero-Effort Listener Optimization

Your existing listeners automatically run in goroutines:

```go
// Your existing listener (no changes needed)
type SendEmailConfirmation struct {
    laravel_listeners.BaseListener[auth_dto.UserDTO]
}

func (l *SendEmailConfirmation) Handle(ctx context.Context, event *app_core.Event[auth_dto.UserDTO]) error {
    // This automatically runs in a goroutine!
    return sendEmail(event.Data.Email)
}

// In your ListenerServiceProvider (automatic setup)
func (p *ListenerServiceProvider) setupGoroutineOptimization(container *app_core.Container) error {
    // The GoroutineServiceProvider automatically optimizes this
    eventManager.Listen("user.created", func(ctx context.Context, event *app_core.Event[any]) error {
        return sendEmailConfirmation.Handle(ctx, event) // ← Runs in goroutine
    })
    return nil
}
```

### 3. SQS Integration

Your SQS workers automatically benefit from goroutine optimization:

```go
// In your SQS worker (existing code)
func processSQSEvent(sqsMessage map[string]interface{}) error {
    // Create event from SQS message
    user := auth_dto.UserDTO{
        ID:        uint(sqsMessage["user_id"].(int)),
        Email:     sqsMessage["email"].(string),
        FirstName: sqsMessage["first_name"].(string),
        LastName:  sqsMessage["last_name"].(string),
    }

    event := &app_core.Event[any]{
        ID:        "sqs_event_123",
        Name:      sqsMessage["event_type"].(string),
        Data:      user,
        Timestamp: time.Now(),
        Source:    "sqs",
    }

    // Dispatch event - automatically uses goroutines for all listeners
    return eventManager.Dispatch(event) // ← All listeners run in goroutines
}
```

## Service Provider Integration

### GoroutineServiceProvider

The `GoroutineServiceProvider` registers all goroutine-optimized services:

```go
type GoroutineServiceProvider struct {
    BaseServiceProvider
}

func (p *GoroutineServiceProvider) Register(container *app_core.Container) error {
    // Register goroutine manager
    container.Singleton("goroutine.manager", func() (any, error) {
        return app_core.NewGoroutineManager[any](nil), nil
    })

    // Register goroutine-aware event dispatcher
    container.Singleton("goroutine.event_dispatcher", func() (any, error) {
        eventBus := app_core.NewEventBus[any]()
        goroutineManager := app_core.NewGoroutineManager[any](nil)
        return app_core.NewGoroutineAwareEventDispatcher[any](eventBus, goroutineManager), nil
    })

    // Register goroutine-aware job dispatcher
    container.Singleton("goroutine.job_dispatcher", func() (any, error) {
        queue := app_core.NewSyncQueue[any]()
        jobDispatcher := app_core.NewJobDispatcher[any](queue)
        goroutineManager := app_core.NewGoroutineManager[any](nil)
        return app_core.NewGoroutineAwareJobDispatcher[any](jobDispatcher, goroutineManager), nil
    })

    return nil
}
```

### ListenerServiceProvider Integration

Your existing `ListenerServiceProvider` automatically gets goroutine optimization:

```go
func (p *ListenerServiceProvider) Boot(container *app_core.Container) error {
    // Set up automatic goroutine optimization for all listeners
    if err := p.setupGoroutineOptimization(container); err != nil {
        log.Printf("Warning: Failed to setup goroutine optimization: %v", err)
        // Don't fail the boot process if goroutine optimization fails
    }
    return nil
}

func (p *ListenerServiceProvider) setupGoroutineOptimization(container *app_core.Container) error {
    // Get the event manager
    eventManagerInstance, err := container.Resolve("event_manager")
    if err != nil {
        return err
    }

    eventManager := eventManagerInstance.(app_core.EventManagerInterface[any])

    // Set up automatic goroutine optimization for specific events
    if sendEmailConfirmationInstance, err := container.Resolve("listener.send_email_confirmation"); err == nil {
        sendEmailConfirmation := sendEmailConfirmationInstance.(*listeners.SendEmailConfirmation)
        
        // Register the listener with goroutine optimization
        eventManager.Listen("user.created", func(ctx context.Context, event *app_core.Event[any]) error {
            // Convert event data to correct type
            if userData, ok := event.Data.(auth_dto.UserDTO); ok {
                typedEvent := &app_core.Event[auth_dto.UserDTO]{
                    ID:        event.ID,
                    Name:      event.Name,
                    Data:      userData,
                    Timestamp: event.Timestamp,
                    Source:    event.Source,
                }
                return sendEmailConfirmation.Handle(ctx, typedEvent) // ← Runs in goroutine
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
event := &app_core.Event[any]{
    ID:        "event_123",
    Name:      "user.created",
    Data:      user,
    Timestamp: time.Now(),
    Source:    "api",
}

// This automatically uses goroutines for all listeners
eventManager.Dispatch(event)

// This definitely uses goroutines
eventManager.DispatchAsync(event)
```

### 2. Job Processing with Automatic Goroutines

```go
// Your existing job dispatching code works unchanged
job := app_core.Job[any]{
    ID: "job_123",
    Data: map[string]interface{}{
        "action": "send_email",
        "to":     "user@example.com",
        "subject": "Welcome!",
    },
}

// This automatically uses goroutines
jobDispatcher.DispatchAsync(job)
```

### 3. Repository Operations with Automatic Goroutines

```go
// Create goroutine-aware repository
userRepo := app_core.NewGoroutineAwareRepository[User](db)

// Operations automatically use goroutines
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
config := &app_core.GoroutineConfig{
    MinWorkers:    2,                    // Minimum workers
    MaxWorkers:    runtime.NumCPU() * 2, // CPU-aware scaling
    QueueSize:     1000,                 // Job queue size
    IdleTimeout:   30 * time.Second,     // Worker idle timeout
    ShutdownTimeout: 5 * time.Second,    // Graceful shutdown timeout
}
```

### Custom Configuration

You can customize the configuration:

```go
// Custom goroutine manager
config := &app_core.GoroutineConfig{
    MinWorkers:    5,
    MaxWorkers:    20,
    QueueSize:     5000,
    IdleTimeout:   60 * time.Second,
    ShutdownTimeout: 10 * time.Second,
}

manager := app_core.NewGoroutineManager[any](config)
```

## Monitoring and Metrics

### Built-in Metrics

The system provides comprehensive metrics:

```go
// Get metrics
metrics := goroutine.GetMetrics()
fmt.Printf("Active workers: %d\n", goroutine.GetActiveWorkerCount())
fmt.Printf("Queue length: %d\n", goroutine.GetQueueLength())
fmt.Printf("Total jobs processed: %d\n", metrics.TotalJobsProcessed)
fmt.Printf("Average processing time: %v\n", metrics.AverageProcessingTime)
```

### Performance Monitoring

```go
// Monitor worker pool performance
for {
    activeWorkers := goroutine.GetActiveWorkerCount()
    queueLength := goroutine.GetQueueLength()
    
    if queueLength > 100 {
        log.Printf("High queue length: %d, active workers: %d", queueLength, activeWorkers)
    }
    
    time.Sleep(10 * time.Second)
}
```

## Migration Guide

### From Existing SQS System

1. **Add GoroutineServiceProvider** to your service providers
2. **No code changes needed** - existing listeners automatically get goroutine optimization
3. **Optional**: Use goroutine-aware repositories for async database operations
4. **Optional**: Use manual goroutine operations for custom concurrency needs

### Example Migration

```go
// Before (existing code)
type SendEmailConfirmation struct {
    laravel_listeners.BaseListener[auth_dto.UserDTO]
}

func (l *SendEmailConfirmation) Handle(ctx context.Context, event *app_core.Event[auth_dto.UserDTO]) error {
    return sendEmail(event.Data.Email) // Runs synchronously
}

// After (same code, but now runs in goroutines automatically)
type SendEmailConfirmation struct {
    laravel_listeners.BaseListener[auth_dto.UserDTO]
}

func (l *SendEmailConfirmation) Handle(ctx context.Context, event *app_core.Event[auth_dto.UserDTO]) error {
    return sendEmail(event.Data.Email) // Now runs in goroutines automatically!
}
```

## Best Practices

### 1. Event Design
- **Keep events small**: Events should contain minimal data
- **Use DTOs**: Use data transfer objects for event data
- **Handle errors**: Always handle errors in event listeners

### 2. Listener Design
- **Stateless listeners**: Listeners should be stateless
- **Idempotent operations**: Make listeners idempotent for retry safety
- **Error handling**: Proper error handling and logging

### 3. Performance Optimization
- **Monitor metrics**: Use built-in metrics to monitor performance
- **Tune worker pools**: Adjust worker pool size based on load
- **Use async operations**: Use async repository operations for database calls

### 4. Error Handling
- **Retry logic**: Use built-in retry mechanisms
- **Circuit breakers**: Implement circuit breakers for external services
- **Dead letter queues**: Handle failed events appropriately

## Conclusion

The goroutine optimization system provides **automatic concurrency** for your existing SQS eventing system with **zero developer effort**. Your existing listeners, jobs, and repository operations automatically run in goroutines, providing significant performance improvements without requiring any code changes.

The system is designed to be:
- **Backward compatible**: Existing code works unchanged
- **Zero configuration**: Works out of the box with sensible defaults
- **Performance focused**: Automatic scaling and optimization
- **Developer friendly**: Laravel-style APIs with full type safety

This integration ensures that your SQS eventing system can handle high loads efficiently while maintaining the familiar Laravel developer experience. 