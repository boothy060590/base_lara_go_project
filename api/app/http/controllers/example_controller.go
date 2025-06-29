package controllers

import (
	exceptions_core "base_lara_go_project/app/core/exceptions"
	facades_core "base_lara_go_project/app/core/facades"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ExampleController demonstrates Laravel-style exception handling
type ExampleController struct{}

// NewExampleController creates a new example controller
func NewExampleController() *ExampleController {
	return &ExampleController{}
}

// ShowUser demonstrates ModelNotFoundException handling
func (c *ExampleController) ShowUser(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		// Report the error and return a response
		exceptions_core.ReportError(err, map[string]interface{}{
			"user_id": ctx.Param("id"),
			"action":  "parse_user_id",
		})

		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	// Simulate finding a user (this would normally come from a service)
	if userID == 0 {
		// Create a ModelNotFoundException (Laravel-style)
		exception := exceptions_core.NewModelNotFoundException("User", userID)

		// Add context to the exception
		exception.AddContext("request_id", ctx.GetString("request_id"))
		exception.AddContext("user_agent", ctx.GetHeader("User-Agent"))

		// Report the exception (this will log to Sentry if configured)
		exceptions_core.Report(exception)

		// Return the error response
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": exception.Message,
			"code":  exception.Code,
		})
		return
	}

	// Simulate successful user retrieval
	ctx.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":    userID,
			"name":  "John Doe",
			"email": "john@example.com",
		},
	})
}

// CreateUser demonstrates ValidationException handling
func (c *ExampleController) CreateUser(ctx *gin.Context) {
	var input struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email" binding:"required,email"`
		Age   int    `json:"age" binding:"required,min=18"`
	}

	if err := ctx.ShouldBindJSON(&input); err != nil {
		// Create validation errors map (Laravel-style)
		validationErrors := map[string][]string{
			"name":  {"The name field is required."},
			"email": {"The email field is required and must be a valid email."},
			"age":   {"The age field is required and must be at least 18."},
		}

		// Create ValidationException
		exception := exceptions_core.NewValidationException(validationErrors)

		// Add request context
		exception.AddContext("request_body", ctx.Request.Body)
		exception.AddContext("ip_address", ctx.ClientIP())

		// Report the validation exception
		exceptions_core.Report(exception)

		// Return validation error response
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": exception.Message,
			"errors":  exception.Errors,
		})
		return
	}

	// Simulate successful user creation
	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"name":  input.Name,
			"email": input.Email,
			"age":   input.Age,
		},
	})
}

// ProtectedRoute demonstrates AuthenticationException handling
func (c *ExampleController) ProtectedRoute(ctx *gin.Context) {
	// Simulate authentication check
	token := ctx.GetHeader("Authorization")
	if token == "" {
		exception := exceptions_core.NewAuthenticationException("No authentication token provided")

		// Add security context
		exception.AddContext("ip_address", ctx.ClientIP())
		exception.AddContext("user_agent", ctx.GetHeader("User-Agent"))
		exception.AddContext("endpoint", ctx.Request.URL.Path)

		// Report the authentication failure
		exceptions_core.Report(exception)

		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": exception.Message,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Access granted to protected route",
	})
}

// AdminRoute demonstrates AuthorizationException handling
func (c *ExampleController) AdminRoute(ctx *gin.Context) {
	// Simulate authorization check
	userRole := ctx.GetHeader("X-User-Role")
	if userRole != "admin" {
		exception := exceptions_core.NewAuthorizationException("Insufficient permissions for admin access")

		// Add authorization context
		exception.AddContext("user_role", userRole)
		exception.AddContext("required_role", "admin")
		exception.AddContext("endpoint", ctx.Request.URL.Path)

		// Report the authorization failure
		exceptions_core.Report(exception)

		ctx.JSON(http.StatusForbidden, gin.H{
			"error": exception.Message,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Admin access granted",
	})
}

// DatabaseQuery demonstrates QueryException handling
func (c *ExampleController) DatabaseQuery(ctx *gin.Context) {
	// Simulate a database query error
	query := "SELECT * FROM users WHERE id = ?"
	bind := []interface{}{"invalid_id"}

	exception := exceptions_core.NewQueryException(
		"Database query failed: invalid input for column 'id'",
		query,
		bind,
	)

	// Add database context
	exception.AddContext("connection", "mysql")
	exception.AddContext("database", "laravel_go")
	exception.AddContext("table", "users")

	// Report the database error
	exceptions_core.Report(exception)

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": "Database operation failed",
		"code":  exception.Code,
	})
}

// FileOperation demonstrates FileNotFoundException handling
func (c *ExampleController) FileOperation(ctx *gin.Context) {
	filePath := ctx.Query("file")
	if filePath == "" {
		filePath = "nonexistent.txt"
	}

	exception := exceptions_core.NewFileNotFoundException(filePath)

	// Add file system context
	exception.AddContext("requested_path", filePath)
	exception.AddContext("current_directory", "/app")

	// Report the file error
	exceptions_core.Report(exception)

	ctx.JSON(http.StatusNotFound, gin.H{
		"error": exception.Message,
		"path":  exception.Path,
	})
}

// Configuration demonstrates ConfigurationException handling
func (c *ExampleController) Configuration(ctx *gin.Context) {
	configKey := ctx.Query("key")
	if configKey == "" {
		configKey = "nonexistent_config"
	}

	exception := exceptions_core.NewConfigurationException(configKey, "Configuration key not found")

	// Add configuration context
	exception.AddContext("environment", "production")
	exception.AddContext("config_file", "config/app.php")

	// Report the configuration error
	exceptions_core.Report(exception)

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": exception.Message,
		"key":   exception.Key,
	})
}

// LoggingExample demonstrates different logging levels
func (c *ExampleController) LoggingExample(ctx *gin.Context) {
	// Debug logging
	facades_core.Debug("Debug message from example controller", map[string]interface{}{
		"user_id": ctx.Query("user_id"),
		"action":  "logging_example",
	})

	// Info logging
	facades_core.Info("User accessed logging example", map[string]interface{}{
		"ip_address": ctx.ClientIP(),
		"user_agent": ctx.GetHeader("User-Agent"),
	})

	// Warning logging
	facades_core.Warning("Deprecated endpoint accessed", map[string]interface{}{
		"endpoint": ctx.Request.URL.Path,
		"method":   ctx.Request.Method,
	})

	// Error logging
	facades_core.Error("An error occurred in logging example", map[string]interface{}{
		"error_code": "LOG_001",
		"timestamp":  "2024-01-01T12:00:00Z",
	})

	// Critical logging
	facades_core.Critical("Critical system issue detected", map[string]interface{}{
		"system_component": "database",
		"severity":         "high",
	})

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Logging examples executed successfully",
		"note":    "Check your logs and Sentry for the logged messages",
	})
}

// ExceptionWithContext demonstrates exception reporting with context
func (c *ExampleController) ExceptionWithContext(ctx *gin.Context) {
	// Simulate a complex error with rich context
	err := exceptions_core.NewException("Complex business logic error", 500)

	// Add rich context to the exception
	err.AddContext("user_id", ctx.Query("user_id"))
	err.AddContext("session_id", ctx.GetHeader("X-Session-ID"))
	err.AddContext("request_id", ctx.GetString("request_id"))
	err.AddContext("business_operation", "user_profile_update")
	err.AddContext("affected_records", []string{"users", "user_profiles", "user_preferences"})
	err.AddContext("performance_metrics", map[string]interface{}{
		"execution_time_ms": 1500,
		"memory_usage_mb":   256,
		"database_queries":  15,
	})

	// Report the exception with context
	exceptions_core.ReportError(err, map[string]interface{}{
		"additional_context": "This is additional context from the controller",
		"request_timestamp":  "2024-01-01T12:00:00Z",
	})

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": "An error occurred with rich context",
		"note":  "Check Sentry for detailed error information",
	})
}
