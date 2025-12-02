package scim

import (
	"errors"
	"fmt"
	"net/http"

	scimerrors "github.com/elimity-com/scim/errors"

	"github.com/theopenlane/ent/generated"
)

var (
	// ErrUserNotFound is returned when a user is not found.
	ErrUserNotFound = errors.New("user not found")
	// ErrGroupNotFound is returned when a group is not found.
	ErrGroupNotFound = errors.New("group not found")
	// ErrInvalidAttributes is returned when resource attributes are invalid.
	ErrInvalidAttributes = errors.New("invalid resource attributes")
	// ErrOrgNotFound is returned when organization context is missing.
	ErrOrgNotFound = errors.New("organization not found in context")
	// ErrUserNotMemberOfOrg is returned when a user is not a member of the organization.
	ErrUserNotMemberOfOrg = errors.New("user is not a member of organization")
	// ErrSSONotEnforced is returned when SCIM operations are attempted but SSO is not enforced for the organization.
	ErrSSONotEnforced = errors.New("SSO must be enforced for the organization to use SCIM provisioning")
	// ErrOrgSettingsNotFound is returned when organization settings are not found.
	ErrOrgSettingsNotFound = errors.New("organization settings not found")
)

// HandleEntError converts ent database errors to SCIM-compliant error responses
// It maps constraint errors to uniqueness violations and validation errors to invalid value errors
func HandleEntError(err error, operation string, detail string) error {
	if generated.IsConstraintError(err) {
		return scimerrors.ScimError{
			ScimType: scimerrors.ScimTypeUniqueness,
			Detail:   detail,
			Status:   http.StatusConflict,
		}
	}

	if generated.IsValidationError(err) {
		return scimerrors.ScimError{
			ScimType: scimerrors.ScimTypeInvalidValue,
			Detail:   fmt.Sprintf("Invalid attributes: %v", err),
			Status:   http.StatusBadRequest,
		}
	}

	return fmt.Errorf("%s: %w", operation, err)
}
