package validator_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/ent/validator"
)

func TestValidatePhoneNumber(t *testing.T) {
	// Valid phone numbers
	validNumbers := []string{
		"+44 123 456 7890",
		"+1 (303) 456-7890",
		"+1-717-456-7890",
		"123-456-7890",
		"123.456.7890",
		"(123) 456-7890",
	}

	for _, number := range validNumbers {
		t.Run(fmt.Sprintf("valid phone number: %s", number), func(t *testing.T) {
			valid := validator.ValidatePhoneNumber(number)
			assert.True(t, valid)
		})
	}

	// Invalid phone numbers
	invalidNumbers := []string{
		"123",          // invalid number
		"abc",          // invalid string
		"+",            // invalid string
		"+15558675309", // not a valid US number
		"555-867+5309", // invalid format
	}

	for _, number := range invalidNumbers {
		t.Run(fmt.Sprintf("invalid phone number: %s", number), func(t *testing.T) {
			valid := validator.ValidatePhoneNumber(number)
			assert.False(t, valid)
		})
	}
}
