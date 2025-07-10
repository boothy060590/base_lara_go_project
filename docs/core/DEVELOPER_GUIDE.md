# Developer Guide

## Introduction

This guide shows you how to use the Laravel-Inspired Go Framework to build high-performance applications with familiar Laravel-style APIs. The framework automatically optimizes your code without requiring any additional configuration.

## Quick Start

### **1. Basic Setup**

```go
// main.go
package main

import (
    "your-project/api/app/core/laravel_core/facades"
)

func main() {
    // Framework automatically initializes all optimizations
    app := facades.App()
    app.Run()
}
```

### **2. Create Your First Controller**

```go
// app/http/controllers/user_controller.go
package controllers

import (
    "your-project/api/app/core/laravel_core/facades"
)

type UserController struct {
    userRepo *facades.Repository
    eventDispatcher *facades.Event
}

func NewUserController() *UserController {
    return &UserController{
        userRepo: facades.Repository(),
        eventDispatcher: facades.Event(),
    }
}

func (c *UserController) Show(id uint) (*User, error) {
    // Framework automatically optimizes this query
    user, err := c.userRepo.Find(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}

func (c *UserController) Store(request *CreateUserRequest) (*User, error) {
    user := &User{
        Name:  request.Name,
        Email: request.Email,
    }
    
    // Framework automatically optimizes this operation
    if err := c.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // Framework automatically optimizes this event dispatch
    c.eventDispatcher.Dispatch(&UserCreated{User: user})
    
    return user, nil
}
```

### **3. Define Your Models**

```go
// app/models/user.go
package models

import (
    "gorm.io/gorm"
)

type User struct {
    ID        uint   `gorm:"primaryKey"`
    Name      string `gorm:"not null"`
    Email     string `gorm:"unique;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Framework automatically provides optimized repository methods
func (u *User) TableName() string {
    return "users"
}
```

### **4. Create Events and Listeners**

```go
// app/events/user_created.go
package events

type UserCreated struct {
    User *User
}

// app/listeners/send_welcome_email.go
package listeners

import (
    "your-project/api/app/core/laravel_core/facades"
)

type SendWelcomeEmail struct {
    mailer *facades.Mail
}

func NewSendWelcomeEmail() *SendWelcomeEmail {
    return &SendWelcomeEmail{
        mailer: facades.Mail(),
    }
}

func (l *SendWelcomeEmail) Handle(event *UserCreated) error {
    // Framework automatically optimizes this email sending
    return l.mailer.Send("welcome", event.User.Email, map[string]interface{}{
        "user": event.User,
    })
}
```

## Core Concepts

### **1. Repository Pattern**

The framework provides an optimized repository pattern that automatically uses goroutines and caching.

```go
// Get repository instance
userRepo := facades.Repository()

// Basic operations (automatically optimized)
user, err := userRepo.Find(1)
users, err := userRepo.FindMany([]uint{1, 2, 3})
err := userRepo.Create(&User{Name: "John", Email: "john@example.com"})
err := userRepo.Update(user)
err := userRepo.Delete(1)

// Async operations (automatic goroutine optimization)
userChan := userRepo.FindAsync(1)
usersChan := userRepo.FindManyAsync([]uint{1, 2, 3})
errChan := userRepo.CreateAsync(&User{Name: "John", Email: "john@example.com"})

// Wait for async results
select {
case result := <-userChan:
    if result.Error != nil {
        // Handle error
    }
    user = result.Data
case <-time.After(5 * time.Second):
    // Handle timeout
}
```

### **2. Event System**

High-performance event dispatching with automatic optimization.

```go
// Get event dispatcher
eventDispatcher := facades.Event()

// Dispatch events (automatically optimized)
err := eventDispatcher.Dispatch(&UserCreated{User: user})
err := eventDispatcher.DispatchAsync(&UserCreated{User: user})

// Listen for events
eventDispatcher.Listen("user.created", func(event *UserCreated) error {
    // Handle event
    return nil
})

// Multiple listeners (processed in parallel)
eventDispatcher.Listen("user.created", &SendWelcomeEmail{})
eventDispatcher.Listen("user.created", &LogUserCreation{})
eventDispatcher.Listen("user.created", &UpdateUserCount{})
```

### **3. Cache System**

Multi-level caching with automatic optimization.

```go
// Get cache instance
cache := facades.Cache()

// Basic operations (automatically optimized)
err := cache.Set("user:1", user, 3600)
user, err := cache.Get("user:1")
err := cache.Delete("user:1")

// Multi-level caching (memory -> Redis -> Database)
user, err := cache.Remember("user:1", 3600, func() (*User, error) {
    return userRepo.Find(1)
})

// Cache tags for easy invalidation
err := cache.Tags("users").Set("user:1", user)
cache.Tags("users").Flush() // Invalidate all user cache
```

### **4. Queue System**

Background job processing with automatic scaling.

```go
// Get queue instance
queue := facades.Queue()

// Push jobs (automatically optimized)
err := queue.Push(&SendEmailJob{Email: "user@example.com"})
err := queue.PushAsync(&ProcessOrderJob{OrderID: 123})

// Process jobs
queue.Process(&SendEmailJob{}, func(job *SendEmailJob) error {
    // Process job
    return nil
})

// Job with retry logic
type SendEmailJob struct {
    Email string
    Retries int
}

func (j *SendEmailJob) Handle() error {
    // Send email logic
    return nil
}

func (j *SendEmailJob) Failed(err error) {
    // Handle failed job
    if j.Retries < 3 {
        j.Retries++
        facades.Queue().Push(j)
    }
}
```

### **5. Mail System**

Asynchronous email processing with template support.

```go
// Get mailer instance
mailer := facades.Mail()

// Send emails (automatically optimized)
err := mailer.Send("welcome", "user@example.com", map[string]interface{}{
    "user": user,
})

// Send with attachments
err := mailer.SendWithAttachments("invoice", "user@example.com", map[string]interface{}{
    "invoice": invoice,
}, []string{"invoice.pdf"})

// Queue emails for background processing
err := mailer.Queue("welcome", "user@example.com", map[string]interface{}{
    "user": user,
})
```

## Advanced Usage

### **1. Custom Repositories**

```go
// app/repositories/user_repository.go
package repositories

import (
    "your-project/api/app/core/go_core"
)

type UserRepository struct {
    *go_core.Repository[User]
}

func NewUserRepository() *UserRepository {
    return &UserRepository{
        Repository: go_core.NewRepository[User](db),
    }
}

// Custom methods with automatic optimization
func (r *UserRepository) FindByEmail(email string) (*User, error) {
    var user User
    err := r.db.Where("email = ?", email).First(&user).Error
    return &user, err
}

func (r *UserRepository) FindActiveUsers() ([]*User, error) {
    var users []*User
    err := r.db.Where("active = ?", true).Find(&users).Error
    return users, err
}
```

### **2. Custom Events**

```go
// app/events/order_placed.go
package events

type OrderPlaced struct {
    Order *Order
    User  *User
}

// app/listeners/process_order.go
package listeners

type ProcessOrder struct {
    orderProcessor *OrderProcessor
}

func (l *ProcessOrder) Handle(event *OrderPlaced) error {
    // Process order with automatic optimization
    return l.orderProcessor.Process(event.Order)
}
```

### **3. Custom Jobs**

```go
// app/jobs/send_email_job.go
package jobs

import (
    "your-project/api/app/core/go_core"
)

type SendEmailJob struct {
    go_core.Job
    Email   string
    Subject string
    Body    string
}

func (j *SendEmailJob) Handle() error {
    // Send email logic
    return nil
}

func (j *SendEmailJob) Failed(err error) {
    // Handle failure
    log.Printf("Failed to send email: %v", err)
}
```

### **4. Custom Cache Drivers**

```go
// app/cache/redis_cache.go
package cache

import (
    "your-project/api/app/core/go_core"
)

type RedisCache struct {
    *go_core.Cache[any]
    client *redis.Client
}

func NewRedisCache() *RedisCache {
    return &RedisCache{
        Cache:  go_core.NewCache[any](),
        client: redis.NewClient(&redis.Options{}),
    }
}

func (c *RedisCache) Get(key string) (any, error) {
    // Custom Redis implementation with automatic optimization
    return c.client.Get(key).Result()
}
```

## Configuration

### **1. Environment Configuration**

```bash
# .env
APP_NAME="My Application"
APP_ENV=production
APP_DEBUG=false

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
```

### **2. Service Providers**

```go
// app/providers/app_service_provider.go
package providers

import (
    "your-project/api/app/core/laravel_core/providers"
)

type AppServiceProvider struct {
    providers.BaseServiceProvider
}

func (p *AppServiceProvider) Register(container *go_core.Container) error {
    // Register your services
    container.Singleton("user.repository", func() (any, error) {
        return repositories.NewUserRepository(), nil
    })
    
    return nil
}

func (p *AppServiceProvider) Boot(container *go_core.Container) error {
    // Boot your services
    return nil
}
```

### **3. Middleware**

```go
// app/http/middleware/auth_middleware.go
package middleware

import (
    "your-project/api/app/core/laravel_core/facades"
)

type AuthMiddleware struct{}

func (m *AuthMiddleware) Handle(request *http.Request, next func() *http.Response) *http.Response {
    // Authentication logic
    token := request.Header.Get("Authorization")
    if token == "" {
        return &http.Response{
            StatusCode: 401,
            Body:       strings.NewReader("Unauthorized"),
        }
    }
    
    // Continue to next middleware/controller
    return next()
}
```

## Performance Monitoring

### **1. Built-in Metrics**

```go
// Get performance metrics
metrics := facades.Performance()

// View current metrics
stats := metrics.GetStats()
fmt.Printf("Operations per second: %d\n", stats.OperationsPerSecond)
fmt.Printf("Average response time: %v\n", stats.AverageResponseTime)
fmt.Printf("Memory usage: %d MB\n", stats.MemoryUsage)
fmt.Printf("Goroutine count: %d\n", stats.GoroutineCount)
```

### **2. Custom Metrics**

```go
// Track custom metrics
metrics := facades.Performance()

// Increment counter
metrics.Increment("user.registrations")

// Record timing
metrics.Timing("database.query", 150*time.Millisecond)

// Set gauge
metrics.Gauge("active.users", 1250)
```

### **3. Health Checks**

```go
// Health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := facades.App().Health()
    
    if health.Status == "healthy" {
        w.WriteHeader(200)
    } else {
        w.WriteHeader(503)
    }
    
    json.NewEncoder(w).Encode(health)
}
```

## Best Practices

### **1. Repository Pattern**

```go
// ✅ Good: Use repository pattern
func (c *UserController) Show(id uint) (*User, error) {
    return c.userRepo.Find(id)
}

// ❌ Bad: Direct database access
func (c *UserController) Show(id uint) (*User, error) {
    var user User
    err := c.db.First(&user, id).Error
    return &user, err
}
```

### **2. Event-Driven Architecture**

```go
// ✅ Good: Use events for side effects
func (c *UserController) Store(request *CreateUserRequest) (*User, error) {
    user := &User{Name: request.Name, Email: request.Email}
    
    if err := c.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // Dispatch event for side effects
    c.eventDispatcher.Dispatch(&UserCreated{User: user})
    
    return user, nil
}

// ❌ Bad: Handle side effects in controller
func (c *UserController) Store(request *CreateUserRequest) (*User, error) {
    user := &User{Name: request.Name, Email: request.Email}
    
    if err := c.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // Side effects in controller (bad)
    c.mailer.SendWelcomeEmail(user)
    c.logger.LogUserCreation(user)
    c.analytics.TrackUserRegistration(user)
    
    return user, nil
}
```

### **3. Async Processing**

```go
// ✅ Good: Use async for non-critical operations
func (c *OrderController) Store(request *CreateOrderRequest) (*Order, error) {
    order := &Order{Items: request.Items}
    
    if err := c.orderRepo.Create(order); err != nil {
        return nil, err
    }
    
    // Async processing for non-critical operations
    c.queue.PushAsync(&ProcessOrderJob{OrderID: order.ID})
    c.eventDispatcher.DispatchAsync(&OrderPlaced{Order: order})
    
    return order, nil
}
```

### **4. Caching Strategy**

```go
// ✅ Good: Use cache for expensive operations
func (c *UserController) Index() ([]*User, error) {
    return c.cache.Remember("users.all", 3600, func() ([]*User, error) {
        return c.userRepo.FindAll()
    })
}

// ✅ Good: Invalidate cache on updates
func (c *UserController) Store(request *CreateUserRequest) (*User, error) {
    user := &User{Name: request.Name, Email: request.Email}
    
    if err := c.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // Invalidate cache
    c.cache.Tags("users").Flush()
    
    return user, nil
}
```

## Troubleshooting

### **1. Performance Issues**

```go
// Check if optimizations are enabled
if facades.Performance().IsOptimized() {
    fmt.Println("Optimizations are enabled")
} else {
    fmt.Println("Optimizations are disabled")
}

// Get detailed performance stats
stats := facades.Performance().GetDetailedStats()
fmt.Printf("Cache hit rate: %.2f%%\n", stats.CacheHitRate)
fmt.Printf("Error rate: %.2f%%\n", stats.ErrorRate)
```

### **2. Memory Issues**

```go
// Check memory usage
memory := facades.Performance().GetMemoryStats()
fmt.Printf("Heap usage: %d MB\n", memory.HeapUsage)
fmt.Printf("Goroutine count: %d\n", memory.GoroutineCount)

// Force garbage collection if needed
facades.Performance().ForceGC()
```

### **3. Context Timeouts**

```go
// Check context configuration
config := facades.Config().Get("context")
timeout := config["default_timeout"].(time.Duration)
fmt.Printf("Default timeout: %v\n", timeout)

// Increase timeout for specific operations
ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

user, err := c.userRepo.FindWithContext(ctx, 1)
```

## Next Steps

- [Configuration Reference](./CONFIGURATION.md) - All configuration options
- [Examples and Tutorials](./EXAMPLES.md) - Real-world examples
- [Performance Optimizations](./PERFORMANCE_OPTIMIZATIONS.md) - How optimizations work
- [Core Architecture](./CORE_ARCHITECTURE.md) - Detailed architecture 