package validator

import (
	"github.com/nyaruka/phonenumbers"
)

// ValidatePhoneNumber validates a phone number using the phone numbers library
// this is light validation as it just need to be valid in _any_ region to be considered valid
func ValidatePhoneNumber(phoneNumber string) bool {
	if phoneNumber == "" {
		return true
	}

	// get all the valid regions supported by the library
	validRegions := phonenumbers.GetSupportedRegions()

	for r := range validRegions {
		num, err := phonenumbers.Parse(phoneNumber, r)
		if err == nil {
			// attempt to validate the number in the region
			if phonenumbers.IsValidNumber(num) {
				return true
			}
		}
	}

	return false
}
