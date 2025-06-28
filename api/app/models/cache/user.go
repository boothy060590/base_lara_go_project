package cache

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/interfaces"
	"fmt"
	"time"
)

type User struct {
	core.BaseModelData
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

// GetBaseKey returns the base key for this model type
func (u *User) GetBaseKey() string {
	return "users"
}

// GetID returns the user ID
func (u *User) GetID() uint {
	return u.GetUint("id")
}

// GetCacheKey returns the cache key for this model
func (u *User) GetCacheKey() string {
	id := u.GetID()
	if id == 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d:data", u.GetTableName(), id)
}

// GetCacheTTL returns the TTL for this model's cache
func (u *User) GetCacheTTL() time.Duration {
	return time.Hour
}

// GetCacheData returns the data to be cached
func (u *User) GetCacheData() interface{} {
	// Get base model data
	baseData := u.GetData()

	// Add struct fields to the cache data
	cacheData := map[string]interface{}{
		"first_name":     u.FirstName,
		"last_name":      u.LastName,
		"email":          u.Email,
		"password":       u.Password,
		"reset_password": u.ResetPassword,
		"mobile_number":  u.MobileNumber,
		"roles":          u.Roles,
	}

	// Merge base data with struct data
	for key, value := range baseData {
		cacheData[key] = value
	}

	return cacheData
}

// GetCacheTags returns cache tags for invalidation
func (u *User) GetCacheTags() []string {
	return []string{
		u.GetTableName(),
		fmt.Sprintf("%s:%d", u.GetTableName(), u.GetID()),
	}
}

// FromCacheData populates the model from cached data
func (u *User) FromCacheData(data map[string]interface{}) error {
	// Initialize the data map if it's nil
	u.Initialize()

	// Fill the model with cached data
	u.Fill(data)

	// Populate struct fields from the data map using reflection
	u.populateStructFields(data)

	return nil
}

// populateStructFields populates the struct fields from the data map
func (u *User) populateStructFields(data map[string]interface{}) {
	// Define field mappings for cleaner assignment (Laravel-style)
	fieldMappings := map[string]func(interface{}){
		"first_name": func(value interface{}) {
			if str, ok := value.(string); ok {
				u.FirstName = str
			}
		},
		"last_name": func(value interface{}) {
			if str, ok := value.(string); ok {
				u.LastName = str
			}
		},
		"email": func(value interface{}) {
			if str, ok := value.(string); ok {
				u.Email = str
			}
		},
		"password": func(value interface{}) {
			if str, ok := value.(string); ok {
				u.Password = str
			}
		},
		"mobile_number": func(value interface{}) {
			if str, ok := value.(string); ok {
				u.MobileNumber = str
			}
		},
		"reset_password": func(value interface{}) {
			if b, ok := value.(bool); ok {
				u.ResetPassword = b
			}
		},
	}

	// Apply field mappings using the helper method
	u.FillFields(data, fieldMappings)

	// Handle roles separately since it's more complex
	if rolesData, ok := data["roles"].([]interface{}); ok {
		u.Roles = make([]Role, 0, len(rolesData))
		for _, roleData := range rolesData {
			if roleMap, ok := roleData.(map[string]interface{}); ok {
				role := &Role{}
				role.Initialize()
				role.Fill(roleMap)

				if name, ok := roleMap["name"].(string); ok {
					role.Name = name
				}
				if description, ok := roleMap["description"].(string); ok {
					role.Description = description
				}

				u.Roles = append(u.Roles, *role)
			}
		}
	}
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

// GetCreatedAt returns the created at time
func (u *User) GetCreatedAt() time.Time {
	if createdAt, ok := u.Get("created_at").(time.Time); ok {
		return createdAt
	}
	return time.Time{}
}

// GetUpdatedAt returns the updated at time
func (u *User) GetUpdatedAt() time.Time {
	if updatedAt, ok := u.Get("updated_at").(time.Time); ok {
		return updatedAt
	}
	return time.Time{}
}

// GetDeletedAt returns the deleted at time
func (u *User) GetDeletedAt() *time.Time {
	if deletedAt, ok := u.Get("deleted_at").(*time.Time); ok {
		return deletedAt
	}
	return nil
}
