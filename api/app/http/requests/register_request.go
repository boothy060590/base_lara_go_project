package requests

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_http "base_lara_go_project/app/core/laravel_core/http"
	"net/http"
)

// RegisterRequest handles user registration validation
type RegisterRequest struct {
	laravel_http.FormRequest
}

// NewRegisterRequest creates a new register request
func NewRegisterRequest(r *http.Request) *RegisterRequest {
	return &RegisterRequest{
		FormRequest: *laravel_http.NewFormRequest(r),
	}
}

// Rules returns the validation rules for registration
func (r *RegisterRequest) Rules() map[string]any {
	return map[string]any{
		"first_name": []any{"required", "string", app_core.Max(50)},
		"last_name":  []any{"required", "string", app_core.Max(50)},
		"email":      []any{"required", "string", "email", app_core.Max(100)},
		"password":   []any{"required", "string", app_core.Min(8), app_core.Max(255)},
		"phone":      []any{"string", app_core.Max(20)},
	}
}

// Messages returns custom validation messages
func (r *RegisterRequest) Messages() map[string]string {
	return map[string]string{
		"first_name.required": "First name is required",
		"last_name.required":  "Last name is required",
		"email.required":      "Email is required",
		"email.email":         "Please provide a valid email address",
		"password.required":   "Password is required",
		"password.min":        "Password must be at least 8 characters",
	}
}

// Authorize determines if the request is authorized
func (r *RegisterRequest) Authorize() bool {
	// TODO: Add any authorization logic here
	return true
}
