package interfaces

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
