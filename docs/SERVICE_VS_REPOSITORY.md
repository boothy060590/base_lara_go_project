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

## Caching Architecture in Service vs Repository Separation

### Caching Responsibilities

#### Repository Layer (Cache Management)
**Purpose**: Handle cache operations and data persistence

**Cache Responsibilities**:
- Cache hit/miss logic
- Cache storage and retrieval
- Cache key generation
- Cache invalidation
- Data serialization/deserialization
- Cache model conversion

**Example**:
```go
type UserRepository struct {
    db    *gorm.DB
    cache core.CacheInterface
}

func (r *UserRepository) FindByID(id uint) (interfaces.UserInterface, error) {
    // ✅ Cache hit/miss logic
    user := &cache.User{}
    found, err := core.GetCachedModelByID("users", id, user)
    if err == nil && found {
        return user, nil // Cache hit - 10x faster
    }

    // ✅ Database query with relationships
    dbUser := &db.User{}
    err = r.db.Preload("Roles.Permissions").First(dbUser, id).Error
    if err != nil {
        return nil, err
    }

    // ✅ Cache storage
    cacheUser := r.convertDBToCache(dbUser)
    r.storeInCache(cacheUser)

    return cacheUser, nil
}

func (r *UserRepository) storeInCache(user *cache.User) {
    // ✅ Automatic serialization and storage
    err := core.CacheModel(user)
    if err != nil {
        return
    }

    // ✅ Email index for fast lookups
    emailCacheKey := fmt.Sprintf("users:email:%s", user.Email)
    r.cache.Set(emailCacheKey, user.GetID(), time.Hour)
}

func (r *UserRepository) removeFromCache(id uint) {
    // ✅ Cache invalidation
    user := &cache.User{}
    user.Set("id", id)
    core.ForgetModel(user)
}
```

#### Service Layer (Cache Strategy)
**Purpose**: Define caching policies and business rules

**Cache Responsibilities**:
- Cache warming strategies
- Cache invalidation policies
- Business-driven cache decisions
- Cache performance monitoring
- Cache consistency rules

**Example**:
```go
type UserService struct {
    userRepo *repositories.UserRepository
}

func (s *UserService) GetUserWithRoles(id uint) (interfaces.UserInterface, error) {
    // ✅ Business rule: Always load roles for authentication
    // No permission check needed - roles are essential for auth
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // ✅ Cache warming: Prefetch related data
    go s.warmUserRelatedCache(user)

    return user, nil
}

func (s *UserService) UpdateUser(id uint, data map[string]interface{}) (interfaces.UserInterface, error) {
    // ✅ Business validation
    if err := s.validateProfileUpdate(data); err != nil {
        return nil, err
    }

    // ✅ Update in repository (handles cache invalidation)
    user, err := s.userRepo.Update(id, data)
    if err != nil {
        return nil, err
    }

    // ✅ Business-driven cache warming
    if email, ok := data["email"].(string); ok {
        // Email changed - warm cache for new email
        go s.warmEmailIndex(user.GetID(), email)
    }

    return user, nil
}

func (s *UserService) warmUserRelatedCache(user interfaces.UserInterface) {
    // ✅ Cache warming strategy
    // Prefetch user's roles and permissions for faster subsequent access
    for _, role := range user.GetRoles() {
        // Warm role cache
        core.CacheModel(&cache.Role{
            Name: role.GetName(),
            // ... other role data
        })
    }
}
```

### Laravel-Style Caching Patterns

#### 1. Automatic Cache Management
```go
// ✅ Repository handles all cache operations automatically
func (r *UserRepository) FindByID(id uint) (interfaces.UserInterface, error) {
    // Automatic cache key generation
    user := &cache.User{}
    found, err := core.GetCachedModelByID("users", id, user)
    if err == nil && found {
        return user, nil
    }

    // Automatic cache storage after database query
    dbUser := &db.User{}
    err = r.db.Preload("Roles.Permissions").First(dbUser, id).Error
    if err != nil {
        return nil, err
    }

    cacheUser := r.convertDBToCache(dbUser)
    r.storeInCache(cacheUser) // Automatic serialization

    return cacheUser, nil
}
```

#### 2. Clean Field Mapping
```go
// ✅ Laravel-style field mapping in cache models
func (u *User) FromCacheData(data map[string]interface{}) error {
    u.Initialize()
    u.Fill(data)
    u.populateStructFields(data) // Clean, maintainable
    return nil
}

func (u *User) populateStructFields(data map[string]interface{}) {
    // ✅ No nested if statements - Laravel-style
    fieldMappings := map[string]func(interface{}) {
        "first_name": func(value interface{}) {
            if str, ok := value.(string); ok {
                u.FirstName = str
            }
        },
        "email": func(value interface{}) {
            if str, ok := value.(string); ok {
                u.Email = str
            }
        },
        // ... more fields
    }
    u.FillFields(data, fieldMappings)
}
```

#### 3. Cache Key Strategy
```go
// ✅ Automatic key generation using base keys
func (u *User) GetBaseKey() string {
    return "users" // Model type identifier
}

func (u *User) GetCacheKey() string {
    return fmt.Sprintf("%s:%d:data", u.GetBaseKey(), u.GetID())
}

// ✅ Index keys for fast lookups
emailCacheKey := fmt.Sprintf("users:email:%s", user.Email)
r.cache.Set(emailCacheKey, user.GetID(), time.Hour)
```

### Cache Performance Benefits

#### 1. Repository-Level Caching
```go
// ✅ Cache hit: ~0.1ms (Redis GET)
// ✅ Cache miss: ~1-5ms (Database + cache storage)
// ✅ Overall: 10-50x speedup for cached data

func (r *UserRepository) FindByID(id uint) (interfaces.UserInterface, error) {
    // O(1) cache operation
    found, err := core.GetCachedModelByID("users", id, user)
    if err == nil && found {
        return user, nil // 10x faster than database
    }

    // O(n) database operation with joins
    err = r.db.Preload("Roles.Permissions").First(dbUser, id).Error
    // ... cache storage
}
```

#### 2. Service-Level Cache Strategy
```go
// ✅ Business-driven cache decisions
func (s *UserService) shouldCacheUser(user interfaces.UserInterface) bool {
    // Cache only active users
    return user.GetLastLoginAt().After(time.Now().Add(-30 * 24 * time.Hour))
}

// ✅ Cache warming for frequently accessed data
func (s *UserService) WarmFrequentlyAccessedUsers() {
    users, _ := s.userRepo.All()
    for _, user := range users {
        if s.shouldCacheUser(user) {
            core.CacheModel(user)
        }
    }
}
```

### Cache Consistency Patterns

#### 1. Repository-Level Invalidation
```go
// ✅ Automatic cache invalidation on data changes
func (r *UserRepository) Update(id uint, data map[string]interface{}) (interfaces.UserInterface, error) {
    // Update database
    dbUser := &db.User{}
    err := r.db.First(dbUser, id).Error
    if err != nil {
        return nil, err
    }

    // Update fields
    // ... field updates

    err = r.db.Save(dbUser).Error
    if err != nil {
        return nil, err
    }

    // Reload with relationships
    err = r.db.Preload("Roles.Permissions").First(dbUser, id).Error
    if err != nil {
        return nil, err
    }

    // ✅ Automatic cache update
    cacheUser := r.convertDBToCache(dbUser)
    r.storeInCache(cacheUser)

    return cacheUser, nil
}
```

#### 2. Service-Level Cache Policies
```go
// ✅ Business-driven cache invalidation
func (s *UserService) DeactivateUser(id uint) error {
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return err
    }

    // Business rule: Cannot deactivate admin users
    if s.isAdminUser(user) {
        return errors.New("cannot deactivate admin users")
    }

    // Update user status
    _, err = s.userRepo.Update(id, map[string]interface{}{
        "status": "inactive",
    })

    // ✅ Cache automatically invalidated by repository
    return err
}
```

### Best Practices Summary

#### Repository Layer
- ✅ Handle all cache operations (get, set, delete)
- ✅ Automatic serialization/deserialization
- ✅ Cache key generation and management
- ✅ Database query optimization with relationships
- ✅ Cache model conversion

#### Service Layer
- ✅ Define cache warming strategies
- ✅ Business-driven cache decisions
- ✅ Cache performance monitoring
- ✅ Cache consistency policies
- ✅ Orchestrate cache operations

#### Cross-Cutting Concerns
- ✅ Use decorators for logging and monitoring
- ✅ Automatic cache management
- ✅ Clean, maintainable code patterns
- ✅ Laravel-style field mapping
- ✅ Type-safe cache operations 