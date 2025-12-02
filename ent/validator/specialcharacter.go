package validator

import (
	"fmt"
	"strings"

	"github.com/theopenlane/utils/rout"
)

// InvalidChars is a list of characters not allowed to be used
// as part of a name/title
const InvalidChars = "!@#$%^*()+{}|:\"<>?`=[]\\;'/~"

// SpecialCharValidator validates a name field for special characters
// returns an error if the name contains special characters.
// Only hyphens, underscores, periods, commas, and ampersands are allowed
func SpecialCharValidator(name string) error {

	if strings.ContainsAny(name, InvalidChars) {
		return fmt.Errorf("%w, field cannot contain special characters", rout.InvalidField("name"))
	}

	return nil
}
