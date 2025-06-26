package validators

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var nameRegex = regexp.MustCompile(`^[A-Za-z-' ]+$`)

func NameField(fl validator.FieldLevel) bool {
	return nameRegex.MatchString(fl.Field().String())
}
