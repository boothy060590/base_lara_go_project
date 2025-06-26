package controllers

import (
	"base_lara_go_project/app/data_objects/auth"
	authEvents "base_lara_go_project/app/events/auth"
	"base_lara_go_project/app/facades"
	"base_lara_go_project/app/http/requests"
	authJobs "base_lara_go_project/app/jobs/auth"
	"base_lara_go_project/app/models"
	UserTransformer "base_lara_go_project/app/transformers"
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

	// Create and dispatch the job synchronously (like $user = dispatchSync(new CreateUserJob(...)))
	createUserJob := &authJobs.CreateUserJob{
		Password:  input.Password,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     input.Email,
		Roles:     []string{"customer"},
	}

	result, err := facades.DispatchSync(createUserJob)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := result.(*models.User)

	// Convert to DTO
	userDTO := ToUserDTO(user)

	// Dispatch UserCreated event asynchronously (like event(new UserWasCreated($user)))
	userCreatedEvent := &authEvents.UserCreated{User: userDTO}
	facades.Event(userCreatedEvent)

	c.JSON(http.StatusOK, gin.H{"message": user.Email + " successfully registered", "roles": user.Roles})
}

func Login(c *gin.Context) {
	var input requests.LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create and dispatch the login job synchronously
	loginJob := &authJobs.LoginUserJob{
		Username: input.Email, // Assuming email is used as username
		Password: input.Password,
	}

	result, err := facades.DispatchSync(loginJob)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email or password is incorrect."})
		return
	}

	loginResult := result.(*authJobs.LoginResult)
	c.JSON(http.StatusOK, gin.H{"token": loginResult.Token, "role": loginResult.Role})
}

func CurrentUser(c *gin.Context) {

	userId, err := token.ExtractTokenID(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create and dispatch the get logged in user job synchronously
	getUserJob := &authJobs.GetLoggedInUserJob{
		UserID: userId,
	}

	result, err := facades.DispatchSync(getUserJob)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := result.(*models.User)
	c.JSON(http.StatusOK, gin.H{"message": "success", "data": UserTransformer.Transform(*user)})
}

// ToUserDTO converts a *models.User to a UserDTO
func ToUserDTO(user *models.User) auth.UserDTO {
	return auth.UserDTO{
		ID:           user.ID,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		Email:        user.Email,
		MobileNumber: user.MobileNumber,
	}
}
