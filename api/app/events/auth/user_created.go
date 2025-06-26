package auth

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/data_objects/auth"
	"encoding/json"
)

type UserCreated struct {
	User auth.UserDTO
}

func (e *UserCreated) GetUser() auth.UserDTO {
	return e.User
}

func (e *UserCreated) GetEventName() string {
	return "UserCreated"
}

func init() {
	core.RegisterEventFactory("UserCreated", func(data map[string]interface{}) (core.EventInterface, error) {
		userData, _ := json.Marshal(data["User"])
		var dto auth.UserDTO
		if err := json.Unmarshal(userData, &dto); err != nil {
			return nil, err
		}
		return &UserCreated{User: dto}, nil
	})
}
