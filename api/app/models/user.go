package models

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system with Laravel-style traits
type User struct {
	ID            uint       `json:"id"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	Password      string     `json:"-"`
	ResetPassword bool       `json:"reset_password"`
	MobileNumber  string     `json:"mobile_number"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty"`
	Roles         []Role     `json:"roles"`
}

// TableName returns the table name for the User
func (User) TableName() string {
	return "users"
}

// HashPassword hashes the user's password if it's not already hashed
func (user *User) HashPassword() error {
	// Only hash if not already hashed
	if !strings.HasPrefix(user.Password, "$2a$") && !strings.HasPrefix(user.Password, "$2b$") && !strings.HasPrefix(user.Password, "$2y$") {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	}
	return nil
}

// GetFullName returns the user's full name
func (user *User) GetFullName() string {
	return user.FirstName + " " + user.LastName
}

// HasRole checks if the user has a specific role
func (user *User) HasRole(roleName string) bool {
	for _, role := range user.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// HasPermission checks if the user has a specific permission
func (user *User) HasPermission(permissionName string) bool {
	for _, role := range user.Roles {
		for _, permission := range role.Permissions {
			if permission.Name == permissionName {
				return true
			}
		}
	}
	return false
}

// IsAdmin checks if the user is an admin
func (user *User) IsAdmin() bool {
	return user.HasRole("admin")
}

// IsActive checks if the user is active
func (user *User) IsActive() bool {
	// TODO: Implement active status check
	return true
}

// GetEmail returns the user's email
func (user *User) GetEmail() string {
	return user.Email
}

// GetFirstName returns the user's first name
func (user *User) GetFirstName() string {
	return user.FirstName
}

// GetLastName returns the user's last name
func (user *User) GetLastName() string {
	return user.LastName
}

// GetPassword returns the user's password
func (user *User) GetPassword() string {
	return user.Password
}

// GetMobileNumber returns the user's mobile number
func (user *User) GetMobileNumber() string {
	return user.MobileNumber
}

// GetResetPassword returns the user's reset password flag
func (user *User) GetResetPassword() bool {
	return user.ResetPassword
}

// GetID returns the user's ID
func (user *User) GetID() uint {
	return user.ID
}

// GetRoles returns the user's roles
func (user *User) GetRoles() []*Role {
	roles := make([]*Role, len(user.Roles))
	for i := range user.Roles {
		roles[i] = &user.Roles[i]
	}
	return roles
}

// GetRoleNames returns the user's role names
func (user *User) GetRoleNames() []string {
	roleNames := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roleNames[i] = role.Name
	}
	return roleNames
}
