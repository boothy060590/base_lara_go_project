# Core Architecture

## Overview

The framework follows a layered architecture with automatic optimizations and smart query routing. The design prioritizes developer experience while delivering high performance through intelligent optimization selection.

## Architecture Layers

### 1. Go Core (`api/app/core/go_core/`)
High-performance foundation with automatic optimizations:

- **Smart Query System**: Three-tier optimization routing
- **Canonical Constructors**: Single APIs with automatic optimization selection
- **Context Integration**: Proper cancellation and timeout handling
- **Performance Features**: Goroutine pools, object pools, connection pooling
- **FastHTTP Integration**: High-performance HTTP server with Gin adapter
- **Statement Caching**: Prepared statement management for database operations

### 2. Laravel Core (`api/app/core/laravel_core/`)
Laravel-style developer experience layer:

- **Facades**: Familiar Laravel-style access patterns
- **Service Providers**: Comprehensive dependency injection and configuration
- **Eloquent Models**: ORM-like interface on top of Go Core
- **HTTP Layer**: Request/response handling with middleware
- **Config System**: Environment-driven configuration with automatic discovery

## Smart Query System

The framework implements a three-tier query optimization system that automatically selects the optimal path for database operations:

### FastPath (Direct Queries)
- **Use Case**: Simple, single-record operations
- **Example**: `repo.Find(1)`, `repo.FindBy("email", "user@example.com")`
- **Optimization**: Direct SQL queries with parameter binding
- **Performance**: Minimal overhead, maximum speed
- **Implementation**: `FastPathExecutor` with minimal statement caching

### BalancedPath (Prepared Statements)
- **Use Case**: Complex queries with conditions
- **Example**: `repo.Where(conditions).Get()`, `repo.Where(conditions).Count()`
- **Optimization**: Prepared statements with caching
- **Performance**: Statement reuse, SQL injection protection
- **Implementation**: `BalancedPathExecutor` with statement cache and connection pooling

### ComplexPath (Advanced Operations)
- **Use Case**: Bulk operations, transactions, complex joins
- **Example**: `repo.Complex().BulkCreate(models)`, `repo.Complex().Transaction()`
- **Optimization**: Advanced query building, transaction management
- **Performance**: Optimized for complex scenarios
- **Implementation**: `ComplexPathExecutor` with full optimization suite

## FastHTTP Integration

The framework uses FastHTTP for high-performance HTTP handling while maintaining Gin compatibility:

### FastHTTP Adapter
```go
// FastHTTPAdapter adapts fasthttp requests to standard http.Handler interface
type FastHTTPAdapter struct {
    handler     http.Handler
    bufferPool  *sync.Pool
    metrics     *HTTPMetrics
    config      *HTTPOptimizationConfig
}
```

### HTTP Optimization Features
- **Zero-Copy Operations**: Minimize memory allocations
- **Connection Pooling**: Efficient HTTP connection management
- **CORS Support**: Built-in CORS handling
- **Performance Metrics**: Request tracking and monitoring
- **Gin Compatibility**: Seamless integration with existing Gin routes

### Router Integration
```go
// RouterServiceProvider integrates FastHTTP with Gin routing
type RouterServiceProvider struct {
    httpOptimizer *app_core.HTTPOptimizer
}

// Routes are loaded from api/routes/ directory
// FastHTTP server handles requests with Gin middleware support
```

## Service Provider System

The framework uses a comprehensive service provider system for dependency injection:

### Core Service Providers
- **AppServiceProvider**: Main application provider that registers all core providers
- **ConfigServiceProvider**: Configuration system with automatic discovery
- **DatabaseServiceProvider**: Database connections and repository registration
- **CacheServiceProvider**: Cache stores with Redis and local options
- **EventServiceProvider**: Event system with work stealing pool integration
- **QueueServiceProvider**: Job queues with ElasticMQ integration
- **MailServiceProvider**: Email services with template support
- **LoggingServiceProvider**: Logging channels with Sentry integration
- **JobServiceProvider**: Background jobs with goroutine optimization
- **HTTPOptimizationServiceProvider**: FastHTTP server configuration
- **GoroutineServiceProvider**: Concurrency optimization
- **ContextServiceProvider**: Context management and optimization

### Service Registration Pattern
```go
// Automatic registration in AppServiceProvider
func (p *AppServiceProvider) Register(container *app_core.Container) error {
    coreProviders := []ServiceProvider{
        &CoreServiceProvider{},
        &ConfigServiceProvider{},
        &DatabaseServiceProvider{},
        &CacheServiceProvider{},
        &EventServiceProvider{},
        &QueueServiceProvider{},
        &MailServiceProvider{},
        &LoggingServiceProvider{},
        &JobServiceProvider{},
        &HTTPOptimizationServiceProvider{},
        &GoroutineServiceProvider{},
        &ContextServiceProvider{},
    }
    
    for _, provider := range coreProviders {
        provider.Register(container)
    }
    return nil
}
```

## Canonical Constructor Pattern

All core services use single, canonical constructors that automatically provide optimized implementations:

```go
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

### Benefits
- **No Confusion**: Single constructor per service type
- **Automatic Optimization**: Framework selects optimal implementation
- **Consistent APIs**: Same interface across all optimizations
- **Zero Configuration**: Sensible defaults that work out of the box

## Context Integration

All operations support context for proper cancellation and timeout handling:

```go
// Context-aware repository operations
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := repo.WithContext(ctx).Find(1)

// Context-aware event dispatching
err := eventBus.WithContext(ctx).Dispatch("user.created", event)

// Context-aware job processing
err := dispatcher.WithContext(ctx).Dispatch(job)
```

## Configuration Strategy

### Environment-Driven Configuration
```go
// Automatic config loading from api/config/
dbConfig, _ := config.Load("database")
cacheConfig, _ := config.Load("cache")
goroutineConfig, _ := config.Load("goroutine")
httpConfig, _ := config.Load("http")
```

### Config File Structure
- **app.go**: Application settings (name, debug, port)
- **database.go**: Database connections and settings
- **cache.go**: Cache stores and TTL settings
- **http.go**: FastHTTP optimization settings
- **queue.go**: Queue and worker configuration
- **mail.go**: Mail configuration
- **logging.go**: Logging channels and handlers

### Profile-Based Configurations
- **Web Apps**: Low latency, fast response times (30s timeouts)
- **APIs**: Moderate timeouts, high throughput (60s timeouts)
- **Background Jobs**: Long timeouts, high performance (300s timeouts)
- **Streaming**: Very long timeouts, large buffers (1800s timeouts)
- **Batch Processing**: Long timeouts, large buffers (1800s timeouts)

## Performance Optimizations

### Automatic Optimizations
- **Goroutine Pools**: Work stealing pools for concurrent operations
- **Object Pools**: Reusable objects for high-frequency operations
- **Connection Pooling**: Efficient database connection management
- **Statement Caching**: Pre-prepared queries for repeated operations
- **FastHTTP Server**: High-performance HTTP handling
- **Zero-Copy Operations**: Minimize memory allocations

### Memory Optimization (Phase 3 - In Progress)
- **Object Pools**: For JSON encoding/decoding and frequently allocated types
- **Zero-Copy Operations**: Minimize memory allocations
- **Memory-Mapped Files**: For large data sets
- **NUMA-Aware Allocation**: For multi-socket systems

## Type Safety & Generics

The framework leverages Go's generics for type-safe implementations:

```go
// Generic repository with type safety
type Repository[T any] interface {
    Find(id uint) (*T, error)
    FindBy(field, value string) (*T, error)
    Where(conditions map[string]any) SmartQuery[T]
    Create(model *T) error
    Update(model *T) error
    Delete(id uint) error
}

// Generic cache with type safety
type Cache[T any] interface {
    Get(key string) (*T, error)
    Set(key string, value *T, ttl time.Duration) error
    Delete(key string) error
}

// Generic event bus with type safety
type EventDispatcher[T any] interface {
    Dispatch(event Event[T]) error
    DispatchAsync(event Event[T]) error
}
```

## Separation of Concerns

### Go Core Responsibilities
- Infrastructure concerns only
- Performance optimizations
- Database operations
- Concurrency management
- Resource management
- HTTP optimization

### Laravel Core Responsibilities
- Developer experience patterns
- Facades and service providers
- HTTP request/response handling
- Middleware and routing
- Application-level abstractions
- Configuration management

### No Business Logic in Core
- Core systems are reusable across applications
- No application-specific dependencies
- Framework remains generic and extensible

## Testing Architecture

### Unit Tests (Mocked)
- **Location**: `api/app/core/tests/go_core/unit/`
- **Purpose**: Test individual components in isolation
- **Database**: Mocked with sqlmock
- **Coverage**: All core functionality

### Integration Tests (Real DB)
- **Location**: `api/app/core/tests/go_core/integration/`
- **Purpose**: Test component interactions with real database
- **Database**: Real MySQL with config-driven connections
- **Coverage**: End-to-end workflows

### Test Consolidation
- **Single Integration Suite**: All real DB tests in one place
- **Config-Driven**: Environment variable and config file support
- **Robust Cleanup**: Automatic test data cleanup
- **Comprehensive Coverage**: All core modules tested

## Application Structure

```
api/
├── app/
│   ├── core/
│   │   ├── go_core/           # High-performance foundation
│   │   └── laravel_core/      # Laravel-style developer experience
│   ├── providers/             # Application service providers
│   ├── repositories/          # Repository implementations
│   ├── models/               # Database models
│   ├── events/               # Event definitions
│   ├── jobs/                 # Background jobs
│   ├── listeners/            # Event listeners
│   └── http/                 # HTTP controllers and middleware
├── config/                   # Configuration files
├── routes/                   # Route definitions
├── bootstrap/                # Application bootstrap
└── database/                 # Database migrations
``` 