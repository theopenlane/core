package objects

import "github.com/go-playground/validator/v10"

type AppValidator struct {
	validator *validator.Validate
}

func (cv *AppValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func NewValidator() *AppValidator {
	return &AppValidator{validator: validator.New()}
}
