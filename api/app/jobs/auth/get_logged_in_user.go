package auth

import (
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/models"
	"errors"
)

type GetLoggedInUserJob struct {
	UserID uint
}

// Handle processes the job and returns the result (like Laravel's handle method)
func (j *GetLoggedInUserJob) Handle() (any, error) {
	var user models.User

	if err := facades.Database.Preload("Roles").First(&user, j.UserID); err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}
