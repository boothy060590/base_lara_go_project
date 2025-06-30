package models

import (
	"time"

	"gorm.io/gorm"
)

// Permission represents a permission in the system
type Permission struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Name        string         `gorm:"type:varchar(64);unique;not null" json:"name"`
	Description string         `gorm:"type:varchar(255)" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Roles       []Role         `gorm:"many2many:role_permissions;" json:"roles"`
}

// TableName returns the table name for the Permission model
func (Permission) TableName() string {
	return "permissions"
}

// GetName returns the permission name
func (p *Permission) GetName() string {
	return p.Name
}

// GetDescription returns the permission description
func (p *Permission) GetDescription() string {
	return p.Description
}

// IsAssignedTo checks if the permission is assigned to a specific role
func (p *Permission) IsAssignedTo(roleName string) bool {
	for _, role := range p.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}
