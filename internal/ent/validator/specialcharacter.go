package validator

import (
	"fmt"
	"strings"

	"github.com/theopenlane/utils/rout"
)

// SpecialCharValidator validates a name field for special characters
// returns an error if the name contains special characters.
// Only hyphens, underscores, periods, commas, and ampersands are allowed
func SpecialCharValidator(name string) error {
	invalidChars := "!@#$%^*()+{}|:\"<>?`=[]\\;'/~"

	if strings.ContainsAny(name, invalidChars) {
		return fmt.Errorf("%w, field cannot contain special characters", rout.InvalidField("name"))
	}

	return nil
}
