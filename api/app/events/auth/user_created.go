package auth

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/data_objects/auth"
	"time"
)

type UserCreated struct {
	app_core.Event[auth.UserDTO]
}

func NewUserCreatedEvent(userData auth.UserDTO) *app_core.Event[auth.UserDTO] {
	return &app_core.Event[auth.UserDTO]{
		ID:        "user_created_" + time.Now().Format("20060102150405"),
		Name:      "user.created",
		Data:      userData,
		Timestamp: time.Now(),
		Source:    "auth_controller",
	}
}
