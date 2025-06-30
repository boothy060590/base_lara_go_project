package auth

import (
	facades_core "base_lara_go_project/app/core/facades"
)

// LoginUserJob handles user login
type LoginUserJob struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Handle processes the login job
func (j *LoginUserJob) Handle() (any, error) {
	// Use service facade for Laravel-style access
	return facades_core.AuthenticateUser(j.Email, j.Password)
}
