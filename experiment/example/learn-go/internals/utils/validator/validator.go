package validator

import (
	valid "github.com/go-playground/validator/v10"
)

type customValidator struct {
	validator *valid.Validate
}

func NewCustomValidator() *customValidator {
	newValidator := valid.New()

	// add custom validator
	newValidator.RegisterValidation("is-date", validateIsDate)
	newValidator.RegisterValidation("phone", validatePhoneNumber)
	newValidator.RegisterValidation("password", validatePassword)

	return &customValidator{
		validator: newValidator,
	}
}

func (cv *customValidator) Validate(i any) error {
	err := cv.validator.Struct(i)
	if err != nil {
		object, _ := err.(valid.ValidationErrors)
		return object
	}
	return cv.validator.Struct(i)
}
