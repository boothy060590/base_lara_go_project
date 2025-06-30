package core

import (
	models_core "base_lara_go_project/app/core/laravel_core/models"
)

// UserDTO represents a user data transfer object
type UserDTO struct {
	ID            uint     `json:"id"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Email         string   `json:"email"`
	MobileNumber  string   `json:"mobile_number"`
	ResetPassword bool     `json:"reset_password"`
	Roles         []string `json:"roles"`
}

// GetID returns the user ID
func (u UserDTO) GetID() uint { return u.ID }

// FromModel creates a UserDTO from a UserInterface
func (u UserDTO) FromModel(model interface{}) BaseDTO {
	if user, ok := model.(models_core.UserInterface); ok {
		return FromUser(user)
	}
	return u
}

// FromUser creates a UserDTO from a UserInterface
func FromUser(user models_core.UserInterface) UserDTO {
	roles := []string{}
	for _, r := range user.GetRoles() {
		roles = append(roles, r.GetName())
	}

	return UserDTO{
		ID:            user.GetID(),
		FirstName:     user.GetFirstName(),
		LastName:      user.GetLastName(),
		Email:         user.GetEmail(),
		MobileNumber:  user.GetMobileNumber(),
		ResetPassword: user.GetResetPassword(),
		Roles:         roles,
	}
}

// Ensure UserDTO implements BaseDTO
var _ BaseDTO = (*UserDTO)(nil)
