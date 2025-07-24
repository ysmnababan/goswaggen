package validator

import (
	"strings"
	"time"
	"unicode"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// validatePhoneNumber is the custom validation function for the "phone" tag
func validatePhoneNumber(fl validator.FieldLevel) bool {
	phone, ok := fl.Field().Interface().(string)
	if !ok {
		return false // Only strings are valid
	}

	// Check if the phone number is empty or does not start with "628"
	log.Info().Str("phone", phone).Send()

	if phone == "" || !strings.HasPrefix(phone, "62") {
		return false
	}

	if len(phone) > 13 {
		return false
	}

	return true
}

func validatePassword(fl validator.FieldLevel) bool {
	password, ok := fl.Field().Interface().(string)
	if !ok {
		return false // Only strings are valid
	}

	var hasLetter, hasNumber bool

	for _, char := range password {
		if unicode.IsLetter(char) {
			hasLetter = true
		}
		if unicode.IsNumber(char) {
			hasNumber = true
		}
		// Early return if both conditions are met
		if hasLetter && hasNumber {
			return true
		}
	}

	return hasLetter && hasNumber
}

func validateIsDate(fl validator.FieldLevel) bool {
	field := fl.Field().String()
	if field == "" {
		return true
	}
	if _, err := time.Parse("2006-01-02", field); err != nil {
		return false
	}

	return true
}
