package services

import (
	"base_lara_go_project/app/models/interfaces"
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
func NewUserService() (*UserService, error) {
	userRepo, exists := repositories.GetUserRepository()
	if !exists {
		return nil, errors.New("user repository not found")
	}
	return &UserService{userRepo: userRepo}, nil
}

// BaseServiceInterface implementation

// Create creates a new user
func (s *UserService) Create(data map[string]interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.Create(data)
}

// CreateWithContext creates a new user with context
func (s *UserService) CreateWithContext(ctx context.Context, data map[string]interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.Create(data) // Repository doesn't support context yet
}

// FindByID finds a user by ID
func (s *UserService) FindByID(id uint) (interfaces.UserInterface, error) {
	return s.userRepo.FindByID(id)
}

// FindByIDWithContext finds a user by ID with context
func (s *UserService) FindByIDWithContext(ctx context.Context, id uint) (interfaces.UserInterface, error) {
	return s.userRepo.FindByID(id) // Repository doesn't support context yet
}

// FindByField finds a user by field
func (s *UserService) FindByField(field string, value interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.FindByField(field, value)
}

// FindByFieldWithContext finds a user by field with context
func (s *UserService) FindByFieldWithContext(ctx context.Context, field string, value interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.FindByField(field, value) // Repository doesn't support context yet
}

// All gets all users
func (s *UserService) All() ([]interfaces.UserInterface, error) {
	return s.userRepo.All()
}

// AllWithContext gets all users with context
func (s *UserService) AllWithContext(ctx context.Context) ([]interfaces.UserInterface, error) {
	return s.userRepo.All() // Repository doesn't support context yet
}

// Paginate gets paginated users
func (s *UserService) Paginate(page, perPage int) ([]interfaces.UserInterface, int64, error) {
	return s.userRepo.Paginate(page, perPage)
}

// PaginateWithContext gets paginated users with context
func (s *UserService) PaginateWithContext(ctx context.Context, page, perPage int) ([]interfaces.UserInterface, int64, error) {
	return s.userRepo.Paginate(page, perPage) // Repository doesn't support context yet
}

// Update updates a user
func (s *UserService) Update(id uint, data map[string]interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.Update(id, data)
}

// UpdateWithContext updates a user with context
func (s *UserService) UpdateWithContext(ctx context.Context, id uint, data map[string]interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.Update(id, data) // Repository doesn't support context yet
}

// UpdateOrCreate updates or creates a user
func (s *UserService) UpdateOrCreate(conditions map[string]interface{}, data map[string]interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.UpdateOrCreate(conditions, data)
}

// UpdateOrCreateWithContext updates or creates a user with context
func (s *UserService) UpdateOrCreateWithContext(ctx context.Context, conditions map[string]interface{}, data map[string]interface{}) (interfaces.UserInterface, error) {
	return s.userRepo.UpdateOrCreate(conditions, data) // Repository doesn't support context yet
}

// Delete deletes a user
func (s *UserService) Delete(id uint) error {
	return s.userRepo.Delete(id)
}

// DeleteWithContext deletes a user with context
func (s *UserService) DeleteWithContext(ctx context.Context, id uint) error {
	return s.userRepo.Delete(id) // Repository doesn't support context yet
}

// DeleteWhere deletes users by conditions
func (s *UserService) DeleteWhere(conditions map[string]interface{}) error {
	return s.userRepo.DeleteWhere(conditions)
}

// DeleteWhereWithContext deletes users by conditions with context
func (s *UserService) DeleteWhereWithContext(ctx context.Context, conditions map[string]interface{}) error {
	return s.userRepo.DeleteWhere(conditions) // Repository doesn't support context yet
}

// Exists checks if a user exists
func (s *UserService) Exists(id uint) (bool, error) {
	return s.userRepo.Exists(id)
}

// ExistsWithContext checks if a user exists with context
func (s *UserService) ExistsWithContext(ctx context.Context, id uint) (bool, error) {
	return s.userRepo.Exists(id) // Repository doesn't support context yet
}

// Count counts all users
func (s *UserService) Count() (int64, error) {
	return s.userRepo.Count()
}

// CountWithContext counts all users with context
func (s *UserService) CountWithContext(ctx context.Context) (int64, error) {
	return s.userRepo.Count() // Repository doesn't support context yet
}

// CountWhere counts users by conditions
func (s *UserService) CountWhere(conditions map[string]interface{}) (int64, error) {
	return s.userRepo.CountWhere(conditions)
}

// CountWhereWithContext counts users by conditions with context
func (s *UserService) CountWhereWithContext(ctx context.Context, conditions map[string]interface{}) (int64, error) {
	return s.userRepo.CountWhere(conditions) // Repository doesn't support context yet
}

// Business Logic Methods

// CreateUser creates a new user with business validation and role assignment
func (s *UserService) CreateUser(userData map[string]interface{}, roleNames []string) (interfaces.UserInterface, error) {
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

	// Note: Password hashing is handled by the User model's BeforeSave hook

	// Create user
	user, err := s.userRepo.Create(userData)
	if err != nil {
		return nil, err
	}

	// Assign roles
	if len(roleNames) > 0 {
		// Get role repository
		roleRepo, exists := repositories.GetRoleRepository()
		if !exists {
			return nil, errors.New("role repository not found")
		}

		// Find and assign each role
		for _, roleName := range roleNames {
			role, err := roleRepo.FindByName(roleName)
			if err != nil {
				// Log the error but continue with other roles
				continue
			}

			// Assign role to user using GORM association
			// Use the repository's database connection to assign the role
			// This is a simplified approach - in a real app, you'd have a proper role assignment method
			// Get the database connection from the repository
			db := s.userRepo.GetDB()
			if db != nil {
				// Create the association in the user_roles table
				err = db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)", user.GetID(), role.ID).Error
				if err != nil {
					continue
				}
			}
		}

		// Reload user with roles
		user, err = s.userRepo.FindByID(user.GetID())
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

// AuthenticateUser validates user credentials
func (s *UserService) AuthenticateUser(email, password string) (interfaces.UserInterface, error) {
	// Find user by email
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.GetPassword()), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active (business rule)
	if !s.isUserActive(user) {
		return nil, errors.New("user account is inactive")
	}

	return user, nil
}

// UpdateUserProfile updates user profile with business validation
func (s *UserService) UpdateUserProfile(id uint, userData map[string]interface{}) (interfaces.UserInterface, error) {
	// Get existing user
	existingUser, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Business validation
	if err := s.validateProfileUpdate(userData); err != nil {
		return nil, err
	}

	// Check email uniqueness if email is being updated
	if email, ok := userData["email"].(string); ok && email != existingUser.GetEmail() {
		userWithEmail, _ := s.userRepo.FindByEmail(email)
		if userWithEmail != nil {
			return nil, errors.New("email is already taken")
		}
	}

	// Hash password if provided
	if password, ok := userData["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		userData["password"] = string(hashedPassword)
	}

	// Update user
	return s.userRepo.Update(id, userData)
}

// DeactivateUser deactivates a user account
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
	return err
}

// GetUserWithRoles gets a user with their roles and permissions
func (s *UserService) GetUserWithRoles(id uint) (interfaces.UserInterface, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Business logic: Check if user has access to view roles
	if !s.canViewUserRoles(user) {
		return nil, errors.New("insufficient permissions to view user roles")
	}

	return user, nil
}

// SearchUsers searches users with business rules
func (s *UserService) SearchUsers(query string, currentUser interfaces.UserInterface) ([]interfaces.UserInterface, error) {
	// Business rule: Only admin users can search all users
	if !s.isAdminUser(currentUser) {
		return nil, errors.New("insufficient permissions to search users")
	}

	// This would typically use a more sophisticated search
	// For now, we'll return all users (repository should implement proper search)
	return s.userRepo.All()
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
func (s *UserService) isUserActive(user interfaces.UserInterface) bool {
	// This would check user status, email verification, etc.
	// For now, return true
	return true
}

// isAdminUser checks if a user is an admin
func (s *UserService) isAdminUser(user interfaces.UserInterface) bool {
	// Check if user has admin role
	for _, role := range user.GetRoles() {
		if role.GetName() == "admin" {
			return true
		}
	}
	return false
}

// canViewUserRoles checks if user can view other users' roles
func (s *UserService) canViewUserRoles(user interfaces.UserInterface) bool {
	// Business rule: Only admin users can view roles
	return s.isAdminUser(user)
}

// isValidEmail validates email format
func (s *UserService) isValidEmail(email string) bool {
	// Simple email validation - in production, use a proper regex
	return len(email) > 0 && len(email) < 255
}
