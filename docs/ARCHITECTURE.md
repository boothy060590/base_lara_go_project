# Base Laravel Go Project - Architecture Documentation

## Table of Contents

1. [Software Architecture](#software-architecture)
2. [Service Layer Architecture](#service-layer-architecture)
3. [API Implementation](#api-implementation)
4. [Frontend Implementation](#frontend-implementation)
5. [Docker Infrastructure](#docker-infrastructure)
6. [Development Workflow](#development-workflow)

---

## Software Architecture

### Overview

This project implements a Laravel-inspired architecture in Go, featuring a service layer with proper separation of concerns, event-driven system with asynchronous processing, multi-queue management, and modern web development practices.

### Core Architecture Principles

#### 1. Separation of Concerns
- **API Layer**: HTTP controllers and middleware
- **Service Layer**: Business logic and orchestration
- **Repository Layer**: Data persistence and retrieval
- **Model Layer**: Data models and interfaces
- **Infrastructure Layer**: External services, queues, and mail

#### 2. Service Layer Architecture
- **Services**: Handle business logic and validation
- **Repositories**: Handle data access and caching
- **Facades**: Provide Laravel-style static access
- **Decorators**: Handle cross-cutting concerns

#### 3. Dependency Injection
- Service providers register dependencies
- Interfaces define contracts
- Facades provide simplified access
- Core package contains fundamental interfaces

#### 4. Event-Driven Design
- Events decouple business logic
- Listeners handle side effects
- Asynchronous processing via queues
- Laravel-style event dispatching

### Architecture Components

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (Vue.js)                    │
├─────────────────────────────────────────────────────────────┤
│                        API Gateway (Nginx)                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   API       │  │   Worker    │  │   MailHog   │         │
│  │  (Gin)      │  │  (Queue)    │  │  (SMTP)     │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   MySQL     │  │ ElasticMQ   │  │   Redis     │         │
│  │  (Database) │  │   (Queue)   │  │  (Cache)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

---

## Service Layer Architecture

### Overview

The service layer provides a clean separation between business logic and data access, following Laravel-style patterns with facades and decorators for cross-cutting concerns.

### Architecture Layers

```
Controllers → Services → Repositories → Models
     ↓           ↓           ↓           ↓
  Facades   Business Logic  CRUD      Cache/DB
     ↓           ↓
Decorators  Cross-Cutting
```

### Service Layer Components

#### 1. Service Interfaces
Base interfaces for common CRUD operations:

```go
type BaseServiceInterface[T any] interface {
    // Create operations
    Create(data map[string]interface{}) (T, error)
    CreateWithContext(ctx context.Context, data map[string]interface{}) (T, error)
    
    // Read operations
    FindByID(id uint) (T, error)
    FindByField(field string, value interface{}) (T, error)
    All() ([]T, error)
    Paginate(page, perPage int) ([]T, int64, error)
    
    // Update operations
    Update(id uint, data map[string]interface{}) (T, error)
    UpdateOrCreate(conditions map[string]interface{}, data map[string]interface{}) (T, error)
    
    // Delete operations
    Delete(id uint) error
    DeleteWhere(conditions map[string]interface{}) error
    
    // Utility operations
    Exists(id uint) (bool, error)
    Count() (int64, error)
    CountWhere(conditions map[string]interface{}) (int64, error)
}
```

#### 2. Service Facades
Laravel-style static access to services:

```go
// Laravel-style facade usage
user, err := facades.CreateUser(userData, roles)
user, err := facades.AuthenticateUser(email, password)

// Facade implementation
type UserServiceFacade struct{}

func (u *UserServiceFacade) Create(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
    if userService, ok := globalUserService.(interface {
        CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error)
    }); ok {
        return userService.CreateUser(userData, roleNames)
    }
    return nil, errors.New("user service not found")
}
```

#### 3. Service Decorators
Cross-cutting concerns without modifying core services:

```go
// Add logging decorator
logger := log.New(log.Writer(), "[USER_SERVICE] ", log.LstdFlags)
loggingDecorator := core.NewLoggingDecorator[interfaces.UserInterface](userService, logger)

// Add caching decorator
cachingDecorator := core.NewCachingDecorator[interfaces.UserInterface](
    userService, 
    facades.CacheInstance, 
    30*time.Minute,
)

// Use decorated service
user, err := loggingDecorator.CreateUser(data)
user, err := cachingDecorator.AuthenticateUser(email, password)
```

### Service vs Repository Separation

#### Repository Layer (Data Access)
**Purpose**: Handle data persistence and retrieval

**Responsibilities**:
- CRUD operations (Create, Read, Update, Delete)
- Query building and execution
- Cache management
- Database-specific logic
- Data mapping between models

**Example**:
```go
type UserRepository struct {
    db    *gorm.DB
    cache core.CacheInterface
}

func (r *UserRepository) FindByID(id uint) (interfaces.UserInterface, error) {
    // Try cache first
    if cached, exists := r.cache.Get(cacheKey); exists {
        return cached, nil
    }
    
    // Get from database
    dbUser := &db.User{}
    err := r.db.Preload("Roles.Permissions").First(dbUser, id).Error
    if err != nil {
        return nil, err
    }
    
    // Convert and cache
    cacheUser := r.convertDBToCache(dbUser)
    r.storeInCache(cacheUser)
    
    return cacheUser, nil
}
```

#### Service Layer (Business Logic)
**Purpose**: Handle business rules and orchestration

**Responsibilities**:
- Business validation
- Complex business operations
- Orchestrating multiple repositories
- Business rule enforcement
- Domain-specific operations
- Security checks

**Example**:
```go
type UserService struct {
    userRepo *repositories.UserRepository
}

func (s *UserService) CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
    // Business validation
    if err := s.validateUserData(userData); err != nil {
        return nil, err
    }

    // Business rule: Check if user already exists
    if email, ok := userData["email"].(string); ok {
        existingUser, _ := s.userRepo.FindByEmail(email)
        if existingUser != nil {
            return nil, errors.New("user with this email already exists")
        }
    }

    // Business logic: Hash password
    if password, ok := userData["password"].(string); ok {
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }
        userData["password"] = string(hashedPassword)
    }

    // Delegate to repository
    return s.userRepo.Create(userData)
}
```

### Cross-Cutting Concerns

Cross-cutting concerns are aspects that affect multiple parts of your application:

- **Logging** - Every operation needs logging
- **Caching** - Multiple services need caching
- **Auditing** - Track changes across the system
- **Security** - Authentication/authorization
- **Performance Monitoring** - Track execution times

**Without decorators**, you'd have to add this code to every service method:
```go
// ❌ BAD: Scattered throughout codebase
func (s *UserService) CreateUser(data map[string]interface{}) (interfaces.UserInterface, error) {
    log.Printf("Creating user...") // Logging
    start := time.Now()           // Performance monitoring
    
    // Business logic...
    
    duration := time.Since(start)
    log.Printf("User created in %v", duration)
    return user, nil
}
```

**With decorators**, you add it once and reuse:
```go
// ✅ GOOD: Add once, use everywhere
loggingDecorator := core.NewLoggingDecorator(userService, logger)
cachingDecorator := core.NewCachingDecorator(userService, cache, 30*time.Minute)

// Automatically logged and cached
user, err := loggingDecorator.CreateUser(data)
user, err := cachingDecorator.AuthenticateUser(email, password)
```

---

## API Implementation

### Technology Stack

- **Language**: Go 1.23.0
- **Framework**: Gin (HTTP router)
- **ORM**: GORM with MySQL driver
- **Authentication**: JWT tokens
- **Validation**: go-playground/validator
- **Queue**: AWS SQS (ElasticMQ for development)
- **Mail**: SMTP (MailHog for development)

### Project Structure

```
api/
├── app/
│   ├── core/                    # Core interfaces and base types
│   │   ├── service_interfaces.go    # Base service interfaces
│   │   ├── service_decorators.go    # Cross-cutting concerns
│   │   ├── base_service.go          # Base service implementation
│   │   ├── base_dto.go         # Base DTO interface
│   │   ├── base_model.go       # Base model with common fields
│   │   ├── database.go         # Database connection and configuration
│   │   ├── event_dispatcher.go # Event dispatching system
│   │   ├── event_registry.go   # Event factory registry
│   │   ├── interfaces.go       # Core service interfaces
│   │   ├── queue_worker.go     # Queue worker implementation
│   │   ├── register.go         # Service registration
│   │   └── registry.go         # Service registry
│   ├── services/               # Business logic services
│   │   └── user_service.go     # User business logic
│   ├── repositories/           # Data access layer
│   │   └── user_repository.go  # User data access
│   ├── facades/                # Service facades
│   │   ├── service.go          # Service facades
│   │   ├── database.go         # Database facade
│   │   ├── event.go            # Event facade
│   │   ├── job.go              # Job facade
│   │   └── mail.go             # Mail facade
│   ├── data_objects/           # Data Transfer Objects
│   │   └── auth/
│   │       └── user_dto.go     # User DTO implementation
│   ├── events/                 # Event definitions
│   │   └── auth/
│   │       └── user_created.go # User creation event
│   ├── http/                   # HTTP layer
│   │   ├── controllers/        # HTTP controllers
│   │   ├── middlewares/        # HTTP middlewares
│   │   └── requests/           # Request validation
│   ├── jobs/                   # Background jobs
│   │   └── auth/
│   │       ├── create_user.go
│   │       ├── get_logged_in_user.go
│   │       └── login_user.go
│   ├── listeners/              # Event listeners
│   │   ├── base_listener.go
│   │   └── send_email_confirmation.go
│   ├── models/                 # Database models
│   │   ├── interfaces/         # Model interfaces
│   │   ├── db/                 # Database models
│   │   │   ├── category.go
│   │   │   ├── permission.go
│   │   │   ├── role.go
│   │   │   ├── service.go
│   │   │   └── user.go
│   │   └── cache/              # Cache models
│   │       └── user.go
│   ├── providers/              # Service providers
│   │   ├── service_provider.go # Service registration
│   │   ├── database_service_provider.go
│   │   ├── event_dispatcher_provider.go
│   │   ├── form_field_validators_provider.go
│   │   ├── job_dispatcher_provider.go
│   │   ├── listener_service_provider.go
│   │   ├── mail_service_provider.go
│   │   ├── message_processor_provider.go
│   │   ├── queue_service_provider.go
│   │   └── router_service_provider.go
│   ├── transformers/           # Data transformers
│   │   └── user_transformer.go
│   ├── utils/                  # Utility functions
│   │   └── token/
│   │       └── token.go
│   └── validators/             # Custom validators
│       └── name_field_validator.go
├── bootstrap/                  # Application bootstrap
│   ├── api/
│   │   └── main.go            # API entry point
│   └── worker/
│       └── main.go            # Worker entry point
├── config/                     # Configuration files
├── database/                   # Database migrations
│   └── migrations/
└── routes/                     # Route definitions
    └── api/
        └── v1/
            └── auth/
                └── auth.go
```

### Key Design Patterns

#### 1. Service Provider Pattern
Service providers register dependencies and configure services:

```go
// Example: Service Provider
func RegisterServices() {
    // Create base user service
    userService, err := services.NewUserService()
    if err == nil {
        // Register the base service
        GlobalServiceContainer.Register("user", userService)
        
        // Set up the service facade
        facades.SetUserService(userService)
        
        log.Println("User service registered successfully")
    }
}
```

#### 2. Facade Pattern
Facades provide simplified access to complex services:

```go
// Example: Service Facade
func CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
    return User().Create(userData, roleNames)
}
```

#### 3. Decorator Pattern
Decorators add functionality without modifying existing code:

```go
// Example: Logging Decorator
type LoggingDecorator[T any] struct {
    *ServiceDecorator[T]
    logger *log.Logger
}

func (l *LoggingDecorator[T]) Create(data map[string]interface{}) (T, error) {
    start := time.Now()
    l.logger.Printf("Creating %T with data: %v", *new(T), data)
    
    result, err := l.service.Create(data)
    
    duration := time.Since(start)
    if err != nil {
        l.logger.Printf("Failed to create %T after %v: %v", *new(T), duration, err)
    } else {
        l.logger.Printf("Successfully created %T after %v", *new(T), duration)
    }
    
    return result, err
}
```

#### 4. Event-Driven Architecture
Events decouple business logic from side effects:

```go
// Dispatch event
providers.DispatchEvent("UserCreated", userData)

// Listen for event
type SendEmailConfirmationListener struct{}

func (l *SendEmailConfirmationListener) Handle(event interface{}) error {
    // Send welcome email
    return nil
}
```

#### 5. Repository Pattern
Models implement interfaces for data access:

```go
type UserInterface interface {
    Create(user *User) error
    FindByID(id uint) (*User, error)
    FindByEmail(email string) (*User, error)
}
```

### Service Layer Benefits

1. **Clear Responsibilities** - Each layer has a specific purpose
2. **Testability** - Easy to mock repositories and test business logic
3. **Maintainability** - Changes to business logic don't affect data access
4. **Reusability** - Cross-cutting concerns can be reused across services
5. **Performance** - Caching and logging can be added without modifying business logic
6. **Scalability** - Easy to add new services and repositories following the same pattern

### Best Practices

1. **Services should contain business logic, not data access**
2. **Repositories should handle data persistence, not business rules**
3. **Use decorators for cross-cutting concerns, not inline code**
4. **Facades provide a clean API for controllers**
5. **Keep services focused on one domain**
6. **Test business logic in isolation from data access**

---

## Frontend Implementation

### Technology Stack

- **Framework**: Vue.js 3 with Composition API
- **Build Tool**: Vite
- **Styling**: SCSS with modern CSS features
- **HTTP Client**: Axios
- **Form Validation**: Custom validators
- **State Management**: Vuex (if needed)

### Project Structure

```
frontend/
├── src/
│   ├── components/            # Reusable Vue components
│   │   └── form/             # Form components
│   │       ├── EmailFormField.vue
│   │       ├── PasswordFormField.vue
│   │       ├── TelephoneFormField.vue
│   │       └── TextFormField.vue
│   ├── Pages/                # Page components
│   │   ├── auth/             # Authentication pages
│   │   │   ├── login/
│   │   │   │   ├── Login.vue
│   │   │   │   └── login.scss
│   │   │   └── register/
│   │   │       ├── Register.vue
│   │   │       └── Register.scss
│   │   └── home/             # Home pages
│   │       ├── admin/
│   │       │   └── Admin.vue
│   │       ├── customer/
│   │       │   └── Customer.vue
│   │       └── Home.vue
│   ├── helpers/              # Helper functions
│   │   └── api/              # API helpers
│   │       ├── api.js        # Base API configuration
│   │       └── auth/         # Auth-specific API calls
│   │           └── authApi.js
│   ├── form_validators/      # Form validation
│   │   ├── index.js          # Validator exports
│   │   ├── login_validator.js
│   │   └── register_validator.js
│   ├── store/                # State management
│   │   └── auth.js           # Authentication state
│   ├── App.vue               # Root component
│   ├── main.js               # Application entry point
│   └── router.js             # Vue Router configuration
├── public/                   # Static assets
├── index.html                # HTML template
├── package.json              # Dependencies
├── vite.config.js            # Vite configuration
└── README.md                 # Frontend documentation
```

### Key Features

#### 1. Form Validation
Custom form validation with real-time feedback:

```javascript
// login_validator.js
export const loginValidator = {
    email: {
        required: true,
        email: true,
        message: 'Please enter a valid email address'
    },
    password: {
        required: true,
        minLength: 8,
        message: 'Password must be at least 8 characters long'
    }
}
```

#### 2. API Integration
Centralized API configuration with authentication:

```javascript
// api.js
import axios from 'axios'

const api = axios.create({
    baseURL: 'https://api.baselaragoproject.test',
    timeout: 10000,
    headers: {
        'Content-Type': 'application/json'
    }
})

// Request interceptor for authentication
api.interceptors.request.use(config => {
    const token = localStorage.getItem('auth_token')
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

export default api
```

#### 3. Component Architecture
Reusable form components with validation:

```vue
<!-- EmailFormField.vue -->
<template>
    <div class="form-field">
        <label :for="id">{{ label }}</label>
        <input
            :id="id"
            type="email"
            :value="modelValue"
            @input="$emit('update:modelValue', $event.target.value)"
            :class="{ 'error': hasError }"
        />
        <span v-if="error" class="error-message">{{ error }}</span>
    </div>
</template>

<script>
export default {
    props: {
        modelValue: String,
        label: String,
        error: String,
        id: String
    },
    emits: ['update:modelValue'],
    computed: {
        hasError() {
            return !!this.error
        }
    }
}
</script>
```

---

## Docker Infrastructure

### Container Architecture

The application uses Docker Compose for development with the following services:

```
┌─────────────────────────────────────────────────────────────┐
│                    Docker Compose                           │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   Frontend  │  │     API     │  │   Worker    │         │
│  │   (Vue.js)  │  │    (Gin)    │  │   (Queue)   │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │    Nginx    │  │   MySQL     │  │  ElasticMQ  │         │
│  │ (Reverse    │  │ (Database)  │  │   (Queue)   │         │
│  │   Proxy)    │  └─────────────┘  └─────────────┘         │
│  └─────────────┘                                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │   MailHog   │  │     DNS     │  │     SSL     │         │
│  │   (SMTP)    │  │  (dnsmasq)  │  │  (Certs)    │         │
│  └─────────────┘  └─────────────┘  └─────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### Service Configuration

#### 1. API Service
```yaml
api:
  build:
    context: ./api
    dockerfile: docker/api/Dockerfile
  volumes:
    - ./api:/app
    - /app/tmp
  environment:
    - APP_ENV=development
    - DB_HOST=db
    - SQS_ENDPOINT=http://elasticmq:9324
  depends_on:
    - db
    - elasticmq
```

#### 2. Worker Service
```yaml
worker:
  build:
    context: ./api
    dockerfile: docker/worker/Dockerfile
  volumes:
    - ./api:/app
  environment:
    - APP_ENV=development
    - DB_HOST=db
    - SQS_ENDPOINT=http://elasticmq:9324
  depends_on:
    - db
    - elasticmq
```

#### 3. Frontend Service
```yaml
frontend:
  build:
    context: ./frontend
    dockerfile: docker/frontend/Dockerfile
  volumes:
    - ./frontend:/app
    - /app/node_modules
  environment:
    - VITE_API_URL=https://api.baselaragoproject.test
```

#### 4. Database Service
```yaml
db:
  image: mysql:8.0
  environment:
    MYSQL_ROOT_PASSWORD: root_password
    MYSQL_DATABASE: dev_base_lara_go
    MYSQL_USER: api_user
    MYSQL_PASSWORD: b4s3L4r4G0212!
  volumes:
    - mysql_data:/var/lib/mysql
    - ./docker/mysql/init.sql:/docker-entrypoint-initdb.d/init.sql
```

#### 5. Queue Service
```yaml
elasticmq:
  image: softwaremill/elasticmq-native
  ports:
    - "9324:9324"
    - "9325:9325"
  volumes:
    - ./docker/elasticmq/elasticmq.conf:/opt/elasticmq.conf
```

### Development Workflow

#### 1. Local Development
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f api

# Access services
# Frontend: https://app.baselaragoproject.test
# API: https://api.baselaragoproject.test
# Mail: http://mail.baselaragoproject.test:8025
```

#### 2. Hot Reloading
- **Frontend**: Vite provides instant hot reloading
- **Backend**: Air watches for Go file changes and restarts
- **Worker**: Automatic restart on code changes

#### 3. Database Management
```bash
# Run migrations
docker-compose exec api go run main.go migrate

# Access database
docker-compose exec db mysql -u api_user -p dev_base_lara_go
```

#### 4. Queue Management
```bash
# View queue status
curl http://localhost:9325/queue/jobs

# Send test message
aws --endpoint-url http://localhost:9324 sqs send-message \
  --queue-url http://localhost:9324/queue/jobs \
  --message-body '{"test": "message"}'
```

---

## Development Workflow

### Code Organization

#### 1. Service Development
```bash
# Create new service
touch api/app/services/new_service.go

# Create new repository
touch api/app/repositories/new_repository.go

# Create new facade methods
# Edit api/app/facades/service.go

# Register in service provider
# Edit api/app/providers/service_provider.go
```

#### 2. Event Development
```bash
# Create new event
touch api/app/events/new_event.go

# Create new listener
touch api/app/listeners/new_listener.go

# Register in event provider
# Edit api/app/providers/event_service_provider.go
```

#### 3. Job Development
```bash
# Create new job
touch api/app/jobs/new_job.go

# Create new processor
# Edit api/app/providers/job_processor_provider.go
```

### Testing Strategy

#### 1. Unit Testing
```bash
# Test services
go test ./api/app/services

# Test repositories
go test ./api/app/repositories

# Test with coverage
go test -cover ./api/app/services
```

#### 2. Integration Testing
```bash
# Test API endpoints
curl -X POST https://api.baselaragoproject.test/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "password123"}'

# Test queue processing
# Check worker logs for job processing
docker-compose logs -f worker
```

#### 3. Frontend Testing
```bash
# Run frontend tests
cd frontend
npm run test

# Run with coverage
npm run test:coverage
```

### Deployment Strategy

#### 1. Development Environment
- Docker Compose for local development
- Hot reloading for rapid iteration
- Local database and queue services

#### 2. Staging Environment
- Docker containers on staging server
- Production-like configuration
- Automated testing and validation

#### 3. Production Environment
- Container orchestration (Kubernetes/Docker Swarm)
- Load balancing and auto-scaling
- Monitoring and logging
- Database clustering and backup

### Performance Optimization

#### 1. Backend Optimization
- Connection pooling for database
- Caching with Redis
- Queue batching and optimization
- Service decorators for cross-cutting concerns

#### 2. Frontend Optimization
- Code splitting and lazy loading
- Asset optimization and compression
- CDN for static assets
- Service worker for caching

#### 3. Infrastructure Optimization
- Load balancing across multiple instances
- Database read replicas
- Queue partitioning and scaling
- CDN for global content delivery

This architecture provides a solid foundation for building scalable, maintainable applications with Laravel-style patterns in Go, while maintaining excellent performance and developer productivity.