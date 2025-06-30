package auth

import (
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

// FromModel creates a UserDTO from a model
func (u UserDTO) FromModel(model interface{}) UserDTO {
	if user, ok := model.(*models.User); ok {
		return FromUser(user)
	}
	return u
}

// FromUser creates a UserDTO from a User
func FromUser(user *models.User) UserDTO {
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
