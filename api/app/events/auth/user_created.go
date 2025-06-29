package auth

import (
	app_core "base_lara_go_project/app/core/app"
	events_core "base_lara_go_project/app/core/events"
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
	events_core.RegisterEventFactory("UserCreated", func(data map[string]interface{}) (app_core.EventInterface, error) {
		userData, _ := json.Marshal(data["User"])
		var dto auth.UserDTO
		if err := json.Unmarshal(userData, &dto); err != nil {
			return nil, err
		}
		return &UserCreated{User: dto}, nil
	})
}
