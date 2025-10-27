package scim

import "errors"

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
)
