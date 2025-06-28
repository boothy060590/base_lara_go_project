package auth

import (
	"base_lara_go_project/app/facades"
)

// GetLoggedInUserJob handles retrieving logged in user
type GetLoggedInUserJob struct {
	UserID uint `json:"user_id"`
}

// Handle processes the get logged in user job
func (j *GetLoggedInUserJob) Handle() (any, error) {
	// Use service facade for Laravel-style access
	return facades.GetUserWithRoles(j.UserID)
}
