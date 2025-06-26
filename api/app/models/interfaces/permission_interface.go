package interfaces

// PermissionInterface defines the interface for permission data
type PermissionInterface interface {
	GetID() uint
	GetName() string
	GetDescription() string
	IsAssignedTo(roleName string) bool
}
