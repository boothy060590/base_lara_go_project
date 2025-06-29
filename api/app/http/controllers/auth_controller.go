package controllers

import (
	facades_core "base_lara_go_project/app/core/facades"
	"base_lara_go_project/app/data_objects/auth"
	authEvents "base_lara_go_project/app/events/auth"
	"base_lara_go_project/app/http/requests"
	"base_lara_go_project/app/utils/token"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {

	var input requests.RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Direct service call - no job needed for sync operations
	userData := map[string]interface{}{
		"first_name":    input.FirstName,
		"last_name":     input.LastName,
		"email":         input.Email,
		"password":      input.Password,
		"mobile_number": input.MobileNumber,
	}

	user, err := facades_core.CreateUser(userData, []string{"customer"})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to DTO using the static method
	userDTO := auth.FromUser(user)

	// Dispatch UserCreated event asynchronously (like event(new UserWasCreated($user)))
	userCreatedEvent := &authEvents.UserCreated{User: userDTO}
	facades_core.EventAsync(userCreatedEvent)

	c.JSON(http.StatusOK, gin.H{"message": user.GetEmail() + " successfully registered", "roles": user.GetRoles()})
}

func Login(c *gin.Context) {
	var input requests.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Direct service call
	user, err := facades_core.AuthenticateUser(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is incorrect."})
		return
	}

	// Check if user has roles
	roles := user.GetRoles()
	if len(roles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User has no roles assigned."})
		return
	}

	// Generate token
	token, err := token.GenerateToken(user.GetID(), roles[0].GetName())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "role": roles[0].GetName()})
}

func CurrentUser(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Direct service call
	user, err := facades_core.GetUserWithRoles(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userDTO := auth.FromUser(user)
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": userDTO})
}
