package validators

// NameValidator validates name fields
type NameValidator struct{}

// Validate validates a name field
func (v *NameValidator) Validate(value string) error {
	// TODO: Implement name validation
	// - Check for minimum length
	// - Check for maximum length
	// - Check for valid characters
	// - Check for profanity
	return nil
}

// NameFieldValidator provides Laravel-style validation for name fields
func NameFieldValidator() *NameValidator {
	return &NameValidator{}
}
