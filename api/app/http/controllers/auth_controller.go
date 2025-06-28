package controllers

import (
	"base_lara_go_project/app/core"
	"base_lara_go_project/app/data_objects/auth"
	authEvents "base_lara_go_project/app/events/auth"
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/http/requests"
	"base_lara_go_project/app/utils/token"
	"net/http"

	db "base_lara_go_project/app/models/db"

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

	user, err := facades.CreateUser(userData, []string{"customer"})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to DTO using the static method
	userDTO := auth.FromUser(user)

	// Dispatch UserCreated event asynchronously (like event(new UserWasCreated($user)))
	userCreatedEvent := &authEvents.UserCreated{User: userDTO}
	facades.EventAsync(userCreatedEvent)

	c.JSON(http.StatusOK, gin.H{"message": user.GetEmail() + " successfully registered", "roles": user.GetRoles()})
}

func Login(c *gin.Context) {
	var input requests.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Direct service call
	user, err := facades.AuthenticateUser(input.Email, input.Password)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is incorrect."})
		return
	}

	// Generate token
	token, err := token.GenerateToken(user.GetID(), user.GetRoles()[0].GetName())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token, "role": user.GetRoles()[0].GetName()})
}

func CurrentUser(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Direct service call
	user, err := facades.GetUserWithRoles(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userDTO := auth.FromUser(user)
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": userDTO})
}

// TestEmailTemplate demonstrates the new email templating system
func TestEmailTemplate(c *gin.Context) {
	// Create a test user
	testUser := &db.User{
		FirstName:    "John",
		LastName:     "Doe",
		Email:        "john.doe@example.com",
		MobileNumber: "+1234567890",
	}

	// Send a test email using the template facade
	err := facades.MailTemplateToUser(testUser, "auth/welcome", core.EmailTemplateData{
		Subject:        "Test Email Template",
		AppName:        "Base Laravel Go Project",
		RecipientEmail: testUser.Email,
		User:           testUser,
		LoginURL:       "https://app.baselaragoproject.test/login",
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send test email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test email sent successfully to " + testUser.Email})
}
