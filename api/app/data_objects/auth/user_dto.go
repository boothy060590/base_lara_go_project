package auth

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/models/interfaces"
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
	if user, ok := model.(interfaces.UserInterface); ok {
		return FromUser(user)
	}
	return u
}

// FromUser creates a UserDTO from a UserInterface
func FromUser(user interfaces.UserInterface) UserDTO {
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
var _ core.BaseDTO = (*UserDTO)(nil)
