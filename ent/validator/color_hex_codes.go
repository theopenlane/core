package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/theopenlane/utils/rout"
)

// HexColorValidator validates a color field for hex color codes
func HexColorValidator(color string) error {
	// Check if it starts with #
	if !strings.HasPrefix(color, "#") {
		return fmt.Errorf("%w, field is not a valid hex color code", rout.InvalidField("color"))
	}

	// Remove the # prefix
	hex := color[1:]

	// Check length - must be 3 or 6 characters
	if len(hex) != 3 && len(hex) != 6 {
		return fmt.Errorf("%w, field is not a valid hex color code", rout.InvalidField("color"))
	}

	// Check if all characters are valid hex digits (0-9, A-F, a-f)
	hexPattern := regexp.MustCompile(`^[0-9A-Fa-f]+$`)
	if !hexPattern.MatchString(hex) {
		return fmt.Errorf("%w, field is not a valid hex color code", rout.InvalidField("color"))
	}

	return nil
}
