package models_core

// UserInterface defines the interface for user data
type UserInterface interface {
	GetID() uint
	GetEmail() string
	GetFirstName() string
	GetLastName() string
	GetFullName() string
	GetPassword() string
	GetRoles() []RoleInterface
	GetMobileNumber() string
	GetResetPassword() bool
	HasRole(roleName string) bool
	HasPermission(permissionName string) bool
}

// RoleInterface defines the interface for role data
type RoleInterface interface {
	GetID() uint
	GetName() string
	GetDescription() string
	HasPermission(permissionName string) bool
}

// PermissionInterface defines the interface for permission data
type PermissionInterface interface {
	GetID() uint
	GetName() string
	GetDescription() string
}
