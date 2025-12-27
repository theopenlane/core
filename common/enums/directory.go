package enums

import (
	"fmt"
	"io"
	"strings"
)

// DirectoryAccountType is the type of principal represented in the directory.
type DirectoryAccountType string

var (
	DirectoryAccountTypeUser    DirectoryAccountType = "USER"
	DirectoryAccountTypeService DirectoryAccountType = "SERVICE"
	DirectoryAccountTypeShared  DirectoryAccountType = "SHARED"
	DirectoryAccountTypeGuest   DirectoryAccountType = "GUEST"
)

// DirectoryAccountTypes lists all account types.
var DirectoryAccountTypes = []string{
	string(DirectoryAccountTypeUser),
	string(DirectoryAccountTypeService),
	string(DirectoryAccountTypeShared),
	string(DirectoryAccountTypeGuest),
}

// Values returns all values as strings.
func (DirectoryAccountType) Values() []string { return DirectoryAccountTypes }

// String returns the string value.
func (r DirectoryAccountType) String() string { return string(r) }

// ToDirectoryAccountType converts a string to DirectoryAccountType.
func ToDirectoryAccountType(v string) *DirectoryAccountType {
	switch strings.ToUpper(v) {
	case DirectoryAccountTypeUser.String():
		return &DirectoryAccountTypeUser
	case DirectoryAccountTypeService.String():
		return &DirectoryAccountTypeService
	case DirectoryAccountTypeShared.String():
		return &DirectoryAccountTypeShared
	case DirectoryAccountTypeGuest.String():
		return &DirectoryAccountTypeGuest
	default:
		return nil
	}
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryAccountType) MarshalGQL(w io.Writer) { _, _ = w.Write([]byte(`"` + r.String() + `"`)) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryAccountType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeDirectoryAccountType, v)
	}
	*r = DirectoryAccountType(str)
	return nil
}

// DirectoryAccountStatus is the lifecycle status of a directory account.
type DirectoryAccountStatus string

var (
	DirectoryAccountStatusActive    DirectoryAccountStatus = "ACTIVE"
	DirectoryAccountStatusInactive  DirectoryAccountStatus = "INACTIVE"
	DirectoryAccountStatusSuspended DirectoryAccountStatus = "SUSPENDED"
	DirectoryAccountStatusDeleted   DirectoryAccountStatus = "DELETED"
)

// DirectoryAccountStatuses lists all account statuses.
var DirectoryAccountStatuses = []string{
	string(DirectoryAccountStatusActive),
	string(DirectoryAccountStatusInactive),
	string(DirectoryAccountStatusSuspended),
	string(DirectoryAccountStatusDeleted),
}

// Values returns all values as strings.
func (DirectoryAccountStatus) Values() []string { return DirectoryAccountStatuses }

// String returns the string value.
func (r DirectoryAccountStatus) String() string { return string(r) }

// ToDirectoryAccountStatus converts a string to DirectoryAccountStatus.
func ToDirectoryAccountStatus(v string) *DirectoryAccountStatus {
	switch strings.ToUpper(v) {
	case DirectoryAccountStatusActive.String():
		return &DirectoryAccountStatusActive
	case DirectoryAccountStatusInactive.String():
		return &DirectoryAccountStatusInactive
	case DirectoryAccountStatusSuspended.String():
		return &DirectoryAccountStatusSuspended
	case DirectoryAccountStatusDeleted.String():
		return &DirectoryAccountStatusDeleted
	default:
		return nil
	}
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryAccountStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryAccountStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeDirectoryAccountStatus, v)
	}
	*r = DirectoryAccountStatus(str)
	return nil
}

// DirectoryAccountMFAState is the MFA state reported by the directory.
type DirectoryAccountMFAState string

var (
	DirectoryAccountMFAStateUnknown  DirectoryAccountMFAState = "UNKNOWN"
	DirectoryAccountMFAStateDisabled DirectoryAccountMFAState = "DISABLED"
	DirectoryAccountMFAStateEnabled  DirectoryAccountMFAState = "ENABLED"
	DirectoryAccountMFAStateEnforced DirectoryAccountMFAState = "ENFORCED"
)

// DirectoryAccountMFAStates lists all MFA states.
var DirectoryAccountMFAStates = []string{
	string(DirectoryAccountMFAStateUnknown),
	string(DirectoryAccountMFAStateDisabled),
	string(DirectoryAccountMFAStateEnabled),
	string(DirectoryAccountMFAStateEnforced),
}

// Values returns all values as strings.
func (DirectoryAccountMFAState) Values() []string { return DirectoryAccountMFAStates }

// String returns the string value.
func (r DirectoryAccountMFAState) String() string { return string(r) }

// ToDirectoryAccountMFAState converts a string to DirectoryAccountMFAState.
func ToDirectoryAccountMFAState(v string) *DirectoryAccountMFAState {
	switch strings.ToUpper(v) {
	case DirectoryAccountMFAStateUnknown.String():
		return &DirectoryAccountMFAStateUnknown
	case DirectoryAccountMFAStateDisabled.String():
		return &DirectoryAccountMFAStateDisabled
	case DirectoryAccountMFAStateEnabled.String():
		return &DirectoryAccountMFAStateEnabled
	case DirectoryAccountMFAStateEnforced.String():
		return &DirectoryAccountMFAStateEnforced
	default:
		return nil
	}
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryAccountMFAState) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryAccountMFAState) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeDirectoryAccountMFAState, v)
	}
	*r = DirectoryAccountMFAState(str)
	return nil
}

// DirectoryGroupClassification is the classification for a directory group.
type DirectoryGroupClassification string

var (
	DirectoryGroupClassificationSecurity     DirectoryGroupClassification = "SECURITY"
	DirectoryGroupClassificationDistribution DirectoryGroupClassification = "DISTRIBUTION"
	DirectoryGroupClassificationTeam         DirectoryGroupClassification = "TEAM"
	DirectoryGroupClassificationDynamic      DirectoryGroupClassification = "DYNAMIC"
)

// DirectoryGroupClassifications lists all group classifications.
var DirectoryGroupClassifications = []string{
	string(DirectoryGroupClassificationSecurity),
	string(DirectoryGroupClassificationDistribution),
	string(DirectoryGroupClassificationTeam),
	string(DirectoryGroupClassificationDynamic),
}

// Values returns all values as strings.
func (DirectoryGroupClassification) Values() []string { return DirectoryGroupClassifications }

// String returns the string value.
func (r DirectoryGroupClassification) String() string { return string(r) }

// ToDirectoryGroupClassification converts a string to DirectoryGroupClassification.
func ToDirectoryGroupClassification(v string) *DirectoryGroupClassification {
	switch strings.ToUpper(v) {
	case DirectoryGroupClassificationSecurity.String():
		return &DirectoryGroupClassificationSecurity
	case DirectoryGroupClassificationDistribution.String():
		return &DirectoryGroupClassificationDistribution
	case DirectoryGroupClassificationTeam.String():
		return &DirectoryGroupClassificationTeam
	case DirectoryGroupClassificationDynamic.String():
		return &DirectoryGroupClassificationDynamic
	default:
		return nil
	}
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryGroupClassification) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryGroupClassification) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeDirectoryGroupClassification, v)
	}
	*r = DirectoryGroupClassification(str)
	return nil
}

// DirectoryGroupStatus is the lifecycle status of a directory group.
type DirectoryGroupStatus string

var (
	DirectoryGroupStatusActive   DirectoryGroupStatus = "ACTIVE"
	DirectoryGroupStatusInactive DirectoryGroupStatus = "INACTIVE"
	DirectoryGroupStatusDeleted  DirectoryGroupStatus = "DELETED"
)

// DirectoryGroupStatuses lists all group statuses.
var DirectoryGroupStatuses = []string{
	string(DirectoryGroupStatusActive),
	string(DirectoryGroupStatusInactive),
	string(DirectoryGroupStatusDeleted),
}

// Values returns all values as strings.
func (DirectoryGroupStatus) Values() []string { return DirectoryGroupStatuses }

// String returns the string value.
func (r DirectoryGroupStatus) String() string { return string(r) }

// ToDirectoryGroupStatus converts a string to DirectoryGroupStatus.
func ToDirectoryGroupStatus(v string) *DirectoryGroupStatus {
	switch strings.ToUpper(v) {
	case DirectoryGroupStatusActive.String():
		return &DirectoryGroupStatusActive
	case DirectoryGroupStatusInactive.String():
		return &DirectoryGroupStatusInactive
	case DirectoryGroupStatusDeleted.String():
		return &DirectoryGroupStatusDeleted
	default:
		return nil
	}
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryGroupStatus) MarshalGQL(w io.Writer) { _, _ = w.Write([]byte(`"` + r.String() + `"`)) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryGroupStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeDirectoryGroupStatus, v)
	}
	*r = DirectoryGroupStatus(str)
	return nil
}

// DirectoryMembershipRole is the membership role reported by the provider.
type DirectoryMembershipRole string

var (
	DirectoryMembershipRoleMember  DirectoryMembershipRole = "MEMBER"
	DirectoryMembershipRoleManager DirectoryMembershipRole = "MANAGER"
	DirectoryMembershipRoleOwner   DirectoryMembershipRole = "OWNER"
)

// DirectoryMembershipRoles lists all membership roles.
var DirectoryMembershipRoles = []string{
	string(DirectoryMembershipRoleMember),
	string(DirectoryMembershipRoleManager),
	string(DirectoryMembershipRoleOwner),
}

// Values returns all values as strings.
func (DirectoryMembershipRole) Values() []string { return DirectoryMembershipRoles }

// String returns the string value.
func (r DirectoryMembershipRole) String() string { return string(r) }

// ToDirectoryMembershipRole converts a string to DirectoryMembershipRole.
func ToDirectoryMembershipRole(v string) *DirectoryMembershipRole {
	switch strings.ToUpper(v) {
	case DirectoryMembershipRoleMember.String():
		return &DirectoryMembershipRoleMember
	case DirectoryMembershipRoleManager.String():
		return &DirectoryMembershipRoleManager
	case DirectoryMembershipRoleOwner.String():
		return &DirectoryMembershipRoleOwner
	default:
		return nil
	}
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryMembershipRole) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryMembershipRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeDirectoryMembershipRole, v)
	}
	*r = DirectoryMembershipRole(str)
	return nil
}

// DirectorySyncRunStatus is the state of a directory sync run.
type DirectorySyncRunStatus string

var (
	DirectorySyncRunStatusPending   DirectorySyncRunStatus = "PENDING"
	DirectorySyncRunStatusRunning   DirectorySyncRunStatus = "RUNNING"
	DirectorySyncRunStatusCompleted DirectorySyncRunStatus = "COMPLETED"
	DirectorySyncRunStatusFailed    DirectorySyncRunStatus = "FAILED"
)

// DirectorySyncRunStatuses lists all sync run statuses.
var DirectorySyncRunStatuses = []string{
	string(DirectorySyncRunStatusPending),
	string(DirectorySyncRunStatusRunning),
	string(DirectorySyncRunStatusCompleted),
	string(DirectorySyncRunStatusFailed),
}

// Values returns all values as strings.
func (DirectorySyncRunStatus) Values() []string { return DirectorySyncRunStatuses }

// String returns the string value.
func (r DirectorySyncRunStatus) String() string { return string(r) }

// ToDirectorySyncRunStatus converts a string to DirectorySyncRunStatus.
func ToDirectorySyncRunStatus(v string) *DirectorySyncRunStatus {
	switch strings.ToUpper(v) {
	case DirectorySyncRunStatusPending.String():
		return &DirectorySyncRunStatusPending
	case DirectorySyncRunStatusRunning.String():
		return &DirectorySyncRunStatusRunning
	case DirectorySyncRunStatusCompleted.String():
		return &DirectorySyncRunStatusCompleted
	case DirectorySyncRunStatusFailed.String():
		return &DirectorySyncRunStatusFailed
	default:
		return nil
	}
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectorySyncRunStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectorySyncRunStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("%w, got: %T", ErrWrongTypeDirectorySyncRunStatus, v)
	}
	*r = DirectorySyncRunStatus(str)
	return nil
}
