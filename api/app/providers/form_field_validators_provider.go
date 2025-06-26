package providers

import (
	"base_lara_go_project/app/validators"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

func RegisterFormFieldValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("nameField", validators.NameField)

		// Add more custom validators here as you create them:
	}
}
