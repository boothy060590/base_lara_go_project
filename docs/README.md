# Laravel-Inspired Go Framework Documentation

## Overview

This is a high-performance, Laravel-inspired Go framework designed to provide familiar developer experience while leveraging Go's performance characteristics. The framework features automatic optimizations, smart query routing, and zero-configuration setup.

## Architecture

### Core Components

- **Go Core (`api/app/core/go_core/`)**: High-performance foundation with automatic optimizations
- **Laravel Core (`api/app/core/laravel_core/`)**: Laravel-style developer experience layer
- **Smart Query System**: Three-tier optimization (FastPath, BalancedPath, ComplexPath)
- **Canonical APIs**: Single constructors with automatic optimization selection
- **FastHTTP Integration**: High-performance HTTP server with Gin routing adapter
- **Service Provider System**: Comprehensive dependency injection and configuration

### Key Features

- **Automatic Optimizations**: Goroutine pools, context management, object pools
- **Smart Query Routing**: Automatic selection of optimal database query path
- **Config-Driven**: Environment-specific customization without code changes
- **Context-Aware**: Proper cancellation and timeout handling
- **Type Safety**: Generic implementations with compile-time checking
- **FastHTTP Server**: High-performance HTTP handling with Gin compatibility
- **Comprehensive Service Providers**: Database, cache, events, queue, mail, logging, jobs

## Technology Stack

### Backend
- **Go 1.24+**: Core application language
- **FastHTTP**: High-performance HTTP server with Gin routing adapter
- **Custom Repository Layer**: Optimized SQL with connection pooling and statement caching
- **ElasticMQ**: Message queue (SQS-compatible)
- **JWT**: Authentication tokens

### Frontend
- **Vue.js 3**: Progressive JavaScript framework
- **Vite**: Fast build tool and development server
- **SCSS**: Advanced CSS preprocessing

## Documentation Structure

### [Core Architecture](./core/ARCHITECTURE.md)
Comprehensive overview of the framework's layered architecture, smart query system, and optimization strategies.

### [Developer Guide](./core/DEVELOPER_GUIDE.md)
Getting started guide, canonical APIs, and best practices for building applications.

### [Configuration](./core/CONFIGURATION.md)
Config system overview, environment variables, and profile-based configurations.

### [Testing](./core/TESTING.md)
Test suite structure, running tests, and integration testing guidelines.

### [Examples](./core/EXAMPLES.md)
Code examples and common patterns for using the framework.

### [Setup](./setup/)
Environment setup, Docker configuration, and deployment guides.

## Quick Start

```go
// Create a repository with automatic optimizations
repo := go_core.NewRepository[User](db)

// Smart query routing - automatically selects optimal path
user, err := repo.Find(1)                    // FastPath: Direct query
users, err := repo.Where(conds).Get()        // BalancedPath: Prepared statement
err := repo.Complex().BulkCreate(users)      // ComplexPath: Advanced operations

// Context-aware operations
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
result, err := repo.WithContext(ctx).Find(1)

// FastHTTP server with Gin routing
router := gin.Default()
router.GET("/api/users", func(c *gin.Context) {
    users, _ := repo.FindAll()
    c.JSON(200, users)
})

// Service provider registration
container := go_core.NewContainer()
provider := &AppServiceProvider{}
provider.Register(container)
provider.Boot(container)
```

## Current Status

- ✅ **Core Architecture**: Complete with smart query system
- ✅ **Canonical APIs**: Single constructors with automatic optimizations
- ✅ **FastHTTP Integration**: High-performance HTTP server with Gin adapter
- ✅ **Service Provider System**: Comprehensive dependency injection
- ✅ **Config System**: Environment-driven configuration with automatic discovery
- ✅ **Test Suite**: Comprehensive coverage with real and mocked scenarios
- ✅ **Repository Layer**: Custom SQL implementation with connection pooling
- 🔄 **Memory Optimization**: In progress (Phase 3)
- 📋 **Authentication**: Planned
- 📋 **WebSocket Support**: Planned
- 📋 **GraphQL Integration**: Planned

## Performance Goals

The framework aims to provide significant performance improvements over traditional web frameworks through:

- **Smart Query Routing**: Automatic optimization selection based on query complexity
- **FastHTTP Server**: High-performance HTTP handling with zero-copy operations
- **Connection Pooling**: Efficient database connection management
- **Statement Caching**: Reduced query preparation overhead
- **Context Optimization**: Proper resource cleanup and cancellation
- **Memory Optimization**: Object pools and zero-copy operations (in progress)

*Note: Specific performance benchmarks will be published once comprehensive testing is complete.*

## Service Provider Architecture

The framework uses a comprehensive service provider system for dependency injection:

```go
// Core Service Providers
- AppServiceProvider: Main application provider
- ConfigServiceProvider: Configuration system
- DatabaseServiceProvider: Database connections
- CacheServiceProvider: Cache stores
- EventServiceProvider: Event system
- QueueServiceProvider: Job queues
- MailServiceProvider: Email services
- LoggingServiceProvider: Logging channels
- JobServiceProvider: Background jobs
- HTTPOptimizationServiceProvider: FastHTTP integration
- GoroutineServiceProvider: Concurrency optimization
- ContextServiceProvider: Context management
```

## Contributing

See the [Developer Guide](./core/DEVELOPER_GUIDE.md) for contribution guidelines and development setup.

## License

[Add your license information here] 