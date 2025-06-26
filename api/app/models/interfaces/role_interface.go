package interfaces

// RoleInterface defines the interface for role data
type RoleInterface interface {
	GetID() uint
	GetName() string
	GetDescription() string
	HasPermission(permissionName string) bool
}
