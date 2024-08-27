package cors

import "fmt"

// ValidationError is returned when the origins do not pass validation
type ValidationError struct {
	Validation string
	Allowed    string
}

// Error returns the ValidationError in string format
func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Validation, e.Allowed)
}

// newValidationError returns a ValidationError
func newValidationError(v string, a string) *ValidationError {
	return &ValidationError{
		Validation: v,
		Allowed:    a,
	}
}
