package requests

type RegisterRequest struct {
	Password             string `json:"password" binding:"required,min=8,max=64"`
	PasswordConfirmation string `json:"password_confirmation" binding:"required,eqfield=Password"`
	FirstName            string `json:"first_name" binding:"required,nameField"`
    LastName             string `json:"last_name" binding:"required,nameField"`
	Email                string `json:"email" binding:"required,email"`
	MobileNumber         string `json:"mobile_number" binding:"required,e164"`
}
