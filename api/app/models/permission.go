package models

import (
	"time"
)

// Permission represents a permission in the system
type Permission struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
	Roles       []Role     `json:"roles"`
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
