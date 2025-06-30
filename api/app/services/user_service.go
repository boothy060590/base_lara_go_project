package services

import (
	"base_lara_go_project/app/models"
	"base_lara_go_project/app/repositories"
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// UserService handles user business logic
type UserService struct {
	userRepo *repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// BaseServiceInterface implementation

// Create creates a new user
func (s *UserService) Create(data map[string]interface{}) (*models.User, error) {
	// Convert map to User
	user := &models.User{}
	// TODO: Implement proper mapping from map to User
	return user, nil
}

// CreateWithContext creates a new user with context
func (s *UserService) CreateWithContext(ctx context.Context, data map[string]interface{}) (*models.User, error) {
	return s.Create(data) // Repository doesn't support context yet
}

// FindByID finds a user by ID
func (s *UserService) FindByID(id uint) (*models.User, error) {
	user, err := s.userRepo.Find(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindByIDWithContext finds a user by ID with context
func (s *UserService) FindByIDWithContext(ctx context.Context, id uint) (*models.User, error) {
	return s.FindByID(id) // Repository doesn't support context yet
}

// FindByField finds a user by field
func (s *UserService) FindByField(field string, value interface{}) (*models.User, error) {
	// TODO: Implement generic field search
	return nil, errors.New("not implemented")
}

// FindByFieldWithContext finds a user by field with context
func (s *UserService) FindByFieldWithContext(ctx context.Context, field string, value interface{}) (*models.User, error) {
	return s.FindByField(field, value) // Repository doesn't support context yet
}

// All gets all users
func (s *UserService) All() ([]*models.User, error) {
	users, _, err := s.userRepo.FindAll(1, 1000) // Get first 1000 users
	if err != nil {
		return nil, err
	}

	// Convert to pointer slice
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}
	return result, nil
}

// AllWithContext gets all users with context
func (s *UserService) AllWithContext(ctx context.Context) ([]*models.User, error) {
	return s.All() // Repository doesn't support context yet
}

// Paginate gets paginated users
func (s *UserService) Paginate(page, perPage int) ([]*models.User, int64, error) {
	users, total, err := s.userRepo.FindAll(page, perPage)
	if err != nil {
		return nil, 0, err
	}

	// Convert to pointer slice
	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}
	return result, total, nil
}

// PaginateWithContext gets paginated users with context
func (s *UserService) PaginateWithContext(ctx context.Context, page, perPage int) ([]*models.User, int64, error) {
	return s.Paginate(page, perPage) // Repository doesn't support context yet
}

// Update updates a user
func (s *UserService) Update(id uint, data map[string]interface{}) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.Find(id)
	if err != nil {
		return nil, err
	}

	// TODO: Update user fields from data map
	err = s.userRepo.Update(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateWithContext updates a user with context
func (s *UserService) UpdateWithContext(ctx context.Context, id uint, data map[string]interface{}) (*models.User, error) {
	return s.Update(id, data) // Repository doesn't support context yet
}

// UpdateOrCreate updates or creates a user
func (s *UserService) UpdateOrCreate(conditions map[string]interface{}, data map[string]interface{}) (*models.User, error) {
	// TODO: Implement update or create logic
	return nil, errors.New("not implemented")
}

// UpdateOrCreateWithContext updates or creates a user with context
func (s *UserService) UpdateOrCreateWithContext(ctx context.Context, conditions map[string]interface{}, data map[string]interface{}) (*models.User, error) {
	return s.UpdateOrCreate(conditions, data) // Repository doesn't support context yet
}

// Delete deletes a user
func (s *UserService) Delete(id uint) error {
	return s.userRepo.Delete(id)
}

// DeleteWithContext deletes a user with context
func (s *UserService) DeleteWithContext(ctx context.Context, id uint) error {
	return s.Delete(id) // Repository doesn't support context yet
}

// DeleteWhere deletes users by conditions
func (s *UserService) DeleteWhere(conditions map[string]interface{}) error {
	// TODO: Implement delete where logic
	return errors.New("not implemented")
}

// DeleteWhereWithContext deletes users by conditions with context
func (s *UserService) DeleteWhereWithContext(ctx context.Context, conditions map[string]interface{}) error {
	return s.DeleteWhere(conditions) // Repository doesn't support context yet
}

// Exists checks if a user exists
func (s *UserService) Exists(id uint) (bool, error) {
	return s.userRepo.Exists(id)
}

// ExistsWithContext checks if a user exists with context
func (s *UserService) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	return s.Exists(id) // Repository doesn't support context yet
}

// Count counts all users
func (s *UserService) Count() (int64, error) {
	return s.userRepo.Count()
}

// CountWithContext counts all users with context
func (s *UserService) CountWithContext(ctx context.Context) (int64, error) {
	return s.Count() // Repository doesn't support context yet
}

// CountWhere counts users by conditions
func (s *UserService) CountWhere(conditions map[string]interface{}) (int64, error) {
	// TODO: Implement count where logic
	return 0, errors.New("not implemented")
}

// CountWhereWithContext counts users by conditions with context
func (s *UserService) CountWhereWithContext(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	return s.CountWhere(conditions) // Repository doesn't support context yet
}

// Business Logic Methods

// CreateUser creates a new user with business validation and role assignment
func (s *UserService) CreateUser(userData map[string]interface{}, roleNames []string) (*models.User, error) {
	// Business validation
	if err := s.validateUserData(userData); err != nil {
		return nil, err
	}

	// Check if user already exists
	if email, ok := userData["email"].(string); ok {
		existingUser, _ := s.userRepo.FindByEmail(email)
		if existingUser != nil {
			return nil, errors.New("user with this email already exists")
		}
	}

	// TODO: Convert userData to User
	user := &models.User{}
	// TODO: Set user fields from userData

	// Create user
	err := s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	// TODO: Implement role assignment
	// This would require access to role repository through dependency injection

	return user, nil
}

// AuthenticateUser validates user credentials
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active (business rule)
	if !s.isUserActive(user) {
		return nil, errors.New("user account is inactive")
	}

	return user, nil
}

// UpdateUserProfile updates user profile with business validation
func (s *UserService) UpdateUserProfile(id uint, userData map[string]interface{}) (*models.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.Find(id)
	if err != nil {
		return nil, err
	}

	// Business validation
	if err := s.validateProfileUpdate(userData); err != nil {
		return nil, err
	}

	// Check email uniqueness if email is being updated
	if email, ok := userData["email"].(string); ok && email != existingUser.Email {
		userWithEmail, _ := s.userRepo.FindByEmail(email)
		if userWithEmail != nil {
			return nil, errors.New("email is already taken")
		}
	}

	// TODO: Update user fields from userData
	err = s.userRepo.Update(existingUser)
	if err != nil {
		return nil, err
	}

	return existingUser, nil
}

// DeactivateUser deactivates a user account
func (s *UserService) DeactivateUser(id uint) error {
	user, err := s.userRepo.Find(id)
	if err != nil {
		return err
	}

	// Business rule: Cannot deactivate admin users
	if s.isAdminUser(user) {
		return errors.New("cannot deactivate admin users")
	}

	// TODO: Update user status
	// user.Status = "inactive"
	err = s.userRepo.Update(user)
	return err
}

// GetUserWithRoles gets a user with their roles and permissions
func (s *UserService) GetUserWithRoles(id uint) (*models.User, error) {
	user, err := s.userRepo.Find(id)
	if err != nil {
		return nil, err
	}

	// Always return user with roles - no permission check needed
	// Roles and permissions are essential for user authentication and authorization
	return user, nil
}

// SearchUsers searches users with business rules
func (s *UserService) SearchUsers(query string, currentUser *models.User) ([]*models.User, error) {
	// Business rule: Only admin users can search all users
	if !s.isAdminUser(currentUser) {
		return nil, errors.New("insufficient permissions to search users")
	}

	// This would typically use a more sophisticated search
	// For now, we'll return all users (repository should implement proper search)
	return s.All()
}

// Private helper methods for business logic

// validateUserData validates user data according to business rules
func (s *UserService) validateUserData(userData map[string]interface{}) error {
	// Email validation
	if email, ok := userData["email"].(string); ok {
		if !s.isValidEmail(email) {
			return errors.New("invalid email format")
		}
	}

	// Password validation
	if password, ok := userData["password"].(string); ok {
		if len(password) < 8 {
			return errors.New("password must be at least 8 characters long")
		}
	}

	// Name validation
	if firstName, ok := userData["first_name"].(string); ok {
		if len(firstName) < 2 {
			return errors.New("first name must be at least 2 characters long")
		}
	}

	return nil
}

// validateProfileUpdate validates profile update data
func (s *UserService) validateProfileUpdate(userData map[string]interface{}) error {
	// Similar to validateUserData but with different rules
	return s.validateUserData(userData)
}

// isUserActive checks if a user account is active
func (s *UserService) isUserActive(user *models.User) bool {
	// This would check user status, email verification, etc.
	// For now, return true
	return true
}

// isAdminUser checks if a user is an admin
func (s *UserService) isAdminUser(user *models.User) bool {
	// TODO: Implement admin check logic
	// Check if user has admin role
	return false
}

// isValidEmail validates email format
func (s *UserService) isValidEmail(email string) bool {
	// Simple email validation - in production, use a proper regex
	return len(email) > 0 && len(email) < 255
}
