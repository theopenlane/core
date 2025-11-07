//go:build cli

package speccli

import cmdpkg "github.com/theopenlane/core/cmd/cli/cmd"

// RequiredFieldMissing creates the standard CLI error for a missing required field.
func RequiredFieldMissing(field string) error {
	return cmdpkg.NewRequiredFieldMissingError(field)
}

// InvalidField wraps the shared CLI invalid-field error helper.
func InvalidField(field, value string) error {
	return cmdpkg.NewInvalidFieldError(field, value)
}
