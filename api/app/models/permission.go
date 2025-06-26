package models

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/interfaces"

	"gorm.io/gorm"
)

type Permission struct {
	core.BaseModel
	gorm.Model
	Name        string `gorm:"type:varchar(64);unique;not null" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	Roles       []Role `gorm:"many2many:role_permissions;" json:"roles"`
}

// Ensure Permission implements PermissionInterface
var _ interfaces.PermissionInterface = (*Permission)(nil)

// TableName returns the table name for the Permission model
func (Permission) TableName() string {
	return "permissions"
}

func (permission *Permission) AfterFind(tx *gorm.DB) (err error) {
	permission.BaseModel = *core.NewBaseModel()
	permission.BaseModel.Set("id", permission.ID)
	permission.BaseModel.Set("name", permission.Name)
	permission.BaseModel.Set("description", permission.Description)
	return nil
}

func (permission *Permission) AfterCreate(tx *gorm.DB) (err error) {
	permission.AfterFind(tx)
	return nil
}

func (permission *Permission) AfterUpdate(tx *gorm.DB) (err error) {
	permission.AfterFind(tx)
	return nil
}

// IsAssignedTo checks if the permission is assigned to a specific role
func (permission *Permission) IsAssignedTo(roleName string) bool {
	for _, role := range permission.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// Interface methods for events
func (permission *Permission) GetID() uint {
	return permission.ID
}

func (permission *Permission) GetName() string {
	return permission.Name
}

func (permission *Permission) GetDescription() string {
	return permission.Description
}
