# Laravel-Style Validation System

This package provides Laravel-style validation with Go generics, supporting the same rule syntax and developer experience as Laravel.

## Features

- ✅ Laravel-style rule syntax (`required|string|max:100`)
- ✅ Array-based rules (`["required", "string", Rule::exists(...)]`)
- ✅ Custom validation rules with closures
- ✅ FormRequest classes for complex validation
- ✅ Built-in Laravel rules (required, string, email, max, min, unique, exists)
- ✅ Custom validation messages
- ✅ Authorization support
- ✅ Type-safe with Go generics

## Basic Usage

### Simple Validation

```go
package main

import (
    "base_lara_go_project/app/core/laravel_core/validators"
)

type User struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    user := User{
        Name:  "John Doe",
        Email: "john@example.com",
    }

    // Laravel-style rules
    rules := map[string]any{
        "name":  "required|string|max:100",
        "email": []any{"required", "email", "unique:users,email"},
    }

    validator := validators.NewValidator(user)
    valid, errors := validator.Rules(rules).Validate()

    if !valid {
        // Handle validation errors
        for field, fieldErrors := range errors {
            for _, error := range fieldErrors {
                fmt.Printf("%s: %s\n", field, error)
            }
        }
    }
}
```

### Using Rule Helper Functions

```go
import (
    "base_lara_go_project/app/core/laravel_core/validators"
)

func validateUser(user User) error {
    rules := map[string]any{
        "name": []any{
            validators.Rule.Required(),
            validators.Rule.String(),
            validators.Rule.Max(100),
        },
        "email": []any{
            validators.Rule.Required(),
            validators.Rule.Email(),
            validators.Rule.Unique("users", "email"),
        },
    }

    validator := validators.NewValidator(user)
    valid, errors := validator.Rules(rules).Validate()

    if !valid {
        return fmt.Errorf("validation failed: %v", errors)
    }

    return nil
}
```

### Custom Validation Rules

```go
func validateWithCustomRule(user User) error {
    // Custom rule with closure
    customRule := validators.Rule.Custom(func(value any, data map[string]any) error {
        if value == nil {
            return nil
        }
        
        email, ok := value.(string)
        if !ok {
            return fmt.Errorf("email must be a string")
        }
        
        // Custom business logic
        if strings.Contains(email, "spam") {
            return fmt.Errorf("spam emails are not allowed")
        }
        
        return nil
    })

    rules := map[string]any{
        "email": []any{
            validators.Rule.Required(),
            validators.Rule.Email(),
            customRule,
        },
    }

    validator := validators.NewValidator(user)
    valid, errors := validator.Rules(rules).Validate()

    if !valid {
        return fmt.Errorf("validation failed: %v", errors)
    }

    return nil
}
```

## FormRequest Classes

### Creating a FormRequest

```go
package requests

import (
    "base_lara_go_project/app/core/laravel_core/validators"
)

type LoginRequest struct {
    validators.FormRequest[LoginData]
}

type LoginData struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func NewLoginRequest(data LoginData) *LoginRequest {
    return &LoginRequest{
        FormRequest: *validators.NewFormRequest(data),
    }
}

// Rules returns the validation rules
func (r *LoginRequest) Rules() map[string]any {
    return map[string]any{
        "email":    []any{"required", "email"},
        "password": []any{"required", "string", "min:8"},
    }
}

// Messages returns custom validation messages
func (r *LoginRequest) Messages() map[string]string {
    return map[string]string{
        "email.required":    "Email address is required",
        "email.email":       "Please enter a valid email address",
        "password.required": "Password is required",
        "password.min":      "Password must be at least 8 characters",
    }
}

// Authorize determines if the request is authorized
func (r *LoginRequest) Authorize() bool {
    // Add any authorization logic here
    return true
}

// Helper methods for accessing validated data
func (r *LoginRequest) GetEmail() string {
    if r.Passes() {
        return r.Data.Email
    }
    return ""
}

func (r *LoginRequest) GetPassword() string {
    if r.Passes() {
        return r.Data.Password
    }
    return ""
}
```

### Using FormRequest in Controllers

```go
func (c *AuthController) Login(ctx *gin.Context) {
    var loginData requests.LoginData
    if err := ctx.ShouldBindJSON(&loginData); err != nil {
        ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Create and validate request
    request := requests.NewLoginRequest(loginData)
    
    if request.Fails() {
        ctx.JSON(422, gin.H{
            "message": "Validation failed",
            "errors":  request.Errors(),
        })
        return
    }

    // Use validated data
    email := request.GetEmail()
    password := request.GetPassword()

    // Process login...
}
```

## Built-in Rules

### String Rules
- `required` - Field must be present and not empty
- `string` - Field must be a string
- `email` - Field must be a valid email address
- `max:100` - Field must not exceed 100 characters
- `min:8` - Field must be at least 8 characters

### Database Rules
- `unique:users,email` - Field must be unique in the users table
- `exists:users,id` - Field must exist in the users table

### Numeric Rules
- `int` - Field must be an integer
- `numeric` - Field must be numeric

### Array Rules
- `array` - Field must be an array
- `min:1` - Array must have at least 1 item
- `max:10` - Array must not have more than 10 items

## Custom Messages

```go
messages := map[string]string{
    "name.required":     "Please provide your name",
    "email.required":    "Email address is required",
    "email.email":       "Please enter a valid email address",
    "password.required": "Password is required",
    "password.min":      "Password must be at least 8 characters",
}

validator := validators.NewValidator(user)
valid, errors := validator.Rules(rules).Messages(messages).Validate()
```

## Conditional Validation

```go
func validateConditionally(user User, isUpdate bool) error {
    rules := map[string]any{
        "name":  "required|string|max:100",
        "email": "required|email",
    }

    // Add password validation only for new users
    if !isUpdate {
        rules["password"] = "required|string|min:8"
    }

    validator := validators.NewValidator(user)
    valid, errors := validator.Rules(rules).Validate()

    if !valid {
        return fmt.Errorf("validation failed: %v", errors)
    }

    return nil
}
```

## Advanced Usage

### Nested Validation

```go
type Address struct {
    Street  string `json:"street"`
    City    string `json:"city"`
    Country string `json:"country"`
}

type User struct {
    Name    string  `json:"name"`
    Email   string  `json:"email"`
    Address Address `json:"address"`
}

func validateUserWithAddress(user User) error {
    rules := map[string]any{
        "name":  "required|string|max:100",
        "email": "required|email",
        "address": map[string]any{
            "street":  "required|string",
            "city":    "required|string",
            "country": "required|string",
        },
    }

    validator := validators.NewValidator(user)
    valid, errors := validator.Rules(rules).Validate()

    if !valid {
        return fmt.Errorf("validation failed: %v", errors)
    }

    return nil
}
```

### Array Validation

```go
type User struct {
    Name   string   `json:"name"`
    Emails []string `json:"emails"`
}

func validateUserWithEmails(user User) error {
    rules := map[string]any{
        "name":   "required|string|max:100",
        "emails": "required|array|min:1|max:5",
        "emails.*": "email", // Validate each email in the array
    }

    validator := validators.NewValidator(user)
    valid, errors := validator.Rules(rules).Validate()

    if !valid {
        return fmt.Errorf("validation failed: %v", errors)
    }

    return nil
}
```

## Integration with Service Container

The validation system is automatically registered with the service container:

```go
// Get validation service
validationInstance, err := container.Resolve("validation.service")
if err != nil {
    return err
}

validationService := validationInstance.(*validators.ValidationService)

// Use validation service
valid, errors := validationService.Validate(user, rules)
```

## Performance Notes

- The validation system uses Go generics for type safety
- Rules are parsed once and cached for reuse
- Custom rules with closures provide maximum flexibility
- FormRequest classes provide a clean separation of concerns

## Best Practices

1. **Use FormRequest classes** for complex validation logic
2. **Keep rules simple** and readable
3. **Use custom messages** for better user experience
4. **Implement authorization** in FormRequest classes
5. **Create reusable custom rules** for common business logic
6. **Use conditional validation** for different scenarios (create vs update) 