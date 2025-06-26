package models

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/interfaces"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
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

// BeforeSave is a GORM hook that hashes the password before saving
func (user *User) BeforeSave(tx *gorm.DB) (err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
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

// Laravel-style static methods using core functions

// Find finds a user by ID
func (User) Find(id uint) (*User, error) {
	var user User
	err := core.Model(&user).Preload("Roles").First(&user, id)
	return &user, err
}

// FindByEmail finds a user by email
func (User) FindByEmail(email string) (*User, error) {
	var user User
	err := core.Model(&user).Preload("Roles").Where("email = ?", email).First(&user)
	return &user, err
}

// Create creates a new user
func (User) Create(user *User) error {
	return core.Create(user)
}

// Save saves the user
func (user *User) Save() error {
	return core.Save(user)
}

// Delete deletes the user
func (user *User) Delete() error {
	return core.Delete(user)
}

// Where creates a query builder for users
func (User) Where(query interface{}, args ...interface{}) core.DatabaseInterface {
	return core.Model(&User{}).Where(query, args...)
}

// All retrieves all users
func (User) All() ([]User, error) {
	var users []User
	err := core.Model(&users).Preload("Roles").Find(&users)
	return users, err
}

// WithRoles preloads roles for the user
func (user *User) WithRoles() error {
	return core.Model(user).Preload("Roles").First(user)
}

// Interface methods for events
func (user *User) GetID() uint {
	return user.ID
}

func (user *User) GetEmail() string {
	return user.Email
}

func (user *User) GetFirstName() string {
	return user.FirstName
}

func (user *User) GetLastName() string {
	return user.LastName
}
