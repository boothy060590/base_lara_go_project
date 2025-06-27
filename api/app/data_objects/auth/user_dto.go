package auth

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models"
)

type UserDTO struct {
	ID            uint     `json:"id"`
	FirstName     string   `json:"first_name"`
	LastName      string   `json:"last_name"`
	Email         string   `json:"email"`
	MobileNumber  string   `json:"mobile_number"`
	ResetPassword bool     `json:"reset_password"`
	Roles         []string `json:"roles"`
}

func (u UserDTO) GetID() uint { return u.ID }

// FromModel implements BaseDTO interface
func (u UserDTO) FromModel(model interface{}) core.BaseDTO {
	if user, ok := model.(*models.User); ok {
		return FromUser(user)
	}
	return u
}

// FromUser creates a UserDTO from a User model (Laravel-style static method)
func FromUser(user *models.User) UserDTO {
	roles := []string{}
	for _, r := range user.Roles {
		roles = append(roles, r.Name)
	}

	return UserDTO{
		ID:            user.ID,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		MobileNumber:  user.MobileNumber,
		ResetPassword: user.ResetPassword,
		Roles:         roles,
	}
}

// Ensure UserDTO implements BaseDTO
var _ core.BaseDTO = (*UserDTO)(nil)
