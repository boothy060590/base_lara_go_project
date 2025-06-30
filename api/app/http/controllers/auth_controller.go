package controllers

import (
	app_core "base_lara_go_project/app/core/go_core"
	"base_lara_go_project/app/http/requests"
	"base_lara_go_project/app/models"
	"base_lara_go_project/app/repositories"
	"base_lara_go_project/app/services"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthController handles authentication requests
type AuthController struct {
	userService *services.UserService
	container   *app_core.Container
}

// NewAuthController creates a new auth controller
func NewAuthController(container *app_core.Container) *AuthController {
	// Get user repository from container
	userRepoInstance, err := container.Resolve("repository.user")
	if err != nil {
		panic("User repository not found in container")
	}
	userRepo := userRepoInstance.(*repositories.UserRepository)

	userService := services.NewUserService(userRepo)
	return &AuthController{
		userService: userService,
		container:   container,
	}
}

// Register handles user registration
func (ac *AuthController) Register(c *gin.Context) {
	// Create and validate request
	request := requests.NewRegisterRequest(c.Request)
	valid, errors := request.Validate()

	if !valid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  errors,
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	// TODO: Dispatch user created event when event system is properly integrated
	// event := auth_events.NewUserCreatedEvent(createdUser)
	// ac.container.GetEventDispatcher().Dispatch(event)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user":    createdUser,
	})
}

// Login handles user login
func (ac *AuthController) Login(c *gin.Context) {
	// Create and validate request
	request := requests.NewLoginRequest(c.Request)
	valid, errors := request.Validate()

	if !valid {
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"message": "Validation failed",
			"errors":  errors,
		})
		return
	}

	// Get credentials with Laravel-style input access
	email := request.GetString("email")
	password := request.GetString("password")

	// Authenticate user
	user, err := ac.userService.AuthenticateUser(email, password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Invalid credentials",
			"error":   err.Error(),
		})
		return
	}

	// Generate token (simplified for now)
	token := "token_" + time.Now().Format("20060102150405")

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user":    user,
		"token":   token,
	})
}

// GetProfile returns the authenticated user's profile
func (ac *AuthController) GetProfile(c *gin.Context) {
	// Get user from context (set by middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
		})
		return
	}

	user, err := ac.userService.FindByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "User not found",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

// UserService handles user business logic
type UserService struct {
	userRepo  app_core.Repository[models.User]
	userCache app_core.Cache[models.User]
}

// NewUserService creates a new user service
func NewUserService(userRepo app_core.Repository[models.User], userCache app_core.Cache[models.User]) *UserService {
	return &UserService{
		userRepo:  userRepo,
		userCache: userCache,
	}
}

// CreateUser creates a new user
func (s *UserService) CreateUser(userData map[string]interface{}) (*models.User, error) {
	// TODO: Implement user creation with password hashing
	// For now, create a placeholder user
	user := &models.User{
		ID:        1,
		FirstName: userData["first_name"].(string),
		LastName:  userData["last_name"].(string),
		Email:     userData["email"].(string),
		Password:  userData["password"].(string), // Will be hashed by BeforeSave hook
	}

	return user, nil
}

// AuthenticateUser authenticates a user
func (s *UserService) AuthenticateUser(email, password string) (*models.User, error) {
	// TODO: Implement proper authentication with password verification
	// For now, return a placeholder user
	user := &models.User{
		ID:        1,
		FirstName: "Test",
		LastName:  "User",
		Email:     email,
	}

	return user, nil
}
