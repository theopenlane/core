package hooks

import (
	"errors"
	"strings"
)

var (
	// ErrInternalServerError is returned when an internal error occurs.
	ErrInternalServerError = errors.New("internal server error")
	// ErrInvalidInput is returned when the input is invalid.
	ErrInvalidInput = errors.New("invalid input")
	// ErrPersonalOrgsNoChildren is returned when personal org attempts to add a child org
	ErrPersonalOrgsNoChildren = errors.New("personal organizations are not allowed to have child organizations")
	// ErrPersonalOrgsNoMembers is returned when personal org attempts to add members
	ErrPersonalOrgsNoMembers = errors.New("personal organizations are not allowed to have members other than the owner")
	// ErrOrgOwnerCannotBeDeleted is returned when an org owner is attempted to be deleted
	ErrOrgOwnerCannotBeDeleted = errors.New("organization owner cannot be deleted, it must be transferred to a new owner first")
	// ErrPersonalOrgsNoUser is returned when personal org has no user associated, so no permissions can be added
	ErrPersonalOrgsNoUser = errors.New("personal organizations missing user association")
	// ErrUserNotInOrg is returned when a user is not a member of an organization when trying to add them to a group
	ErrUserNotInOrg = errors.New("user not in organization")
	// ErrUnsupportedFGARole is returned when a role is assigned that is not supported in our fine grained authorization system
	ErrUnsupportedFGARole = errors.New("unsupported role")
	// ErrMissingRole is returned when an update request is made that contains no role
	ErrMissingRole = errors.New("missing role in update")
	// ErrUserAlreadyOrgMember is returned when an user attempts to be invited to an org they are already a member of
	ErrUserAlreadyOrgMember = errors.New("user already member of organization")
	// ErrUserAlreadySubscriber is returned when an user attempts to subscribe to an organization but is already a subscriber
	ErrUserAlreadySubscriber = errors.New("user already a subscriber")
	// ErrEmailRequired is returned when an email is required but not provided
	ErrEmailRequired = errors.New("email is required but not provided")
	// ErrMaxAttempts is returned when a user has reached the max attempts to resend an invitation to an org
	ErrMaxAttempts = errors.New("too many attempts to resend org invitation")
	// ErrMissingRecipientEmail is returned when an email is required but not provided
	ErrMissingRecipientEmail = errors.New("recipient email is required but not provided")
	// ErrMissingRequiredName is returned when a name is required but not provided
	ErrMissingRequiredName = errors.New("name or display name is required but not provided")
	// ErrTooManyAvatarFiles is returned when a user attempts to upload more than one avatar file
	ErrTooManyAvatarFiles = errors.New("too many avatar files uploaded, only one is allowed")
	// ErrFailedToRegisterListener is returned when a listener fails to register
	ErrFailedToRegisterListener = errors.New("failed to register listener")
	// ErrNoControls is returned when a subcontrol has no controls assigned
	ErrNoControls = errors.New("subcontrol must have at least one control assigned")
	// ErrUnableToCast is returned when a type assertion fails
	ErrUnableToCast = errors.New("unable to cast")
	// ErrTooManySubscriptions is returned when an organization has too many subscriptions
	ErrTooManySubscriptions = errors.New("organization has too many subscriptions")
	// ErrManagedGroup is returned when a user attempts to modify a managed group
	ErrManagedGroup = errors.New("managed groups cannot be modified")
	// ErrMaxAttemptsOrganization is returned when the max attempts have been reached to create an organization via onboarding
	ErrMaxAttemptsOrganization = errors.New("too many attempts to create organization")
	// ErrEmailDomainNotAllowed is returned when an email domain is not allowed to be used for an organization
	ErrEmailDomainNotAllowed = errors.New("email domain not allowed in organization")
	// ErrUserNotFound is returned when a user is not found in the system
	ErrUserNotFound = errors.New("user not found")
	// ErrPersonalOrgNoSubscription is returned when the org is a personal org and we don't want to create a subscription in stripe
	ErrPersonalOrgNoSubscription = errors.New("detected personal organization, do not create stripe subscription")
)

// IsUniqueConstraintError reports if the error resulted from a DB uniqueness constraint violation.
// e.g. duplicate value in unique index.
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	for _, s := range []string{
		"Error 1062",                 // MySQL
		"violates unique constraint", // Postgres
		"UNIQUE constraint failed",   // SQLite
	} {
		if strings.Contains(err.Error(), s) {
			return true
		}
	}

	return false
}
