package objects

import "github.com/go-playground/validator/v10"

// AppValidator is a wrapper around the validator package
type AppValidator struct {
	validator *validator.Validate
}

// Validate runs the validation against the provided struct
func (cv *AppValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// NewValidator creates a new instance of the AppValidator
func NewValidator() *AppValidator {
	return &AppValidator{validator: validator.New()}
}
