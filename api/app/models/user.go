package models

import (
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system with Laravel-style traits
type User struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	FirstName     string         `gorm:"type:varchar(255);not null" json:"first_name"`
	LastName      string         `gorm:"type:varchar(255);not null" json:"last_name"`
	Email         string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password      string         `gorm:"size:255;not null;" json:"password"`
	ResetPassword bool           `gorm:"default:false" json:"reset_password"`
	MobileNumber  string         `gorm:"type:varchar(20)" json:"mobile_number"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Roles         []Role         `gorm:"many2many:user_roles;" json:"roles"`
}

// TableName returns the table name for the User
func (User) TableName() string {
	return "users"
}

// BeforeSave is a GORM hook that hashes the password before saving
func (user *User) BeforeSave(tx *gorm.DB) (err error) {
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
