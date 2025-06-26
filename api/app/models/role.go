package models

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/interfaces"

	"gorm.io/gorm"
)

type Role struct {
	core.BaseModel
	gorm.Model
	Name        string       `gorm:"type:varchar(64);unique;not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// Ensure Role implements RoleInterface
var _ interfaces.RoleInterface = (*Role)(nil)

// TableName returns the table name for the Role model
func (Role) TableName() string {
	return "roles"
}

func (role *Role) AfterFind(tx *gorm.DB) (err error) {
	role.BaseModel = *core.NewBaseModel()
	role.BaseModel.Set("id", role.ID)
	role.BaseModel.Set("name", role.Name)
	role.BaseModel.Set("description", role.Description)
	return nil
}

func (role *Role) AfterCreate(tx *gorm.DB) (err error) {
	role.AfterFind(tx)
	return nil
}

func (role *Role) AfterUpdate(tx *gorm.DB) (err error) {
	role.AfterFind(tx)
	return nil
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

// AddPermission adds a permission to the role
func (role *Role) AddPermission(permission *Permission) {
	role.Permissions = append(role.Permissions, *permission)
}

// RemovePermission removes a permission from the role
func (role *Role) RemovePermission(permissionName string) {
	var filteredPermissions []Permission
	for _, permission := range role.Permissions {
		if permission.Name != permissionName {
			filteredPermissions = append(filteredPermissions, permission)
		}
	}
	role.Permissions = filteredPermissions
}

// Interface methods for events
func (role *Role) GetID() uint {
	return role.ID
}

func (role *Role) GetName() string {
	return role.Name
}

func (role *Role) GetDescription() string {
	return role.Description
}

// Laravel-style static methods using core functions

// Find finds a role by ID
func (Role) Find(id uint) (*Role, error) {
	var role Role
	err := core.Model(&role).Preload("Permissions").First(&role, id)
	return &role, err
}

// FindByName finds a role by name
func (Role) FindByName(name string) (*Role, error) {
	var role Role
	err := core.Model(&role).Preload("Permissions").Where("name = ?", name).First(&role)
	return &role, err
}

// Create creates a new role
func (Role) Create(role *Role) error {
	return core.Create(role)
}

// Save saves the role
func (role *Role) Save() error {
	return core.Save(role)
}

// Delete deletes the role
func (role *Role) Delete() error {
	return core.Delete(role)
}

// Where creates a query builder for roles
func (Role) Where(query interface{}, args ...interface{}) core.DatabaseInterface {
	return core.Model(&Role{}).Where(query, args...)
}

// All retrieves all roles
func (Role) All() ([]Role, error) {
	var roles []Role
	err := core.Model(&roles).Preload("Permissions").Find(&roles)
	return roles, err
}

// WithPermissions preloads permissions for the role
func (role *Role) WithPermissions() error {
	return core.Model(role).Preload("Permissions").First(role)
}
