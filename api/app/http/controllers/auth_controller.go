package controllers

import (
	app_core "base_lara_go_project/app/core/go_core"
	laravel_http "base_lara_go_project/app/core/laravel_core/http"
	"base_lara_go_project/app/http/requests"
	"base_lara_go_project/app/services"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthController handles authentication requests
type AuthController struct {
	laravel_http.BaseController
	userService *services.UserService
	container   *app_core.Container
}

// NewAuthController creates a new auth controller
func NewAuthController(container *app_core.Container) *AuthController {
	// Try to get user service from container first
	var userService *services.UserService
	
	if userServiceInstance, err := container.Resolve("service.user"); err == nil {
		userService = userServiceInstance.(*services.UserService)
	} else {
		// For now, create a mock service when repository is not available
		// This allows the application to start even without database
		userService = services.NewMockUserService()
	}

	return &AuthController{
		BaseController: laravel_http.BaseController{},
		userService:    userService,
		container:      container,
	}
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	// Create and validate request
	request := requests.NewRegisterRequest(c.Request)
	valid, errors := request.Validate()

	if !valid {
		ac.ValidationErrorResponse(c, errors)
		return
	}

	// Create user data map with Laravel-style input access
	userData := map[string]interface{}{
		"first_name":    request.GetString("first_name"),
		"last_name":     request.GetString("last_name"),
		"email":         request.GetString("email"),
		"password":      request.GetString("password"),
		"mobile_number": request.GetString("phone"),
	}

	// Create user
	createdUser, err := ac.userService.CreateUser(userData, []string{"user"})
	if err != nil {
		ac.ServerErrorResponse(c, "Failed to create user: "+err.Error())
		return
	}

	// TODO: Dispatch user created event when event system is properly integrated
	// event := auth_events.NewUserCreatedEvent(createdUser)
	// ac.container.GetEventDispatcher().Dispatch(event)

	ac.CreatedResponse(c, createdUser)
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	// Create and validate request
	request := requests.NewLoginRequest(c.Request)
	valid, errors := request.Validate()

	if !valid {
		ac.ValidationErrorResponse(c, errors)
		return
	}

	// Get credentials with Laravel-style input access
	email := request.GetString("email")
	password := request.GetString("password")

	// Authenticate user
	user, err := ac.userService.AuthenticateUser(email, password)
	if err != nil {
		ac.UnauthorizedResponse(c, "Invalid credentials")
		return
	}

	// Generate token (simplified for now)
	token := "token_" + time.Now().Format("20060102150405")

	// Use optimized login success response
	loginData := map[string]interface{}{
		"user":  user,
		"token": token,
	}
	
	ac.SuccessResponse(c, loginData, "Login successful")
}

// GetProfile returns the authenticated user's profile
func (ac *AuthController) GetProfile(c *gin.Context) {
	// Get user from context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		ac.UnauthorizedResponse(c, "User not authenticated")
		return
	}

	user, err := ac.userService.FindByID(userID.(uint))
	if err != nil {
		ac.NotFoundResponse(c, "User not found: "+err.Error())
		return
	}

	ac.ResourceResponse(c, user)
}

