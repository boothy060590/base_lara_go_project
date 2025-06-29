package facades_core

import (
	models_core "base_lara_go_project/app/models/interfaces"
	"errors"
)

// Global service instances
var globalUserService interface{}

// Service facade provides Laravel-style static access to services
type Service struct{}

// User provides static access to user service operations
func (s *Service) User() *UserServiceFacade {
	return &UserServiceFacade{}
}

// Global service instance
var ServiceInstance = &Service{}

// UserServiceFacade provides static methods for user operations
type UserServiceFacade struct{}

// Create creates a new user with business validation
func (u *UserServiceFacade) Create(userData map[string]interface{}, roleNames []string) (models_core.UserInterface, error) {
	if globalUserService == nil {
		return nil, errors.New("user service not found")
	}
	if userService, ok := globalUserService.(interface {
		CreateUser(userData map[string]interface{}, roleNames []string) (models_core.UserInterface, error)
	}); ok {
		return userService.CreateUser(userData, roleNames)
	}
	return nil, errors.New("user service not found")
}

// Authenticate authenticates a user
func (u *UserServiceFacade) Authenticate(email, password string) (models_core.UserInterface, error) {
	if globalUserService == nil {
		return nil, errors.New("user service not found")
	}
	if userService, ok := globalUserService.(interface {
		AuthenticateUser(email, password string) (models_core.UserInterface, error)
	}); ok {
		return userService.AuthenticateUser(email, password)
	}
	return nil, errors.New("user service not found")
}

// UpdateProfile updates a user profile with business validation
func (u *UserServiceFacade) UpdateProfile(id uint, userData map[string]interface{}) (models_core.UserInterface, error) {
	if globalUserService == nil {
		return nil, errors.New("user service not found")
	}
	if userService, ok := globalUserService.(interface {
		UpdateUserProfile(id uint, userData map[string]interface{}) (models_core.UserInterface, error)
	}); ok {
		return userService.UpdateUserProfile(id, userData)
	}
	return nil, errors.New("user service not found")
}

// Deactivate deactivates a user account
func (u *UserServiceFacade) Deactivate(id uint) error {
	if globalUserService == nil {
		return errors.New("user service not found")
	}
	if userService, ok := globalUserService.(interface {
		DeactivateUser(id uint) error
	}); ok {
		return userService.DeactivateUser(id)
	}
	return errors.New("user service not found")
}

// GetWithRoles gets a user with roles and permissions
func (u *UserServiceFacade) GetWithRoles(id uint) (models_core.UserInterface, error) {
	if globalUserService == nil {
		return nil, errors.New("user service not found")
	}
	if userService, ok := globalUserService.(interface {
		GetUserWithRoles(id uint) (models_core.UserInterface, error)
	}); ok {
		return userService.GetUserWithRoles(id)
	}
	return nil, errors.New("user service not found")
}

// Search searches users with business rules
func (u *UserServiceFacade) Search(query string, currentUser models_core.UserInterface) ([]models_core.UserInterface, error) {
	if globalUserService == nil {
		return nil, errors.New("user service not found")
	}
	if userService, ok := globalUserService.(interface {
		SearchUsers(query string, currentUser models_core.UserInterface) ([]models_core.UserInterface, error)
	}); ok {
		return userService.SearchUsers(query, currentUser)
	}
	return nil, errors.New("user service not found")
}

// Helper functions for easy access (Laravel-style static methods)

// User creates a new user
func User() *UserServiceFacade {
	return ServiceInstance.User()
}

// CreateUser creates a new user (static helper)
func CreateUser(userData map[string]interface{}, roleNames []string) (models_core.UserInterface, error) {
	return User().Create(userData, roleNames)
}

// AuthenticateUser authenticates a user (static helper)
func AuthenticateUser(email, password string) (models_core.UserInterface, error) {
	return User().Authenticate(email, password)
}

// UpdateUserProfile updates a user profile (static helper)
func UpdateUserProfile(id uint, userData map[string]interface{}) (models_core.UserInterface, error) {
	return User().UpdateProfile(id, userData)
}

// DeactivateUser deactivates a user (static helper)
func DeactivateUser(id uint) error {
	return User().Deactivate(id)
}

// GetUserWithRoles gets a user with roles (static helper)
func GetUserWithRoles(id uint) (models_core.UserInterface, error) {
	return User().GetWithRoles(id)
}

// SearchUsers searches users (static helper)
func SearchUsers(query string, currentUser models_core.UserInterface) ([]models_core.UserInterface, error) {
	return User().Search(query, currentUser)
}

// SetUserService sets the global user service instance
func SetUserService(service interface{}) {
	globalUserService = service
}
