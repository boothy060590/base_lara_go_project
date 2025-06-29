package auth

import (
	facades_core "base_lara_go_project/app/core/facades"
)

// CreateUserJob handles user creation
type CreateUserJob struct {
	UserData  map[string]interface{} `json:"user_data"`
	RoleNames []string               `json:"role_names"`
}

// Handle processes the user creation job
func (j *CreateUserJob) Handle() (any, error) {
	// Use service facade for Laravel-style access
	return facades_core.CreateUser(j.UserData, j.RoleNames)
}
