package auth

import "base_lara_go_project/app/core"

type UserDTO struct {
	ID           uint   `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Email        string `json:"email"`
	MobileNumber string `json:"mobile_number"`
}

func (u UserDTO) GetID() uint { return u.ID }

// Ensure UserDTO implements BaseDTO
var _ core.BaseDTO = (*UserDTO)(nil)
