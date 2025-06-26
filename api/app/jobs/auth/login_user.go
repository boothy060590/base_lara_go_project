package auth

import (
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/models"
	"base_lara_go_project/app/utils/token"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type LoginUserJob struct {
	Username string
	Password string
}

// LoginResult holds the login response data
type LoginResult struct {
	Token string
	Role  string
}

// Handle processes the job and returns the result (like Laravel's handle method)
func (j *LoginUserJob) Handle() (any, error) {
	var err error

	user := models.User{}

	err = facades.Database.Preload("Roles").Where("username = ?", j.Username).First(&user)

	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(j.Password))

	if err != nil && errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return nil, err
	}

	roleName := ""
	if len(user.Roles) > 0 {
		roleName = user.Roles[0].Name
	}

	jwtToken, err := token.GenerateToken(user.ID, roleName)

	if err != nil {
		return nil, err
	}

	return &LoginResult{
		Token: jwtToken,
		Role:  roleName,
	}, nil
}
