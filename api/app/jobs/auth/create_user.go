package auth

import (
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/models"

	"gorm.io/gorm"
)

type CreateUserJob struct {
	Password  string
	FirstName string
	LastName  string
	Email     string
	Roles     []string
}

// Handle processes the job and returns the result (like Laravel's handle method)
func (j *CreateUserJob) Handle() (any, error) {
	user := models.User{
		Password:  j.Password,
		FirstName: j.FirstName,
		LastName:  j.LastName,
		Email:     j.Email,
	}

	var roles []models.Role
	for _, roleName := range j.Roles {
		var role models.Role
		err := facades.Database.Where("name = ?", roleName).First(&role)
		if err == gorm.ErrRecordNotFound {
			role = models.Role{Name: roleName}
			facades.Database.Create(&role)
		} else if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	user.Roles = roles

	if err := facades.Database.Create(&user); err != nil {
		return nil, err
	}

	// Preload roles for return
	facades.Database.Model(&user).Preload("Roles").First(&user)
	return &user, nil
}
