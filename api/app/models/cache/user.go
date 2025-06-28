package cache

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/interfaces"
)

type User struct {
	core.CachedModel
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Email         string `json:"email"`
	Password      string `json:"password"`
	ResetPassword bool   `json:"reset_password"`
	MobileNumber  string `json:"mobile_number"`
	Roles         []Role `json:"roles"`
}

// Ensure User implements UserInterface
var _ interfaces.UserInterface = (*User)(nil)

// GetTableName returns the table name
func (u *User) GetTableName() string {
	return "users"
}

// GetID returns the user ID
func (u *User) GetID() uint {
	return u.GetUint("id")
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

// Interface methods for events
func (user *User) GetEmail() string {
	return user.Email
}

func (user *User) GetFirstName() string {
	return user.FirstName
}

func (user *User) GetLastName() string {
	return user.LastName
}

// GetPassword returns the user's password
func (user *User) GetPassword() string {
	return user.Password
}

// GetRoles returns the user's roles
func (user *User) GetRoles() []interfaces.RoleInterface {
	roles := make([]interfaces.RoleInterface, len(user.Roles))
	for i := range user.Roles {
		roles[i] = &user.Roles[i]
	}
	return roles
}

// Role model for cache
type Role struct {
	core.CachedModel
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Users       []User       `json:"users"`
	Permissions []Permission `json:"permissions"`
}

// GetTableName returns the table name
func (r *Role) GetTableName() string {
	return "roles"
}

// GetID returns the role ID
func (r *Role) GetID() uint {
	return r.GetUint("id")
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

// Interface methods for events
func (role *Role) GetName() string {
	return role.Name
}

func (role *Role) GetDescription() string {
	return role.Description
}

// GetPermissions returns the role's permissions
func (role *Role) GetPermissions() []interfaces.PermissionInterface {
	perms := make([]interfaces.PermissionInterface, len(role.Permissions))
	for i := range role.Permissions {
		perms[i] = &role.Permissions[i]
	}
	return perms
}

// Permission model for cache
type Permission struct {
	core.CachedModel
	Name        string `json:"name"`
	Description string `json:"description"`
	Roles       []Role `json:"roles"`
}

// GetTableName returns the table name
func (p *Permission) GetTableName() string {
	return "permissions"
}

// GetID returns the permission ID
func (p *Permission) GetID() uint {
	return p.GetUint("id")
}

// Interface methods for events
func (p *Permission) GetName() string {
	return p.Name
}

func (p *Permission) GetDescription() string {
	return p.Description
}

// IsAssignedTo checks if the permission is assigned to a role with the given name
func (p *Permission) IsAssignedTo(roleName string) bool {
	for _, role := range p.Roles {
		if role.Name == roleName {
			return true
		}
	}
	return false
}

// GetMobileNumber returns the user's mobile number
func (user *User) GetMobileNumber() string {
	return user.MobileNumber
}

// GetResetPassword returns the user's reset password flag
func (user *User) GetResetPassword() bool {
	return user.ResetPassword
}
