# Examples and Tutorials

## Overview

This document provides real-world examples and tutorials showing how to use the Laravel-Inspired Go Framework to build high-performance applications. All examples demonstrate automatic optimization without requiring any additional configuration.

## Quick Start Tutorial

### **Building a User Management System**

Let's build a complete user management system with automatic optimization.

#### **1. Project Structure**

```
myapp/
├── main.go
├── app/
│   ├── models/
│   │   └── user.go
│   ├── controllers/
│   │   └── user_controller.go
│   ├── repositories/
│   │   └── user_repository.go
│   ├── events/
│   │   └── user_created.go
│   ├── listeners/
│   │   └── send_welcome_email.go
│   ├── jobs/
│   │   └── process_user_registration.go
│   └── providers/
│       └── app_service_provider.go
└── config/
    └── app.go
```

#### **2. Main Application**

```go
// main.go
package main

import (
    "log"
    "net/http"
    "your-project/api/app/core/laravel_core/facades"
)

func main() {
    // Framework automatically initializes all optimizations
    app := facades.App()
    
    // Register routes
    http.HandleFunc("/users", handleUsers)
    http.HandleFunc("/users/", handleUser)
    
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
    controller := NewUserController()
    
    switch r.Method {
    case "GET":
        users, err := controller.Index()
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        // Return JSON response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
        
    case "POST":
        var request CreateUserRequest
        if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
            http.Error(w, err.Error(), 400)
            return
        }
        
        user, err := controller.Store(&request)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(user)
    }
}

func handleUser(w http.ResponseWriter, r *http.Request) {
    controller := NewUserController()
    
    // Extract user ID from URL
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 3 {
        http.Error(w, "Invalid user ID", 400)
        return
    }
    
    id, err := strconv.ParseUint(parts[2], 10, 32)
    if err != nil {
        http.Error(w, "Invalid user ID", 400)
        return
    }
    
    user, err := controller.Show(uint(id))
    if err != nil {
        http.Error(w, err.Error(), 404)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}
```

#### **3. User Model**

```go
// app/models/user.go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name" gorm:"not null"`
    Email     string    `json:"email" gorm:"unique;not null"`
    Active    bool      `json:"active" gorm:"default:true"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

func (u *User) TableName() string {
    return "users"
}
```

#### **4. User Repository**

```go
// app/repositories/user_repository.go
package repositories

import (
    "your-project/api/app/core/go_core"
    "your-project/api/app/models"
)

type UserRepository struct {
    *go_core.Repository[models.User]
}

func NewUserRepository() *UserRepository {
    return &UserRepository{
        Repository: go_core.NewRepository[models.User](db),
    }
}

// Custom methods with automatic optimization
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := r.db.Where("email = ?", email).First(&user).Error
    return &user, err
}

func (r *UserRepository) FindActiveUsers() ([]*models.User, error) {
    var users []*models.User
    err := r.db.Where("active = ?", true).Find(&users).Error
    return users, err
}

func (r *UserRepository) FindManyByIDs(ids []uint) ([]*models.User, error) {
    var users []*models.User
    err := r.db.Where("id IN ?", ids).Find(&users).Error
    return users, err
}
```

#### **5. User Controller**

```go
// app/controllers/user_controller.go
package controllers

import (
    "your-project/api/app/core/laravel_core/facades"
    "your-project/api/app/models"
    "your-project/api/app/repositories"
)

type UserController struct {
    userRepo         *repositories.UserRepository
    eventDispatcher  *facades.Event
    cache           *facades.Cache
}

func NewUserController() *UserController {
    return &UserController{
        userRepo:        repositories.NewUserRepository(),
        eventDispatcher: facades.Event(),
        cache:          facades.Cache(),
    }
}

func (c *UserController) Index() ([]*models.User, error) {
    // Framework automatically optimizes this with caching and goroutines
    return c.cache.Remember("users.all", 3600, func() ([]*models.User, error) {
        return c.userRepo.FindAll()
    })
}

func (c *UserController) Show(id uint) (*models.User, error) {
    // Framework automatically optimizes this query
    return c.userRepo.Find(id)
}

func (c *UserController) Store(request *CreateUserRequest) (*models.User, error) {
    user := &models.User{
        Name:  request.Name,
        Email: request.Email,
    }
    
    // Framework automatically optimizes this operation
    if err := c.userRepo.Create(user); err != nil {
        return nil, err
    }
    
    // Framework automatically optimizes this event dispatch
    c.eventDispatcher.Dispatch(&UserCreated{User: user})
    
    // Invalidate cache
    c.cache.Tags("users").Flush()
    
    return user, nil
}

func (c *UserController) Update(id uint, request *UpdateUserRequest) (*models.User, error) {
    user, err := c.userRepo.Find(id)
    if err != nil {
        return nil, err
    }
    
    user.Name = request.Name
    user.Email = request.Email
    
    // Framework automatically optimizes this operation
    if err := c.userRepo.Update(user); err != nil {
        return nil, err
    }
    
    // Framework automatically optimizes this event dispatch
    c.eventDispatcher.Dispatch(&UserUpdated{User: user})
    
    // Invalidate cache
    c.cache.Tags("users").Flush()
    
    return user, nil
}

func (c *UserController) Delete(id uint) error {
    user, err := c.userRepo.Find(id)
    if err != nil {
        return err
    }
    
    // Framework automatically optimizes this operation
    if err := c.userRepo.Delete(user); err != nil {
        return err
    }
    
    // Framework automatically optimizes this event dispatch
    c.eventDispatcher.Dispatch(&UserDeleted{User: user})
    
    // Invalidate cache
    c.cache.Tags("users").Flush()
    
    return nil
}
```

#### **6. Events and Listeners**

```go
// app/events/user_created.go
package events

import "your-project/api/app/models"

type UserCreated struct {
    User *models.User
}

// app/events/user_updated.go
package events

import "your-project/api/app/models"

type UserUpdated struct {
    User *models.User
}

// app/events/user_deleted.go
package events

import "your-project/api/app/models"

type UserDeleted struct {
    User *models.User
}

// app/listeners/send_welcome_email.go
package listeners

import (
    "your-project/api/app/core/laravel_core/facades"
    "your-project/api/app/events"
)

type SendWelcomeEmail struct {
    mailer *facades.Mail
}

func NewSendWelcomeEmail() *SendWelcomeEmail {
    return &SendWelcomeEmail{
        mailer: facades.Mail(),
    }
}

func (l *SendWelcomeEmail) Handle(event *events.UserCreated) error {
    // Framework automatically optimizes this email sending
    return l.mailer.Send("welcome", event.User.Email, map[string]interface{}{
        "user": event.User,
    })
}

// app/listeners/log_user_activity.go
package listeners

import (
    "log"
    "your-project/api/app/events"
)

type LogUserActivity struct{}

func (l *LogUserActivity) Handle(event *events.UserCreated) error {
    log.Printf("User created: %s (%s)", event.User.Name, event.User.Email)
    return nil
}

func (l *LogUserActivity) HandleUpdate(event *events.UserUpdated) error {
    log.Printf("User updated: %s (%s)", event.User.Name, event.User.Email)
    return nil
}

func (l *LogUserActivity) HandleDelete(event *events.UserDeleted) error {
    log.Printf("User deleted: %s (%s)", event.User.Name, event.User.Email)
    return nil
}

// app/listeners/update_user_count.go
package listeners

import (
    "your-project/api/app/core/laravel_core/facades"
    "your-project/api/app/events"
)

type UpdateUserCount struct {
    cache *facades.Cache
}

func NewUpdateUserCount() *UpdateUserCount {
    return &UpdateUserCount{
        cache: facades.Cache(),
    }
}

func (l *UpdateUserCount) Handle(event *events.UserCreated) error {
    // Increment user count
    l.cache.Increment("user_count")
    return nil
}

func (l *UpdateUserCount) HandleDelete(event *events.UserDeleted) error {
    // Decrement user count
    l.cache.Decrement("user_count")
    return nil
}
```

#### **7. Background Jobs**

```go
// app/jobs/process_user_registration.go
package jobs

import (
    "your-project/api/app/core/go_core"
    "your-project/api/app/models"
)

type ProcessUserRegistration struct {
    go_core.Job
    User *models.User
}

func (j *ProcessUserRegistration) Handle() error {
    // Process user registration in background
    // Framework automatically optimizes this with work stealing
    
    // Send welcome email
    if err := j.sendWelcomeEmail(); err != nil {
        return err
    }
    
    // Update analytics
    if err := j.updateAnalytics(); err != nil {
        return err
    }
    
    // Send notification to admin
    if err := j.notifyAdmin(); err != nil {
        return err
    }
    
    return nil
}

func (j *ProcessUserRegistration) sendWelcomeEmail() error {
    // Email sending logic
    return nil
}

func (j *ProcessUserRegistration) updateAnalytics() error {
    // Analytics update logic
    return nil
}

func (j *ProcessUserRegistration) notifyAdmin() error {
    // Admin notification logic
    return nil
}

func (j *ProcessUserRegistration) Failed(err error) {
    // Handle job failure
    log.Printf("Failed to process user registration: %v", err)
}
```

#### **8. Service Provider**

```go
// app/providers/app_service_provider.go
package providers

import (
    "your-project/api/app/core/go_core"
    "your-project/api/app/repositories"
    "your-project/api/app/listeners"
)

type AppServiceProvider struct {
    go_core.BaseServiceProvider
}

func (p *AppServiceProvider) Register(container *go_core.Container) error {
    // Register repositories
    container.Singleton("user.repository", func() (any, error) {
        return repositories.NewUserRepository(), nil
    })
    
    // Register listeners
    container.Singleton("listener.send_welcome_email", func() (any, error) {
        return listeners.NewSendWelcomeEmail(), nil
    })
    
    container.Singleton("listener.log_user_activity", func() (any, error) {
        return &listeners.LogUserActivity{}, nil
    })
    
    container.Singleton("listener.update_user_count", func() (any, error) {
        return listeners.NewUpdateUserCount(), nil
    })
    
    return nil
}

func (p *AppServiceProvider) Boot(container *go_core.Container) error {
    // Register event listeners
    eventDispatcher := container.Resolve("event.dispatcher").(*go_core.EventDispatcher)
    
    eventDispatcher.Listen("user.created", container.Resolve("listener.send_welcome_email").(*listeners.SendWelcomeEmail))
    eventDispatcher.Listen("user.created", container.Resolve("listener.log_user_activity").(*listeners.LogUserActivity))
    eventDispatcher.Listen("user.created", container.Resolve("listener.update_user_count").(*listeners.UpdateUserCount))
    
    eventDispatcher.Listen("user.updated", container.Resolve("listener.log_user_activity").(*listeners.LogUserActivity))
    eventDispatcher.Listen("user.deleted", container.Resolve("listener.log_user_activity").(*listeners.LogUserActivity))
    eventDispatcher.Listen("user.deleted", container.Resolve("listener.update_user_count").(*listeners.UpdateUserCount))
    
    return nil
}
```

## Advanced Examples

### **E-Commerce System**

#### **Product Management with Caching**

```go
// app/controllers/product_controller.go
package controllers

import (
    "your-project/api/app/core/laravel_core/facades"
    "your-project/api/app/models"
)

type ProductController struct {
    productRepo *repositories.ProductRepository
    cache      *facades.Cache
    eventDispatcher *facades.Event
}

func (c *ProductController) Index() ([]*models.Product, error) {
    // Framework automatically optimizes with caching and goroutines
    return c.cache.Remember("products.all", 1800, func() ([]*models.Product, error) {
        return c.productRepo.FindAll()
    })
}

func (c *ProductController) Show(id uint) (*models.Product, error) {
    // Framework automatically optimizes with caching
    return c.cache.Remember(fmt.Sprintf("product:%d", id), 3600, func() (*models.Product, error) {
        return c.productRepo.Find(id)
    })
}

func (c *ProductController) Search(query string) ([]*models.Product, error) {
    // Framework automatically optimizes search with parallel processing
    return c.productRepo.SearchAsync(query)
}
```

#### **Order Processing with Events**

```go
// app/controllers/order_controller.go
package controllers

import (
    "your-project/api/app/core/laravel_core/facades"
    "your-project/api/app/models"
)

type OrderController struct {
    orderRepo *repositories.OrderRepository
    eventDispatcher *facades.Event
    queue *facades.Queue
}

func (c *OrderController) Store(request *CreateOrderRequest) (*models.Order, error) {
    order := &models.Order{
        UserID: request.UserID,
        Items:  request.Items,
        Total:  request.Total,
    }
    
    // Framework automatically optimizes this operation
    if err := c.orderRepo.Create(order); err != nil {
        return nil, err
    }
    
    // Framework automatically optimizes these event dispatches
    c.eventDispatcher.Dispatch(&OrderPlaced{Order: order})
    c.eventDispatcher.Dispatch(&InventoryUpdated{Order: order})
    c.eventDispatcher.Dispatch(&PaymentProcessed{Order: order})
    
    // Queue background jobs
    c.queue.PushAsync(&ProcessOrderJob{OrderID: order.ID})
    c.queue.PushAsync(&SendOrderConfirmationJob{OrderID: order.ID})
    c.queue.PushAsync(&UpdateInventoryJob{OrderID: order.ID})
    
    return order, nil
}
```

### **API Gateway with Rate Limiting**

```go
// app/middleware/rate_limit_middleware.go
package middleware

import (
    "net/http"
    "your-project/api/app/core/laravel_core/facades"
)

type RateLimitMiddleware struct {
    cache *facades.Cache
}

func (m *RateLimitMiddleware) Handle(w http.ResponseWriter, r *http.Request, next func()) {
    // Get client IP
    clientIP := r.RemoteAddr
    
    // Check rate limit
    key := fmt.Sprintf("rate_limit:%s", clientIP)
    count, err := m.cache.Increment(key)
    if err != nil {
        http.Error(w, "Rate limit error", 500)
        return
    }
    
    // Set expiration for first request
    if count == 1 {
        m.cache.Expire(key, 60) // 1 minute
    }
    
    // Check if limit exceeded
    if count > 100 { // 100 requests per minute
        http.Error(w, "Rate limit exceeded", 429)
        return
    }
    
    // Continue to next middleware/controller
    next()
}
```

### **Real-Time Chat System**

```go
// app/controllers/chat_controller.go
package controllers

import (
    "your-project/api/app/core/laravel_core/facades"
    "your-project/api/app/models"
)

type ChatController struct {
    messageRepo *repositories.MessageRepository
    eventDispatcher *facades.Event
    cache *facades.Cache
}

func (c *ChatController) SendMessage(request *SendMessageRequest) (*models.Message, error) {
    message := &models.Message{
        UserID:    request.UserID,
        RoomID:    request.RoomID,
        Content:   request.Content,
        Timestamp: time.Now(),
    }
    
    // Framework automatically optimizes this operation
    if err := c.messageRepo.Create(message); err != nil {
        return nil, err
    }
    
    // Framework automatically optimizes this event dispatch
    c.eventDispatcher.Dispatch(&MessageSent{Message: message})
    
    // Cache recent messages
    c.cache.Remember(fmt.Sprintf("room:%d:messages", request.RoomID), 300, func() ([]*models.Message, error) {
        return c.messageRepo.FindByRoom(request.RoomID, 50)
    })
    
    return message, nil
}

func (c *ChatController) GetMessages(roomID uint) ([]*models.Message, error) {
    // Framework automatically optimizes with caching
    return c.cache.Remember(fmt.Sprintf("room:%d:messages", roomID), 300, func() ([]*models.Message, error) {
        return c.messageRepo.FindByRoom(roomID, 50)
    })
}
```

### **File Upload System**

```go
// app/controllers/file_controller.go
package controllers

import (
    "your-project/api/app/core/laravel_core/facades"
    "your-project/api/app/models"
)

type FileController struct {
    fileRepo *repositories.FileRepository
    queue *facades.Queue
    eventDispatcher *facades.Event
}

func (c *FileController) Upload(file *multipart.FileHeader) (*models.File, error) {
    // Process file upload
    fileModel := &models.File{
        Name:     file.Filename,
        Size:     file.Size,
        MimeType: file.Header.Get("Content-Type"),
        Path:     fmt.Sprintf("/uploads/%s", file.Filename),
    }
    
    // Framework automatically optimizes this operation
    if err := c.fileRepo.Create(fileModel); err != nil {
        return nil, err
    }
    
    // Queue background processing
    c.queue.PushAsync(&ProcessFileJob{FileID: fileModel.ID})
    c.queue.PushAsync(&GenerateThumbnailJob{FileID: fileModel.ID})
    c.queue.PushAsync(&ScanVirusJob{FileID: fileModel.ID})
    
    // Framework automatically optimizes this event dispatch
    c.eventDispatcher.Dispatch(&FileUploaded{File: fileModel})
    
    return fileModel, nil
}
```

## Performance Testing Examples

### **Load Testing**

```go
// test/load_test.go
package test

import (
    "testing"
    "your-project/api/app/core/laravel_core/facades"
)

func TestUserCreationLoad(t *testing.T) {
    // Test concurrent user creation
    const numUsers = 1000
    const concurrency = 10
    
    // Create work channel
    workChan := make(chan int, numUsers)
    results := make(chan error, numUsers)
    
    // Start workers
    for i := 0; i < concurrency; i++ {
        go func() {
            for work := range workChan {
                // Create user
                user := &models.User{
                    Name:  fmt.Sprintf("User %d", work),
                    Email: fmt.Sprintf("user%d@example.com", work),
                }
                
                // Framework automatically optimizes this
                err := userRepo.Create(user)
                results <- err
            }
        }()
    }
    
    // Send work
    for i := 0; i < numUsers; i++ {
        workChan <- i
    }
    close(workChan)
    
    // Collect results
    for i := 0; i < numUsers; i++ {
        if err := <-results; err != nil {
            t.Errorf("Failed to create user: %v", err)
        }
    }
}
```

### **Cache Performance Test**

```go
// test/cache_test.go
package test

import (
    "testing"
    "your-project/api/app/core/laravel_core/facades"
)

func TestCachePerformance(t *testing.T) {
    cache := facades.Cache()
    
    // Test cache set performance
    start := time.Now()
    for i := 0; i < 10000; i++ {
        err := cache.Set(fmt.Sprintf("key:%d", i), fmt.Sprintf("value:%d", i), 3600)
        if err != nil {
            t.Errorf("Failed to set cache: %v", err)
        }
    }
    setDuration := time.Since(start)
    
    // Test cache get performance
    start = time.Now()
    for i := 0; i < 10000; i++ {
        _, err := cache.Get(fmt.Sprintf("key:%d", i))
        if err != nil {
            t.Errorf("Failed to get cache: %v", err)
        }
    }
    getDuration := time.Since(start)
    
    t.Logf("Cache set performance: %v for 10,000 operations", setDuration)
    t.Logf("Cache get performance: %v for 10,000 operations", getDuration)
}
```

## Next Steps

- [Developer Guide](./DEVELOPER_GUIDE.md) - How to use the framework
- [Configuration Reference](./CONFIGURATION.md) - All configuration options
- [Performance Optimizations](./PERFORMANCE_OPTIMIZATIONS.md) - How optimizations work
- [Core Architecture](./CORE_ARCHITECTURE.md) - Detailed architecture 