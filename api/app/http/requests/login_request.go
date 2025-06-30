package requests

import (
	laravel_http "base_lara_go_project/app/core/laravel_core/http"
	"net/http"
)

// LoginRequest handles user login validation
type LoginRequest struct {
	laravel_http.FormRequest
}

// NewLoginRequest creates a new login request
func NewLoginRequest(r *http.Request) *LoginRequest {
	return &LoginRequest{
		FormRequest: *laravel_http.NewFormRequest(r),
	}
}

// Rules returns the validation rules for login
func (r *LoginRequest) Rules() map[string]any {
	return map[string]any{
		"email":    []any{"required", "string", "email"},
		"password": []any{"required", "string"},
	}
}

// Messages returns custom validation messages
func (r *LoginRequest) Messages() map[string]string {
	return map[string]string{
		"email.required":    "Email is required",
		"email.email":       "Please provide a valid email address",
		"password.required": "Password is required",
	}
}

// Authorize determines if the request is authorized
func (r *LoginRequest) Authorize() bool {
	// TODO: Add any authorization logic here
	return true
}
