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
	ErrUserAlreadySubscriber = errors.New("subscriber already exists")
	// ErrEmailRequired is returned when an email is required but not provided
	ErrEmailRequired = errors.New("email is required but not provided")
	// ErrMaxAttempts is returned when a user has reached the max attempts to resend an invitation to an org
	ErrMaxAttempts = errors.New("too many attempts to resend org invitation")
	// ErrMaxSubscriptionAttempts is returned when a user has reached the max attempts to subscribe to an org
	ErrMaxSubscriptionAttempts = errors.New("too many attempts to resend org subscription email")
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
	// ErrNoSubscriptions is returned when an organization has no subscriptions
	ErrNoSubscriptions = errors.New("organization has no subscriptions")
	// ErrTooManySubscriptions is returned when an organization has too many subscriptions
	ErrTooManySubscriptions = errors.New("organization has too many subscriptions")
	// ErrTooManyPrices is returned when an organization has too many subscriptions
	ErrTooManyPrices = errors.New("organization has too many prices on a subscription")
	// ErrNoPrices is returned when a subscription has no price
	ErrNoPrices = errors.New("subscription has no price")
	// ErrManagedGroup is returned when a user attempts to modify a managed group
	ErrManagedGroup = errors.New("managed groups cannot be modified")
	// ErrMaxAttemptsOrganization is returned when the max attempts have been reached to create an organization via onboarding
	ErrMaxAttemptsOrganization = errors.New("too many attempts to create organization")
	// ErrEmailDomainNotAllowed is returned when an email domain is not allowed to be used for an organization
	ErrEmailDomainNotAllowed = errors.New("email domain not allowed in organization")
	// ErrUserNotFound is returned when a user is not found in the system
	ErrUserNotFound = errors.New("user not found")
	// ErrCadenceOrCronRequired is returned when a user does not provide either a cadence or cron
	ErrCadenceOrCronRequired = errors.New("either cadence or cron must be specified")
	// ErrEitherCadenceOrCron is returned when both a cadence and cron is specified
	ErrEitherCadenceOrCron = errors.New("only one of cadence or cron must be specified")
	// ErrZeroTimeNotAllowed is returned when you try to set a non usable time value
	ErrZeroTimeNotAllowed = errors.New("time cannot be empty. Provide a valid time/date")
	// ErrFutureTimeNotAllowed is returned when you try to set a time into the future.
	// future being any second/minute past the current time of validation
	ErrFutureTimeNotAllowed = errors.New("time cannot be in the future")
	// ErrMissingFileID is returned when a file ID is required but not provided
	ErrMissingFileID = errors.New("missing file id")

	ErrCannotCreateNegativeUsage = errors.New("cannot create usage record with negative delta")

	ErrFailedToQueryUsage = errors.New("failed to query usage record")
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
