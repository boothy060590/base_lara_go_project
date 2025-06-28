# Service vs Repository: Proper Separation of Concerns

This document explains the proper separation between services and repositories, and how cross-cutting concerns fit into the architecture.

## The Problem: Overlap Between Services and Repositories

Initially, there was significant overlap between services and repositories:

```go
// ❌ BAD: Service just wrapping repository methods
type UserService struct {
    userRepo *repositories.UserRepository
}

func (s *UserService) FindByID(id uint) (interfaces.UserInterface, error) {
    return s.userRepo.FindByID(id) // Just a pass-through!
}

func (s *UserService) Create(data map[string]interface{}) (interfaces.UserInterface, error) {
    return s.userRepo.Create(data) // Just a pass-through!
}
```

This violates the **Single Responsibility Principle** and creates unnecessary layers.

## Proper Separation of Concerns

### Repository Layer (Data Access)
**Purpose**: Handle data persistence and retrieval

**Responsibilities**:
- CRUD operations (Create, Read, Update, Delete)
- Query building and execution
- Cache management
- Database-specific logic
- Data mapping between models
- Connection management

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

func (r *UserRepository) Create(data map[string]interface{}) (interfaces.UserInterface, error) {
    // Create in database
    dbUser := &db.User{}
    // ... map data to model
    err := r.db.Create(dbUser).Error
    if err != nil {
        return nil, err
    }
    
    // Convert and cache
    cacheUser := r.convertDBToCache(dbUser)
    r.storeInCache(cacheUser)
    
    return cacheUser, nil
}
```

### Service Layer (Business Logic)
**Purpose**: Handle business rules and orchestration

**Responsibilities**:
- Business validation
- Complex business operations
- Orchestrating multiple repositories
- Business rule enforcement
- Domain-specific operations
- Security checks
- Data transformation for business needs

**Example**:
```go
type UserService struct {
    userRepo *repositories.UserRepository
}

func (s *UserService) CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
    // ✅ Business validation
    if err := s.validateUserData(userData); err != nil {
        return nil, err
    }

    // ✅ Business rule: Check if user already exists
    if email, ok := userData["email"].(string); ok {
        existingUser, _ := s.userRepo.FindByEmail(email)
        if existingUser != nil {
            return nil, errors.New("user with this email already exists")
        }
    }

    // ✅ Business logic: Hash password
    if password, ok := userData["password"].(string); ok {
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if err != nil {
            return nil, err
        }
        userData["password"] = string(hashedPassword)
    }

    // ✅ Delegate to repository
    user, err := s.userRepo.Create(userData)
    if err != nil {
        return nil, err
    }

    // ✅ Business logic: Assign roles
    // This would involve a RoleService
    return user, nil
}

func (s *UserService) AuthenticateUser(email, password string) (interfaces.UserInterface, error) {
    // ✅ Get user from repository
    user, err := s.userRepo.FindByEmail(email)
    if err != nil {
        return nil, errors.New("invalid credentials")
    }

    // ✅ Business logic: Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.GetPassword()), []byte(password)); err != nil {
        return nil, errors.New("invalid credentials")
    }

    // ✅ Business rule: Check if user is active
    if !s.isUserActive(user) {
        return nil, errors.New("user account is inactive")
    }

    return user, nil
}
```

## Cross-Cutting Concerns Explained

**Cross-cutting concerns** are aspects of a program that affect multiple parts of the application and cannot be cleanly separated into a single module. They "cut across" the typical boundaries of object-oriented design.

### What Are Cross-Cutting Concerns?

1. **Logging** - Every operation might need to be logged
2. **Caching** - Multiple services might need caching
3. **Auditing** - Track changes across the system
4. **Security** - Authentication/authorization checks
5. **Performance Monitoring** - Track execution times
6. **Error Handling** - Consistent error processing
7. **Transaction Management** - Database transaction handling

### The Problem Without Decorators

Without decorators, you'd have to add logging/caching to every service method:

```go
// ❌ BAD: Logging scattered throughout services
func (s *UserService) CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
    log.Printf("Creating user with email: %s", userData["email"])
    start := time.Now()
    
    // Business logic...
    user, err := s.userRepo.Create(userData)
    
    duration := time.Since(start)
    log.Printf("User created in %v", duration)
    
    return user, err
}

func (s *UserService) AuthenticateUser(email, password string) (interfaces.UserInterface, error) {
    log.Printf("Authenticating user: %s", email)
    start := time.Now()
    
    // Business logic...
    user, err := s.userRepo.FindByEmail(email)
    
    duration := time.Since(start)
    log.Printf("Authentication completed in %v", duration)
    
    return user, err
}
```

### Solution: Decorators

Decorators allow you to add functionality without modifying existing code:

```go
// ✅ GOOD: Use decorators for cross-cutting concerns
userService, _ := services.NewUserService()

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
user, err := loggingDecorator.CreateUser(userData, roles) // Automatically logged
user, err := cachingDecorator.AuthenticateUser(email, password) // Automatically cached
```

## Updated Architecture

### Before (Overlapping)
```
Controller → Service → Repository → Model
     ↓           ↓           ↓         ↓
  Facades   CRUD Wrapper  CRUD      Cache/DB
```

### After (Proper Separation)
```
Controller → Service → Repository → Model
     ↓           ↓           ↓         ↓
  Facades   Business Logic  CRUD     Cache/DB
     ↓           ↓
Decorators  Cross-Cutting
```

## When to Use Each Layer

### Use Repository For:
- ✅ Simple CRUD operations
- ✅ Data queries and filtering
- ✅ Cache management
- ✅ Database-specific logic
- ✅ Data mapping

### Use Service For:
- ✅ Business validation
- ✅ Complex business operations
- ✅ Orchestrating multiple repositories
- ✅ Security checks
- ✅ Business rule enforcement
- ✅ Data transformation for business needs

### Use Decorators For:
- ✅ Logging operations
- ✅ Caching frequently accessed data
- ✅ Auditing changes
- ✅ Performance monitoring
- ✅ Error handling patterns

## Example: Complete Flow

```go
// 1. Controller receives request
func (c *AuthController) Register(ctx *gin.Context) {
    var request requests.RegisterRequest
    ctx.ShouldBindJSON(&request)
    
    userData := map[string]interface{}{
        "first_name": request.FirstName,
        "last_name":  request.LastName,
        "email":      request.Email,
        "password":   request.Password,
    }
    
    // 2. Use facade (with decorators applied)
    user, err := facades.CreateUser(userData, []string{"customer"})
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    ctx.JSON(201, gin.H{"user": user})
}

// 3. Facade calls service
func CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
    return User().Create(userData, roleNames) // Calls decorated service
}

// 4. Service handles business logic
func (s *UserService) CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
    // Business validation
    if err := s.validateUserData(userData); err != nil {
        return nil, err
    }
    
    // Business rule: Check uniqueness
    if existingUser, _ := s.userRepo.FindByEmail(userData["email"].(string)); existingUser != nil {
        return nil, errors.New("email already exists")
    }
    
    // Business logic: Hash password
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(userData["password"].(string)), bcrypt.DefaultCost)
    userData["password"] = string(hashedPassword)
    
    // 5. Delegate to repository
    return s.userRepo.Create(userData)
}

// 6. Repository handles data access
func (r *UserRepository) Create(data map[string]interface{}) (interfaces.UserInterface, error) {
    // Create in database
    dbUser := &db.User{}
    // ... map data
    r.db.Create(dbUser)
    
    // Cache the result
    cacheUser := r.convertDBToCache(dbUser)
    r.storeInCache(cacheUser)
    
    return cacheUser, nil
}
```

## Benefits of This Approach

1. **Clear Responsibilities** - Each layer has a specific purpose
2. **Testability** - Easy to mock repositories and test business logic
3. **Maintainability** - Changes to business logic don't affect data access
4. **Reusability** - Cross-cutting concerns can be reused across services
5. **Performance** - Caching and logging can be added without modifying business logic
6. **Scalability** - Easy to add new services and repositories following the same pattern

## Best Practices

1. **Services should contain business logic, not data access**
2. **Repositories should handle data persistence, not business rules**
3. **Use decorators for cross-cutting concerns, not inline code**
4. **Facades provide a clean API for controllers**
5. **Keep services focused on one domain**
6. **Test business logic in isolation from data access** 