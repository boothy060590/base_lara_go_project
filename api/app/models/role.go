package models

import (
	"time"

	"gorm.io/gorm"
)

// Role represents a role in the system
type Role struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `gorm:"type:varchar(64);unique;not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Users       []User         `gorm:"many2many:user_roles;" json:"users"`
	Permissions []Permission   `gorm:"many2many:role_permissions;" json:"permissions"`
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
