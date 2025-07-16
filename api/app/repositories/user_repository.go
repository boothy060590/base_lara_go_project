package repositories

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/models"
	"context"
	"fmt"
	"time"
)

// UserRepository provides intelligent optimization for user data access
type UserRepository struct {
	repository app_core.Repository[models.User]
	cache      app_core.Cache[models.User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(repository app_core.Repository[models.User], cache app_core.Cache[models.User]) *UserRepository {
	return &UserRepository{
		repository: repository,
		cache:      cache,
	}
}

// Fast Path Operations (Tier 1 - minimal overhead)
// Perfect for simple lookups that happen frequently

// Find retrieves a user by ID - uses fast path (no optimization overhead)
func (r *UserRepository) Find(id uint) (*models.User, error) {
	// Fast path: direct SQL with minimal overhead
	return r.repository.Find(id)
}

// FindByEmail retrieves a user by email - uses fast path
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	// Fast path: single field lookup with minimal overhead
	return r.repository.FindBy("email", email)
}

// Exists checks if a user exists - uses fast path
func (r *UserRepository) Exists(id uint) (bool, error) {
	// Fast path: optimized EXISTS query
	return r.repository.Exists(id)
}

// Count returns total user count - uses fast path
func (r *UserRepository) Count() (int64, error) {
	// Fast path: simple COUNT query
	return r.repository.Count()
}

// Create creates a new user - uses fast path
func (r *UserRepository) Create(user *models.User) error {
	// Fast path: direct INSERT with minimal overhead
	return r.repository.Create(user)
}

// Update updates a user - uses fast path
func (r *UserRepository) Update(user *models.User) error {
	// Fast path: direct UPDATE with minimal overhead
	return r.repository.Update(user)
}

// Delete deletes a user - uses fast path
func (r *UserRepository) Delete(id uint) error {
	// Fast path: direct soft delete
	return r.repository.Delete(id)
}

// Balanced Path Operations (Tier 2 - moderate optimization)
// Good for queries with 1-3 conditions, basic pagination

// FindByRole retrieves users by role - uses balanced path
func (r *UserRepository) FindByRole(roleName string) ([]models.User, error) {
	// Balanced path: statement caching + connection pooling
	return r.repository.Where(map[string]any{"role": roleName}).Get()
}

// FindActive retrieves active users - uses balanced path
func (r *UserRepository) FindActive() ([]models.User, error) {
	// Balanced path: simple WHERE condition
	return r.repository.Where(map[string]any{"status": "active"}).Get()
}

// FindAll retrieves all users with pagination - uses balanced path
func (r *UserRepository) FindAll(page, perPage int) ([]models.User, int64, error) {
	// Balanced path: pagination with moderate optimization
	return r.repository.Where(map[string]any{}).Paginate(page, perPage)
}

// FindByStatus retrieves users by status with optional role filter
func (r *UserRepository) FindByStatus(status string, role ...string) ([]models.User, error) {
	conditions := map[string]any{"status": status}
	
	// If role is provided, add it to conditions
	if len(role) > 0 && role[0] != "" {
		conditions["role"] = role[0]
	}
	
	// Auto-escalation: if conditions become complex, this will automatically
	// escalate to complex path when needed
	return r.repository.Where(conditions).Get()
}

// Complex Path Operations (Tier 3 - full optimization)
// For complex queries, reporting, bulk operations

// GetUserReports generates complex user reports - uses complex path
func (r *UserRepository) GetUserReports(ctx context.Context) ([]models.User, error) {
	// Complex path: full optimization suite enabled
	return r.repository.Complex().
		LeftJoin("user_profiles", "users.id = user_profiles.user_id").
		LeftJoin("user_roles", "users.id = user_roles.user_id").
		GroupBy("users.id").
		WithBatching(true).
		WithAsync(false). // Synchronous for reporting
		Build().
		WithContext(ctx).
		WithMetrics(true).
		Get()
}

// BulkCreateUsers creates multiple users efficiently - uses complex path
func (r *UserRepository) BulkCreateUsers(users []*models.User) error {
	// Complex path: bulk operations with full optimization
	return r.repository.Complex().
		WithBatching(true).
		WithAsync(true).
		BulkCreate(users)
}

// BulkUpdateUsers updates multiple users efficiently - uses complex path
func (r *UserRepository) BulkUpdateUsers(users []*models.User) error {
	// Complex path: bulk operations with full optimization
	return r.repository.Complex().
		WithBatching(true).
		WithAsync(true).
		BulkUpdate(users)
}

// GetUserAnalytics performs complex analytics queries - uses complex path
func (r *UserRepository) GetUserAnalytics(ctx context.Context) ([]models.User, error) {
	// Complex path: raw SQL with full optimization
	return r.repository.Complex().
		Raw(`
			SELECT u.*, 
				   COUNT(ul.id) as login_count,
				   AVG(ua.session_duration) as avg_session_duration
			FROM users u
			LEFT JOIN user_logins ul ON u.id = ul.user_id
			LEFT JOIN user_activities ua ON u.id = ua.user_id
			WHERE u.created_at >= DATE_SUB(NOW(), INTERVAL 30 DAY)
			GROUP BY u.id
			HAVING login_count > 0
			ORDER BY login_count DESC
		`).
		WithContext(ctx).
		WithMetrics(true).
		Get()
}

// StreamActiveUsers streams active users for large datasets - uses complex path
func (r *UserRepository) StreamActiveUsers(ctx context.Context) (<-chan models.User, error) {
	// Complex path: streaming with work stealing for large datasets
	return r.repository.Complex().
		WithWorkStealing(true).
		WithPipeline(true).
		Build().
		WithContext(ctx).
		Stream()
}

// Manual Optimization Control
// For cases where you want to explicitly control optimization level

// FindWithOptimization allows manual control of optimization level
func (r *UserRepository) FindWithOptimization(id uint, level app_core.QueryComplexity) (*models.User, error) {
	// Manual optimization control
	return r.repository.WithOptimization(level).Find(id)
}

// SearchUsers with automatic complexity detection
func (r *UserRepository) SearchUsers(filters map[string]any) ([]models.User, error) {
	// The smart repository will automatically choose the right optimization level
	// based on the number and complexity of conditions
	return r.repository.Where(filters).Get()
}

// Transaction Support with Smart Optimization
func (r *UserRepository) CreateUserWithProfile(ctx context.Context, user *models.User, profile map[string]any) error {
	// Transaction with smart optimization
	return r.repository.WithContext(ctx).Transaction(func(repo app_core.Repository[models.User]) error {
		// Create user (fast path)
		if err := repo.Create(user); err != nil {
			return err
		}
		
		// Create profile (would need a separate repository for UserProfile)
		// This is just an example of how transactions work
		
		return nil
	})
}

// Performance Monitoring
func (r *UserRepository) GetPerformanceStats() map[string]any {
	// Get performance statistics from complex path
	return r.repository.Complex().Build().GetStats()
}

// Cache Integration Examples
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

// Context-Aware Operations
func (r *UserRepository) FindWithContext(ctx context.Context, id uint) (*models.User, error) {
	// Context-aware fast path operation
	return r.repository.WithContext(ctx).Find(id)
}

func (r *UserRepository) SearchWithContext(ctx context.Context, filters map[string]any) ([]models.User, error) {
	// Context-aware smart query with automatic optimization
	return r.repository.WithContext(ctx).Where(filters).Get()
}

// Utility methods for demonstration
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	user, err := r.FindByEmail(email)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

func (r *UserRepository) CountByStatus(status string) (int64, error) {
	// This would use balanced path due to WHERE condition
	results, err := r.repository.Where(map[string]any{"status": status}).Get()
	if err != nil {
		return 0, err
	}
	return int64(len(results)), nil
}

// Advanced Usage Examples

// GetUsersByComplexCriteria demonstrates auto-escalation
func (r *UserRepository) GetUsersByComplexCriteria(
	status string, 
	roles []string, 
	createdAfter time.Time, 
	hasLogins bool,
) ([]models.User, error) {
	// Start with balanced path
	query := r.repository.Where(map[string]any{
		"status": status,
		"created_at_gt": createdAfter,
	})
	
	// This will auto-escalate to complex path due to complexity
	if len(roles) > 0 {
		roleInterfaces := make([]any, len(roles))
		for i, role := range roles {
			roleInterfaces[i] = role
		}
		query = query.WhereIn("role", roleInterfaces)
	}
	
	// If we need joins, manually escalate to complex path
	if hasLogins {
		return r.repository.Complex().
			Join("user_logins", "users.id = user_logins.user_id").
			WithBatching(true).
			Build().
			Get()
	}
	
	return query.Get()
}