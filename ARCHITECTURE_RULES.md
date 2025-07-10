# Architecture Rules & Implementation Strategy

## Core Philosophy

**Goal**: Build a Laravel-inspired Go framework that provides familiar developer experience while delivering orders of magnitude better performance through automatic optimizations.

**Principle**: Zero-configuration optimization that works out of the box, with config-driven customization for different use cases.

---

## 1. Layered Architecture

### Go Core (`api/app/core/go_core/`)
- **Rule**: Contains high-performance, type-safe foundation with automatic optimizations
- **Rule**: Must be generic-based for compile-time type safety
- **Rule**: No application-specific logic - pure infrastructure concerns
- **Rule**: All optimizations must be automatic and zero-configuration by default
- **Rule**: Performance optimizations include: goroutine pools, context management, object pools, atomic operations, channel patterns

### Laravel Core (`api/app/core/laravel_core/`)
- **Rule**: Provides Laravel-style developer experience on top of Go Core
- **Rule**: Familiar APIs and patterns (facades, service providers, configuration)
- **Rule**: Automatic integration of optimizations from Go Core
- **Rule**: Config-driven customization through environment variables

---

## 2. Performance-First Design

### Automatic Optimizations
- **Rule**: All core services must have automatic goroutine optimization
- **Rule**: Context optimization must be baked into all operations
- **Rule**: Object pools for in-memory operations (JSON encoding/decoding)
- **Rule**: Atomic operations for counters and metrics
- **Rule**: Channel-based pipelines for data processing
- **Rule**: Work-stealing goroutine pools for optimal resource utilization

### Zero Configuration
- **Rule**: Developers should not need to think about optimizations
- **Rule**: Sensible defaults that work for most use cases
- **Rule**: Config-driven customization for specific needs
- **Rule**: Service providers automatically register optimized versions

---

## 3. Type Safety & Generics

### Generic Implementations
- **Rule**: Use generics for type-safe implementations
- **Rule**: Interfaces should be generic where possible
- **Rule**: Compile-time type checking over runtime reflection
- **Rule**: Generic constraints ensure correct data types
- **Rule**: Repository pattern must be generic: `Repository[T]`
- **Rule**: Cache must be generic: `Cache[T]`
- **Rule**: Events must be generic: `Event[T]`

### Interface Design
- **Rule**: Interfaces should be small and focused
- **Rule**: Composition over inheritance
- **Rule**: Context-aware interfaces where appropriate
- **Rule**: Performance interfaces for metrics and monitoring

---

## 4. Configuration Strategy

### Laravel-Style Config Files (`api/config/`)
- **Rule**: All configuration must be in `api/config/` following Laravel patterns
- **Rule**: Environment variable support with sensible defaults
- **Rule**: Profile-based configurations (web, API, background, streaming, batch)
- **Rule**: Operation-specific timeouts and limits
- **Rule**: Config files: `goroutine.go`, `context.go`, `go_channels.go`, etc.

### Environment Variables
- **Rule**: All settings must be configurable via environment variables
- **Rule**: Sensible defaults for all environments
- **Rule**: Profile-based environment variables for different use cases
- **Rule**: No hardcoded values in core systems

---

## 5. Service Provider Pattern

### Automatic Registration
- **Rule**: Service providers automatically register optimized versions of services
- **Rule**: Context-aware versions should be the default, not optional
- **Rule**: Service providers should handle all configuration loading
- **Rule**: No manual registration required from developers

### Core Service Providers
- **Rule**: `CoreServiceProvider` - registers all core services with optimizations
- **Rule**: `ContextServiceProvider` - provides context optimization utilities
- **Rule**: `GoroutineServiceProvider` - provides goroutine optimization
- **Rule**: `PerformanceServiceProvider` - provides performance monitoring

---

## 6. Separation of Concerns

### Core vs Application Logic
- **Rule**: Go Core contains only infrastructure concerns
- **Rule**: Laravel Core contains only framework patterns and developer experience
- **Rule**: No business logic in core systems
- **Rule**: No application-specific dependencies in core
- **Rule**: Core systems must be reusable across different applications

### Module Boundaries
- **Rule**: Clear boundaries between different core modules
- **Rule**: Interfaces define module contracts
- **Rule**: Dependency injection through service container
- **Rule**: No circular dependencies between modules

---

## 7. Developer Experience

### Laravel Familiarity
- **Rule**: APIs should look and feel like Laravel
- **Rule**: Facades for easy access to services
- **Rule**: Service providers for dependency injection
- **Rule**: Configuration management like Laravel
- **Rule**: Event/listener system like Laravel
- **Rule**: Queue system like Laravel

### Performance Transparency
- **Rule**: Optimizations should be invisible to developers
- **Rule**: No performance-related code in application logic
- **Rule**: Automatic profiling and metrics
- **Rule**: Configurable performance settings

---

## 8. Safety First

### Concurrency Safety
- **Rule**: Object pools only for in-memory operations (JSON encoding/decoding)
- **Rule**: Never reuse objects that interact with external systems (databases, APIs)
- **Rule**: Context optimization must respect cancellation and timeouts
- **Rule**: Goroutine pools must have proper cleanup and shutdown

### Resource Management
- **Rule**: Automatic cleanup of resources
- **Rule**: Timeout protection for all operations
- **Rule**: Memory-efficient object pools
- **Rule**: Proper error handling and propagation

---

## 9. Adding New Core Features

### Implementation Process
1. **Generalize in Go Core**: Create generic, type-safe implementation
2. **Apply Optimizations**: Integrate goroutine, context, and performance optimizations
3. **Laravel Integration**: Create Laravel-style facade and service provider
4. **Configuration**: Add config file and environment variables
5. **Documentation**: Update README and create usage examples

### Example: Adding Filesystem Feature
```go
// 1. Go Core - Generic filesystem interface
type Filesystem[T any] interface {
    Get(path string) (*T, error)
    Put(path string, data *T) error
    Delete(path string) error
    Exists(path string) (bool, error)
}

// 2. Optimized implementations
type S3Filesystem[T any] struct {
    // With goroutine optimization, context awareness, performance tracking
}

// 3. Laravel Core - Facade
facades.Storage().Get("file.txt")

// 4. Configuration
// api/config/filesystem.go
// FILESYSTEM_DEFAULT_DRIVER=s3
// FILESYSTEM_S3_TIMEOUT=30
```

---

## 10. Performance Expectations

### Benchmarks
- **Rule**: 10-50x faster than Laravel for typical operations
- **Rule**: 2-5x better than other Go frameworks
- **Rule**: 3-10x better than Node.js frameworks
- **Rule**: Real-world benchmarks required for validation

### Optimization Targets
- **Rule**: Repository operations: 30ms vs 300ms (Laravel)
- **Rule**: Event dispatching: 5ms vs 50ms (Laravel)
- **Rule**: Cache operations: 1ms vs 10ms (Laravel)
- **Rule**: Job processing: 100ms vs 1000ms (Laravel)

---

## 11. Code Quality Standards

### Structure
- **Rule**: Clear file organization in both cores
- **Rule**: Comprehensive documentation for all public APIs
- **Rule**: Example usage in README files
- **Rule**: Performance documentation for optimizations

### Testing
- **Rule**: Unit tests for all core functionality
- **Rule**: Performance benchmarks for optimizations
- **Rule**: Integration tests for Laravel-style features
- **Rule**: Safety tests for concurrency features

---

## 12. Configuration Profiles

### Use Case Optimization
- **Web Apps**: Low latency, fast response times (30s timeouts)
- **APIs**: Moderate timeouts, high throughput (60s timeouts)
- **Background Jobs**: Long timeouts, high performance (300s timeouts)
- **Streaming**: Very long timeouts, large buffers (1800s timeouts)
- **Batch Processing**: Long timeouts, large buffers (1800s timeouts)

### Environment Optimization
- **Development**: Fast timeouts for quick feedback
- **Staging**: Moderate timeouts for testing
- **Production**: Optimized for real workloads
- **Streaming**: Long timeouts for large operations

---

## 13. Implementation Checklist

When adding new features:

- [ ] **Go Core**: Generic, type-safe implementation
- [ ] **Optimizations**: Goroutine, context, performance optimizations
- [ ] **Safety**: Concurrency-safe, resource management
- [ ] **Laravel Core**: Facade and service provider integration
- [ ] **Configuration**: Config file and environment variables
- [ ] **Documentation**: README and usage examples
- [ ] **Testing**: Unit tests and performance benchmarks
- [ ] **Profiles**: Different configurations for different use cases

---

## 14. Key Principles Summary

1. **Performance First**: Automatic optimizations that work out of the box
2. **Laravel Familiarity**: Developer experience that feels like Laravel
3. **Type Safety**: Generic implementations with compile-time checking
4. **Config-Driven**: Environment-specific customization without code changes
5. **Safety First**: Concurrency-safe, resource-managed operations
6. **Zero Configuration**: Sensible defaults that work for most use cases
7. **Separation of Concerns**: Clear boundaries between core and application logic
8. **Automatic Integration**: Service providers handle all optimization setup

---

## Usage

Copy this rule set when making structural changes to ensure consistency with the architectural vision. This serves as the single source of truth for implementation decisions and helps maintain the performance-first, Laravel-inspired design philosophy. 