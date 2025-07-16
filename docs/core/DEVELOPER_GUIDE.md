# Developer Guide

## Getting Started

This guide covers the essential patterns and best practices for building applications with the Laravel-inspired Go framework.

## Core Concepts

### Canonical APIs

The framework uses single, canonical constructors for all core services:

```go
import go_core "base_lara_go_project/app/core/go_core"

// Repository with automatic optimization selection
repo := go_core.NewRepository[User](db)

// Event bus with work stealing pool integration
eventBus := go_core.NewEventBus[any](wsp, ca, pgo)

// Cache with context-aware operations
cache := go_core.NewLocalCache[any]()

// Job dispatcher with goroutine pool management
dispatcher := go_core.NewJobDispatcher[any](queue, wsp, ca, pgo)

// HTTP optimizer with FastHTTP integration
optimizer := go_core.NewHTTPOptimizer(config)
```

### Smart Query System

The repository automatically selects the optimal query path:

```go
// FastPath: Direct queries for simple operations
user, err := repo.Find(1)
user, err := repo.FindBy("email", "user@example.com")

// BalancedPath: Prepared statements for complex queries
users, err := repo.Where(map[string]any{"active": true}).Get()
count, err := repo.Where(map[string]any{"role": "admin"}).Count()

// ComplexPath: Advanced operations
err := repo.Complex().BulkCreate(users)
err := repo.Complex().Transaction(func(tx Repository[User]) error {
    // Transaction operations
    return nil
})
```

## Service Provider System

The framework uses a comprehensive service provider system for dependency injection:

### Application Bootstrap

```go
// Initialize the global service container
container := go_core.NewContainer()

// Create provider manager
providerManager := laravel_providers.NewProviderManager(container)

// Register the main AppServiceProvider
appProvider := &providers.AppServiceProvider{}
if err := providerManager.Register(appProvider); err != nil {
    panic(err)
}

// Boot all providers
if err := providerManager.Boot(); err != nil {
    panic(err)
}
```

### Service Resolution

```go
// Resolve services from container
repoInstance, err := container.Resolve("repository.user")
userRepo := repoInstance.(*repositories.UserRepository)

cacheInstance, err := container.Resolve("cache")
cache := cacheInstance.(go_core.Cache[any])

eventDispatcher, err := container.Resolve("event_dispatcher")
dispatcher := eventDispatcher.(go_core.EventDispatcher[any])
```

## Repository Usage

### Basic CRUD Operations

```go
type User struct {
    ID        uint      `json:"id" db:"id"`
    Name      string    `json:"name" db:"name"`
    Email     string    `json:"email" db:"email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
    DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// Implement required methods
func (u *User) GetTableName() string { return "users" }
func (u *User) GetPrimaryKey() string { return "id" }
func (u *User) GetFillableFields() []string { return []string{"name", "email"} }

// Create repository
repo := go_core.NewRepository[User](db)

// Create
user := &User{Name: "John Doe", Email: "john@example.com"}
err := repo.Create(user)

// Read
user, err := repo.Find(1)
user, err := repo.FindBy("email", "john@example.com")

// Update
user.Name = "Jane Doe"
err := repo.Update(user)

// Delete (soft delete)
err := repo.Delete(1)
```

### Query Building

```go
// Simple conditions
users, err := repo.Where(map[string]any{
    "active": true,
    "role": "admin",
}).Get()

// Complex queries with operators
users, err := repo.Where(map[string]any{
    "created_at_gt": time.Now().AddDate(0, 0, -7),
    "email_like": "%@example.com",
}).Get()

// Pagination
users, err := repo.Where(map[string]any{}).
    Limit(10).
    Offset(20).
    Get()

// Ordering
users, err := repo.Where(map[string]any{}).
    OrderBy("created_at", "desc").
    Get()
```

### Context-Aware Operations

```go
// Context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Context-aware repository
ctxRepo := repo.WithContext(ctx)
user, err := ctxRepo.Find(1)

// Context-aware event dispatching
err := eventBus.WithContext(ctx).Dispatch("user.created", user)
```

## FastHTTP Integration

The framework uses FastHTTP for high-performance HTTP handling with Gin compatibility:

### Router Setup

```go
// RouterServiceProvider automatically sets up FastHTTP integration
router := gin.Default()

// Add routes
router.GET("/api/users", func(c *gin.Context) {
    users, _ := repo.FindAll()
    c.JSON(200, users)
})

// FastHTTP server is automatically configured
// Routes are loaded from api/routes/ directory
```

### HTTP Optimization Configuration

```go
// HTTP optimization settings in config/http.go
httpConfig := map[string]any{
    "enable_fasthttp":        true,
    "read_timeout":          30,
    "write_timeout":         30,
    "max_connections":       15000,
    "enable_compression":    true,
    "zero_copy_enabled":     true,
}
```

## Event System

### Event Dispatching

```go
// Create event bus with optimizations
eventBus := go_core.NewEventBus[any](wsp, ca, pgo)

// Register listeners
eventBus.AddListener("user.created", func(event interface{}) error {
    user := event.(*User)
    // Handle user creation
    return nil
})

// Dispatch events
err := eventBus.Dispatch("user.created", user)

// Async dispatch
err := eventBus.DispatchAsync("user.created", user)
```

### Event Store

```go
// Store events for later retrieval
eventStore := go_core.NewMemoryEventStore[any]()

// Store event
err := eventStore.Store("user.created", user)

// Retrieve events
events, err := eventStore.GetByEventName("user.created")
events, err := eventStore.GetByTimeRange(start, end)
```

## Cache System

### Basic Caching

```go
// Create cache
cache := go_core.NewLocalCache[any]()

// Set value
err := cache.Set("user:1", user, 30*time.Minute)

// Get value
var user User
err := cache.Get("user:1", &user)

// Delete value
err := cache.Delete("user:1")

// Check existence
exists, err := cache.Exists("user:1")
```

### Advanced Caching

```go
// Get or set pattern
user, err := cache.GetOrSet("user:1", func() (*User, error) {
    return repo.Find(1)
}, 30*time.Minute)

// Batch operations
err := cache.SetMany(map[string]interface{}{
    "user:1": user1,
    "user:2": user2,
}, 30*time.Minute)

// Pattern deletion
err := cache.DeletePattern("user:*")
```

## Job System

### Job Dispatching

```go
// Create job dispatcher with optimizations
dispatcher := go_core.NewJobDispatcher[any](queue, wsp, ca, pgo)

// Define job
type EmailJob struct {
    To      string
    Subject string
    Body    string
}

func (j *EmailJob) Execute() error {
    // Send email logic
    return nil
}

// Dispatch job
err := dispatcher.Dispatch(&EmailJob{
    To:      "user@example.com",
    Subject: "Welcome",
    Body:    "Welcome to our platform!",
})

// Async dispatch
err := dispatcher.DispatchAsync(&EmailJob{...})
```

## Configuration

### Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=3306
DB_DATABASE=myapp
DB_USERNAME=root
DB_PASSWORD=password

# Cache
CACHE_DRIVER=redis
CACHE_HOST=localhost
CACHE_PORT=6379

# HTTP
HTTP_ENABLE_FASTHTTP=true
HTTP_MAX_CONNECTIONS=15000
HTTP_READ_TIMEOUT=30

# Goroutine
GOROUTINE_MAX_WORKERS=100
GOROUTINE_QUEUE_SIZE=1000
```

### Config Files

```go
// Load configuration
dbConfig, err := config.Load("database")
cacheConfig, err := config.Load("cache")
httpConfig, err := config.Load("http")
goroutineConfig, err := config.Load("goroutine")
```

### Using Config Facade

```go
import facades_core "base_lara_go_project/app/core/laravel_core/facades"

// Using the config facade
config := facades_core.Config()
appName := config.GetString("app.name")

// Or use the global functions
appName := facades_core.GetString("app.name")
debugMode := facades_core.GetBool("app.debug")
maxConnections := facades_core.GetInt("http.max_connections")
```

## HTTP Controllers

### Base Controller

```go
import laravel_http "base_lara_go_project/app/core/laravel_core/http"

type UserController struct {
    laravel_http.BaseController
    userRepo *repositories.UserRepository
}

func NewUserController(userRepo *repositories.UserRepository) *UserController {
    return &UserController{
        userRepo: userRepo,
    }
}

func (c *UserController) Index(ctx *gin.Context) {
    users, err := c.userRepo.FindAll(1, 10)
    if err != nil {
        c.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch users", err)
        return
    }
    
    c.SuccessResponse(ctx, users, "Users retrieved successfully")
}
```

## Middleware

### JWT Middleware

```go
func JWTMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
            c.Abort()
            return
        }
        
        // Validate token
        // Set user in context
        c.Next()
    }
}
```

## Testing

### Unit Tests

```go
func TestUserRepository_Find(t *testing.T) {
    // Create mock database
    db, mock, err := sqlmock.New()
    require.NoError(t, err)
    defer db.Close()
    
    // Create repository
    repo := go_core.NewRepository[User](db)
    
    // Set up expectations
    mock.ExpectQuery("SELECT (.+) FROM users WHERE id = (.+) AND deleted_at IS NULL").
        WithArgs(1).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email"}).
            AddRow(1, "John Doe", "john@example.com"))
    
    // Execute test
    user, err := repo.Find(1)
    
    // Assertions
    require.NoError(t, err)
    assert.Equal(t, "John Doe", user.Name)
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

### Integration Tests

```go
func TestUserRepositoryIntegration(t *testing.T) {
    // Use real database with config-driven connection
    dbConfig := config.DatabaseConfig()
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
        dbConfig["username"], dbConfig["password"],
        dbConfig["host"], dbConfig["port"], dbConfig["database"])
    
    db, err := sql.Open("mysql", dsn)
    require.NoError(t, err)
    defer db.Close()
    
    // Create repository
    repo := go_core.NewRepository[User](db)
    
    // Create test user
    user := &User{Name: "Test User", Email: "test@example.com"}
    err = repo.Create(user)
    require.NoError(t, err)
    
    // Test find
    found, err := repo.Find(user.ID)
    require.NoError(t, err)
    assert.Equal(t, user.Name, found.Name)
    
    // Cleanup
    repo.Delete(user.ID)
}
```

## Best Practices

### 1. Use Canonical Constructors
Always use the single canonical constructor for each service type:
```go
// ✅ Correct
repo := go_core.NewRepository[User](db)
eventBus := go_core.NewEventBus[any](wsp, ca, pgo)

// ❌ Avoid legacy constructors
// repo := go_core.NewInfrastructureOptimizedRepository[User](db)
```

### 2. Leverage Smart Query System
Let the framework automatically select the optimal query path:
```go
// ✅ Let framework choose optimal path
user, err := repo.Find(1)                    // FastPath
users, err := repo.Where(conds).Get()        // BalancedPath
err := repo.Complex().BulkCreate(users)      // ComplexPath
```

### 3. Use Context for Timeouts
Always provide context for operations that might take time:
```go
// ✅ Context-aware operations
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
result, err := repo.WithContext(ctx).Find(1)
```

### 4. Configure via Environment
Use environment variables and config files for customization:
```go
// ✅ Environment-driven configuration
// Set HTTP_MAX_CONNECTIONS=15000 in .env
maxConn := facades_core.GetInt("http.max_connections", 10000)
```

### 5. Use Service Providers
Register services through the service provider system:
```go
// ✅ Service provider registration
container.Singleton("repository.user", func() (any, error) {
    return repositories.NewUserRepository(repo, cache), nil
})
```

## Common Patterns

### Repository Pattern with Cache

```go
type UserRepository struct {
    repository go_core.Repository[models.User]
    cache      go_core.Cache[models.User]
}

func (r *UserRepository) FindWithCache(id uint) (*models.User, error) {
    // Check cache first
    if cached, err := r.cache.Get(fmt.Sprintf("user:%d", id)); err == nil {
        return cached, nil
    }
    
    // Use fast path for database lookup
    user, err := r.repository.Find(id)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    if user != nil {
        r.cache.Set(fmt.Sprintf("user:%d", id), user, 5*time.Minute)
    }
    
    return user, nil
}
```

### Event-Driven Architecture

```go
// Define event
type UserCreated struct {
    User *models.User
}

func (e *UserCreated) GetName() string { return "user.created" }
func (e *UserCreated) GetData() interface{} { return e.User }

// Register listener
eventBus.AddListener("user.created", func(event interface{}) error {
    user := event.(*models.User)
    // Send welcome email
    return nil
})

// Dispatch event
err := eventBus.Dispatch("user.created", &UserCreated{User: user})
``` 