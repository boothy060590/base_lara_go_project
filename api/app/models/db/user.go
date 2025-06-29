package db

import (
	database_core "base_lara_go_project/app/core/database"
	"base_lara_go_project/app/models/interfaces"

	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	database_core.DatabaseModel
	FirstName     string `gorm:"type:varchar(255);not null" json:"first_name"`
	LastName      string `gorm:"type:varchar(255);not null" json:"last_name"`
	Email         string `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	Password      string `gorm:"size:255;not null;" json:"password"`
	ResetPassword bool   `gorm:"default:false" json:"reset_password"`
	MobileNumber  string `gorm:"type:varchar(20)" json:"mobile_number"`
	Roles         []Role `gorm:"many2many:user_roles;" json:"roles"`
}

// Ensure User implements UserInterface
var _ interfaces.UserInterface = (*User)(nil)

// TableName returns the table name for the User model
func (User) TableName() string {
	return "users"
}

// GetTableName returns the table name
func (u *User) GetTableName() string {
	return u.TableName()
}

// GetID returns the user ID
func (u *User) GetID() uint {
	return u.Model.ID
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

// GetMobileNumber returns the user's mobile number
func (user *User) GetMobileNumber() string {
	return user.MobileNumber
}

// GetResetPassword returns the user's reset password flag
func (user *User) GetResetPassword() bool {
	return user.ResetPassword
}

// Role model for database
type Role struct {
	database_core.DatabaseModel
	Name        string       `gorm:"type:varchar(64);unique;not null" json:"name"`
	Description string       `gorm:"type:varchar(255)" json:"description"`
	Users       []User       `gorm:"many2many:user_roles;" json:"users"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

// TableName returns the table name for the Role model
func (Role) TableName() string {
	return "roles"
}

// GetTableName returns the table name
func (r *Role) GetTableName() string {
	return r.TableName()
}

// GetID returns the role ID
func (r *Role) GetID() uint {
	return r.Model.ID
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

// Permission model for database
type Permission struct {
	database_core.DatabaseModel
	Name        string `gorm:"type:varchar(64);unique;not null" json:"name"`
	Description string `gorm:"type:varchar(255)" json:"description"`
	Roles       []Role `gorm:"many2many:role_permissions;" json:"roles"`
}

// TableName returns the table name for the Permission model
func (Permission) TableName() string {
	return "permissions"
}

// GetTableName returns the table name
func (p *Permission) GetTableName() string {
	return p.TableName()
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
