package models

import (
	"time"
)

// Role represents a role in the system
type Role struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	DeletedAt   time.Time    `json:"deleted_at,omitempty"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// TableName returns the table name for the Role model
func (Role) TableName() string {
	return "roles"
}

// HasPermission checks if the role has a specific permission
func (role *Role) HasPermission(permissionName string) bool {
	for _, permission := range role.Permissions {
		if permission.Name == permissionName {
			return true
		}
	}
	return false
}

// GetName returns the role name
func (role *Role) GetName() string {
	return role.Name
}

// GetDescription returns the role description
func (role *Role) GetDescription() string {
	return role.Description
}

// GetID returns the role ID
func (role *Role) GetID() uint {
	return role.ID
}
