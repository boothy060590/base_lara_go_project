package repositories

import (
	"fmt"
	"time"

	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/cache"
	"base_lara_go_project/app/models/db"
	"base_lara_go_project/app/models/interfaces"

	"gorm.io/gorm"
)

// UserRepository handles user data operations with cache/database decision logic
type UserRepository struct {
	db    *gorm.DB
	cache core.CacheInterface
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, cache core.CacheInterface) *UserRepository {
	return &UserRepository{
		db:    db,
		cache: cache,
	}
}

// GetDB returns the database connection
func (r *UserRepository) GetDB() *gorm.DB {
	return r.db
}

// FindByID finds a user by ID, trying cache first then database
func (r *UserRepository) FindByID(id uint) (interfaces.UserInterface, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("users:%d:data", id)

	if cachedData, exists := r.cache.Get(cacheKey); exists {
		// Deserialize from cache
		user := &cache.User{}
		if err := user.FromCacheData(cachedData.(map[string]interface{})); err == nil {
			return user, nil
		}
	}

	// If not in cache or deserialization failed, get from database
	dbUser := &db.User{}
	err := r.db.Preload("Roles.Permissions").First(dbUser, id).Error
	if err != nil {
		return nil, err
	}

	// Convert to cache model and store in cache
	cacheUser := r.convertDBToCache(dbUser)
	r.storeInCache(cacheUser)

	return cacheUser, nil
}

// FindByEmail finds a user by email, trying cache first then database
func (r *UserRepository) FindByEmail(email string) (interfaces.UserInterface, error) {
	// Try to get from cache first using email index
	emailCacheKey := fmt.Sprintf("users:email:%s", email)

	if userID, exists := r.cache.Get(emailCacheKey); exists {
		if id, ok := userID.(uint); ok {
			return r.FindByID(id)
		}
	}

	// If not in cache, get from database
	dbUser := &db.User{}
	err := r.db.Preload("Roles.Permissions").Where("email = ?", email).First(dbUser).Error
	if err != nil {
		return nil, err
	}

	// Convert to cache model and store in cache
	cacheUser := r.convertDBToCache(dbUser)
	r.storeInCache(cacheUser)

	// Store email index
	r.cache.Set(emailCacheKey, dbUser.ID, time.Hour)

	return cacheUser, nil
}

// Create creates a new user in database and cache
func (r *UserRepository) Create(userData map[string]interface{}) (interfaces.UserInterface, error) {
	// Create in database
	dbUser := &db.User{}

	// Set fields from userData
	if firstName, ok := userData["first_name"].(string); ok {
		dbUser.FirstName = firstName
	}
	if lastName, ok := userData["last_name"].(string); ok {
		dbUser.LastName = lastName
	}
	if email, ok := userData["email"].(string); ok {
		dbUser.Email = email
	}
	if password, ok := userData["password"].(string); ok {
		dbUser.Password = password
	}
	if mobileNumber, ok := userData["mobile_number"].(string); ok {
		dbUser.MobileNumber = mobileNumber
	}

	err := r.db.Create(dbUser).Error
	if err != nil {
		return nil, err
	}

	// Convert to cache model and store in cache
	cacheUser := r.convertDBToCache(dbUser)
	r.storeInCache(cacheUser)

	return cacheUser, nil
}

// Update updates a user in database and cache
func (r *UserRepository) Update(id uint, userData map[string]interface{}) (interfaces.UserInterface, error) {
	// Update in database
	dbUser := &db.User{}
	err := r.db.First(dbUser, id).Error
	if err != nil {
		return nil, err
	}

	// Update fields from userData
	if firstName, ok := userData["first_name"].(string); ok {
		dbUser.FirstName = firstName
	}
	if lastName, ok := userData["last_name"].(string); ok {
		dbUser.LastName = lastName
	}
	if email, ok := userData["email"].(string); ok {
		dbUser.Email = email
	}
	if password, ok := userData["password"].(string); ok {
		dbUser.Password = password
	}
	if mobileNumber, ok := userData["mobile_number"].(string); ok {
		dbUser.MobileNumber = mobileNumber
	}

	err = r.db.Save(dbUser).Error
	if err != nil {
		return nil, err
	}

	// Reload with relationships
	err = r.db.Preload("Roles.Permissions").First(dbUser, id).Error
	if err != nil {
		return nil, err
	}

	// Convert to cache model and update cache
	cacheUser := r.convertDBToCache(dbUser)
	r.storeInCache(cacheUser)

	return cacheUser, nil
}

// Delete deletes a user from database and cache
func (r *UserRepository) Delete(id uint) error {
	// Delete from database
	err := r.db.Delete(&db.User{}, id).Error
	if err != nil {
		return err
	}

	// Remove from cache
	r.removeFromCache(id)

	return nil
}

// FindByField finds a user by any field
func (r *UserRepository) FindByField(field string, value interface{}) (interfaces.UserInterface, error) {
	dbUser := &db.User{}
	err := r.db.Preload("Roles.Permissions").Where(field+" = ?", value).First(dbUser).Error
	if err != nil {
		return nil, err
	}

	// Convert to cache model and store in cache
	cacheUser := r.convertDBToCache(dbUser)
	r.storeInCache(cacheUser)

	return cacheUser, nil
}

// All gets all users
func (r *UserRepository) All() ([]interfaces.UserInterface, error) {
	var dbUsers []db.User
	err := r.db.Preload("Roles.Permissions").Find(&dbUsers).Error
	if err != nil {
		return nil, err
	}

	var users []interfaces.UserInterface
	for _, dbUser := range dbUsers {
		cacheUser := r.convertDBToCache(&dbUser)
		users = append(users, cacheUser)
	}

	return users, nil
}

// Paginate gets paginated users
func (r *UserRepository) Paginate(page, perPage int) ([]interfaces.UserInterface, int64, error) {
	var dbUsers []db.User
	var total int64

	// Get total count
	err := r.db.Model(&db.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * perPage
	err = r.db.Preload("Roles.Permissions").Offset(offset).Limit(perPage).Find(&dbUsers).Error
	if err != nil {
		return nil, 0, err
	}

	var users []interfaces.UserInterface
	for _, dbUser := range dbUsers {
		cacheUser := r.convertDBToCache(&dbUser)
		users = append(users, cacheUser)
	}

	return users, total, nil
}

// UpdateOrCreate updates or creates a user
func (r *UserRepository) UpdateOrCreate(conditions map[string]interface{}, data map[string]interface{}) (interfaces.UserInterface, error) {
	dbUser := &db.User{}

	// Try to find existing user
	err := r.db.Where(conditions).First(dbUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new user
			return r.Create(data)
		}
		return nil, err
	}

	// Update existing user
	return r.Update(dbUser.ID, data)
}

// DeleteWhere deletes users by conditions
func (r *UserRepository) DeleteWhere(conditions map[string]interface{}) error {
	var users []db.User
	err := r.db.Where(conditions).Find(&users).Error
	if err != nil {
		return err
	}

	// Delete from database
	err = r.db.Where(conditions).Delete(&db.User{}).Error
	if err != nil {
		return err
	}

	// Remove from cache
	for _, user := range users {
		r.removeFromCache(user.ID)
	}

	return nil
}

// Exists checks if a user exists
func (r *UserRepository) Exists(id uint) (bool, error) {
	var count int64
	err := r.db.Model(&db.User{}).Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

// Count counts all users
func (r *UserRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&db.User{}).Count(&count).Error
	return count, err
}

// CountWhere counts users by conditions
func (r *UserRepository) CountWhere(conditions map[string]interface{}) (int64, error) {
	var count int64
	err := r.db.Model(&db.User{}).Where(conditions).Count(&count).Error
	return count, err
}

// convertDBToCache converts a database user to a cache user
func (r *UserRepository) convertDBToCache(dbUser *db.User) *cache.User {
	cacheUser := &cache.User{
		FirstName:     dbUser.FirstName,
		LastName:      dbUser.LastName,
		Email:         dbUser.Email,
		Password:      dbUser.Password,
		ResetPassword: dbUser.ResetPassword,
		MobileNumber:  dbUser.MobileNumber,
	}

	// Initialize the data map
	cacheUser.Initialize()

	// Set base model data
	cacheUser.Set("id", dbUser.ID)
	cacheUser.Set("created_at", dbUser.CreatedAt)
	cacheUser.Set("updated_at", dbUser.UpdatedAt)
	if dbUser.DeletedAt.Valid {
		cacheUser.Set("deleted_at", dbUser.DeletedAt.Time)
	}

	// Convert roles
	for _, dbRole := range dbUser.Roles {
		cacheRole := &cache.Role{
			Name:        dbRole.Name,
			Description: dbRole.Description,
		}

		// Initialize role data map
		cacheRole.Initialize()

		cacheRole.Set("id", dbRole.ID)
		cacheRole.Set("created_at", dbRole.CreatedAt)
		cacheRole.Set("updated_at", dbRole.UpdatedAt)

		// Convert permissions
		for _, dbPermission := range dbRole.Permissions {
			cachePermission := &cache.Permission{
				Name:        dbPermission.Name,
				Description: dbPermission.Description,
			}

			// Initialize permission data map
			cachePermission.Initialize()

			cachePermission.Set("id", dbPermission.ID)
			cachePermission.Set("created_at", dbPermission.CreatedAt)
			cachePermission.Set("updated_at", dbPermission.UpdatedAt)
			cacheRole.Permissions = append(cacheRole.Permissions, *cachePermission)
		}

		cacheUser.Roles = append(cacheUser.Roles, *cacheRole)
	}

	return cacheUser
}

// storeInCache stores a user in cache
func (r *UserRepository) storeInCache(user *cache.User) {
	cacheKey := user.GetCacheKey()
	if cacheKey != "" {
		r.cache.Set(cacheKey, user.GetCacheData(), user.GetCacheTTL())

		// Store email index
		emailCacheKey := fmt.Sprintf("users:email:%s", user.Email)
		r.cache.Set(emailCacheKey, user.GetID(), time.Hour)
	}
}

// removeFromCache removes a user from cache
func (r *UserRepository) removeFromCache(id uint) {
	cacheKey := fmt.Sprintf("users:%d:data", id)
	r.cache.Delete(cacheKey)

	// Also remove any email indexes (we'd need to get the email first, but for simplicity we'll just let them expire)
}
