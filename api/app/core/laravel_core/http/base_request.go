package http

import (
	app_core "base_lara_go_project/app/core/go_core"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BaseRequest provides common request functionality
type BaseRequest struct{}

// Authorize determines if the request is authorized
func (r *BaseRequest) Authorize(ctx *gin.Context) bool {
	// Default implementation - always authorized
	// Override in specific requests
	return true
}

// Rules returns the validation rules
func (r *BaseRequest) Rules() map[string]string {
	// Default implementation - no rules
	// Override in specific requests
	return map[string]string{}
}

// Messages returns custom validation messages
func (r *BaseRequest) Messages() map[string]string {
	// Default implementation - no custom messages
	// Override in specific requests
	return map[string]string{}
}

// Attributes returns custom attribute names
func (r *BaseRequest) Attributes() map[string]string {
	// Default implementation - no custom attributes
	// Override in specific requests
	return map[string]string{}
}

// PrepareForValidation prepares the request for validation
func (r *BaseRequest) PrepareForValidation(ctx *gin.Context) {
	// Default implementation - no preparation needed
	// Override in specific requests
}

// WithValidator sets a custom validator
func (r *BaseRequest) WithValidator(validator interface{}) *BaseRequest {
	// TODO: Implement custom validator support
	return r
}

// Validate validates the request
func (r *BaseRequest) Validate(ctx *gin.Context) error {
	// TODO: Implement validation logic
	return nil
}

// Validated returns only the validated data
func (r *BaseRequest) Validated(ctx *gin.Context) map[string]interface{} {
	// TODO: Implement validated data extraction
	return make(map[string]interface{})
}

// All returns all request data
func (r *BaseRequest) All(ctx *gin.Context) map[string]interface{} {
	// TODO: Implement all data extraction
	return make(map[string]interface{})
}

// Input returns a specific input value
func (r *BaseRequest) Input(ctx *gin.Context, key string) interface{} {
	// TODO: Implement input extraction
	return nil
}

// Only returns only the specified keys
func (r *BaseRequest) Only(ctx *gin.Context, keys []string) map[string]interface{} {
	// TODO: Implement only keys extraction
	return make(map[string]interface{})
}

// Except returns all except the specified keys
func (r *BaseRequest) Except(ctx *gin.Context, keys []string) map[string]interface{} {
	// TODO: Implement except keys extraction
	return make(map[string]interface{})
}

// FormRequest provides Laravel-style form request validation
type FormRequest struct {
	validator *app_core.Validator[map[string]any]
	request   *http.Request
}

// NewFormRequest creates a new form request
func NewFormRequest(r *http.Request) *FormRequest {
	return &FormRequest{
		request: r,
	}
}

// Rules returns the validation rules for this request
func (fr *FormRequest) Rules() map[string]any {
	return map[string]any{}
}

// Messages returns custom validation messages
func (fr *FormRequest) Messages() map[string]string {
	return map[string]string{}
}

// Validate validates the request data
func (fr *FormRequest) Validate() (bool, map[string][]string) {
	// Parse request data
	data := fr.parseRequestData()

	// Create validator
	fr.validator = app_core.NewValidator(data)
	fr.validator.Rules(fr.Rules()).Messages(fr.Messages())

	// Validate
	return fr.validator.Validate()
}

// parseRequestData parses the request data into a map
func (fr *FormRequest) parseRequestData() map[string]any {
	data := make(map[string]any)

	// Parse form data
	if err := fr.request.ParseForm(); err == nil {
		for key, values := range fr.request.Form {
			if len(values) == 1 {
				data[key] = values[0]
			} else {
				data[key] = values
			}
		}
	}

	// Parse multipart form data
	if err := fr.request.ParseMultipartForm(32 << 20); err == nil {
		for key, values := range fr.request.MultipartForm.Value {
			if len(values) == 1 {
				data[key] = values[0]
			} else {
				data[key] = values
			}
		}
	}

	return data
}

// Laravel-style Request Input Access Methods

// Get retrieves a value from the request data
func (fr *FormRequest) Get(key string) any {
	if fr.validator == nil {
		return nil
	}
	return fr.validator.Get(key)
}

// GetString retrieves a string value from the request data
func (fr *FormRequest) GetString(key string) string {
	if fr.validator == nil {
		return ""
	}
	return fr.validator.GetString(key)
}

// GetInt retrieves an integer value from the request data
func (fr *FormRequest) GetInt(key string) int {
	if fr.validator == nil {
		return 0
	}
	return fr.validator.GetInt(key)
}

// GetFloat retrieves a float value from the request data
func (fr *FormRequest) GetFloat(key string) float64 {
	if fr.validator == nil {
		return 0.0
	}
	return fr.validator.GetFloat(key)
}

// GetBool retrieves a boolean value from the request data
func (fr *FormRequest) GetBool(key string) bool {
	if fr.validator == nil {
		return false
	}
	return fr.validator.GetBool(key)
}

// Has checks if a key exists in the request data
func (fr *FormRequest) Has(key string) bool {
	if fr.validator == nil {
		return false
	}
	return fr.validator.Has(key)
}

// Input retrieves a value from the request data (alias for Get)
func (fr *FormRequest) Input(key string) any {
	return fr.Get(key)
}

// All returns all request data
func (fr *FormRequest) All() map[string]any {
	if fr.validator == nil {
		return make(map[string]any)
	}
	return fr.validator.All()
}

// Only returns only the specified keys
func (fr *FormRequest) Only(keys []string) map[string]any {
	if fr.validator == nil {
		return make(map[string]any)
	}
	return fr.validator.Only(keys)
}

// Except returns all except the specified keys
func (fr *FormRequest) Except(keys []string) map[string]any {
	if fr.validator == nil {
		return make(map[string]any)
	}
	return fr.validator.Except(keys)
}

// Validated returns only the validated data (keys that have validation rules)
func (fr *FormRequest) Validated() map[string]any {
	if fr.validator == nil {
		return make(map[string]any)
	}
	return fr.validator.Validated()
}

// Request returns the underlying HTTP request
func (fr *FormRequest) Request() *http.Request {
	return fr.request
}
