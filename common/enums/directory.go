package enums

import "io"

// DirectoryAccountType is the type of principal represented in the directory.
type DirectoryAccountType string

var (
	DirectoryAccountTypeUser    DirectoryAccountType = "USER"
	DirectoryAccountTypeService DirectoryAccountType = "SERVICE"
	DirectoryAccountTypeShared  DirectoryAccountType = "SHARED"
	DirectoryAccountTypeGuest   DirectoryAccountType = "GUEST"
)

var directoryAccountTypeValues = []DirectoryAccountType{
	DirectoryAccountTypeUser, DirectoryAccountTypeService, DirectoryAccountTypeShared, DirectoryAccountTypeGuest,
}

// Values returns all values as strings.
func (DirectoryAccountType) Values() []string { return stringValues(directoryAccountTypeValues) }

// String returns the string value.
func (r DirectoryAccountType) String() string { return string(r) }

// ToDirectoryAccountType converts a string to DirectoryAccountType.
func ToDirectoryAccountType(v string) *DirectoryAccountType {
	return parse(v, directoryAccountTypeValues, nil)
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryAccountType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryAccountType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// DirectoryAccountStatus is the lifecycle status of a directory account.
type DirectoryAccountStatus string

var (
	DirectoryAccountStatusActive    DirectoryAccountStatus = "ACTIVE"
	DirectoryAccountStatusInactive  DirectoryAccountStatus = "INACTIVE"
	DirectoryAccountStatusSuspended DirectoryAccountStatus = "SUSPENDED"
	DirectoryAccountStatusDeleted   DirectoryAccountStatus = "DELETED"
)

var directoryAccountStatusValues = []DirectoryAccountStatus{
	DirectoryAccountStatusActive, DirectoryAccountStatusInactive, DirectoryAccountStatusSuspended, DirectoryAccountStatusDeleted,
}

// Values returns all values as strings.
func (DirectoryAccountStatus) Values() []string { return stringValues(directoryAccountStatusValues) }

// String returns the string value.
func (r DirectoryAccountStatus) String() string { return string(r) }

// ToDirectoryAccountStatus converts a string to DirectoryAccountStatus.
func ToDirectoryAccountStatus(v string) *DirectoryAccountStatus {
	return parse(v, directoryAccountStatusValues, nil)
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryAccountStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryAccountStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// DirectoryAccountMFAState is the MFA state reported by the directory.
type DirectoryAccountMFAState string

var (
	DirectoryAccountMFAStateUnknown  DirectoryAccountMFAState = "UNKNOWN"
	DirectoryAccountMFAStateDisabled DirectoryAccountMFAState = "DISABLED"
	DirectoryAccountMFAStateEnabled  DirectoryAccountMFAState = "ENABLED"
	DirectoryAccountMFAStateEnforced DirectoryAccountMFAState = "ENFORCED"
)

var directoryAccountMFAStateValues = []DirectoryAccountMFAState{
	DirectoryAccountMFAStateUnknown, DirectoryAccountMFAStateDisabled, DirectoryAccountMFAStateEnabled, DirectoryAccountMFAStateEnforced,
}

// Values returns all values as strings.
func (DirectoryAccountMFAState) Values() []string {
	return stringValues(directoryAccountMFAStateValues)
}

// String returns the string value.
func (r DirectoryAccountMFAState) String() string { return string(r) }

// ToDirectoryAccountMFAState converts a string to DirectoryAccountMFAState.
func ToDirectoryAccountMFAState(v string) *DirectoryAccountMFAState {
	return parse(v, directoryAccountMFAStateValues, nil)
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryAccountMFAState) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryAccountMFAState) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// DirectoryGroupClassification is the classification for a directory group.
type DirectoryGroupClassification string

var (
	DirectoryGroupClassificationSecurity     DirectoryGroupClassification = "SECURITY"
	DirectoryGroupClassificationDistribution DirectoryGroupClassification = "DISTRIBUTION"
	DirectoryGroupClassificationTeam         DirectoryGroupClassification = "TEAM"
	DirectoryGroupClassificationDynamic      DirectoryGroupClassification = "DYNAMIC"
)

var directoryGroupClassificationValues = []DirectoryGroupClassification{
	DirectoryGroupClassificationSecurity, DirectoryGroupClassificationDistribution,
	DirectoryGroupClassificationTeam, DirectoryGroupClassificationDynamic,
}

// Values returns all values as strings.
func (DirectoryGroupClassification) Values() []string {
	return stringValues(directoryGroupClassificationValues)
}

// String returns the string value.
func (r DirectoryGroupClassification) String() string { return string(r) }

// ToDirectoryGroupClassification converts a string to DirectoryGroupClassification.
func ToDirectoryGroupClassification(v string) *DirectoryGroupClassification {
	return parse(v, directoryGroupClassificationValues, nil)
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryGroupClassification) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryGroupClassification) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// DirectoryGroupStatus is the lifecycle status of a directory group.
type DirectoryGroupStatus string

var (
	DirectoryGroupStatusActive   DirectoryGroupStatus = "ACTIVE"
	DirectoryGroupStatusInactive DirectoryGroupStatus = "INACTIVE"
	DirectoryGroupStatusDeleted  DirectoryGroupStatus = "DELETED"
)

var directoryGroupStatusValues = []DirectoryGroupStatus{
	DirectoryGroupStatusActive, DirectoryGroupStatusInactive, DirectoryGroupStatusDeleted,
}

// Values returns all values as strings.
func (DirectoryGroupStatus) Values() []string { return stringValues(directoryGroupStatusValues) }

// String returns the string value.
func (r DirectoryGroupStatus) String() string { return string(r) }

// ToDirectoryGroupStatus converts a string to DirectoryGroupStatus.
func ToDirectoryGroupStatus(v string) *DirectoryGroupStatus {
	return parse(v, directoryGroupStatusValues, nil)
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryGroupStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryGroupStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// DirectoryMembershipRole is the membership role reported by the provider.
type DirectoryMembershipRole string

var (
	DirectoryMembershipRoleMember  DirectoryMembershipRole = "MEMBER"
	DirectoryMembershipRoleManager DirectoryMembershipRole = "MANAGER"
	DirectoryMembershipRoleOwner   DirectoryMembershipRole = "OWNER"
)

var directoryMembershipRoleValues = []DirectoryMembershipRole{
	DirectoryMembershipRoleMember, DirectoryMembershipRoleManager, DirectoryMembershipRoleOwner,
}

// Values returns all values as strings.
func (DirectoryMembershipRole) Values() []string {
	return stringValues(directoryMembershipRoleValues)
}

// String returns the string value.
func (r DirectoryMembershipRole) String() string { return string(r) }

// ToDirectoryMembershipRole converts a string to DirectoryMembershipRole.
func ToDirectoryMembershipRole(v string) *DirectoryMembershipRole {
	return parse(v, directoryMembershipRoleValues, nil)
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectoryMembershipRole) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectoryMembershipRole) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

// DirectorySyncRunStatus is the state of a directory sync run.
type DirectorySyncRunStatus string

var (
	DirectorySyncRunStatusPending   DirectorySyncRunStatus = "PENDING"
	DirectorySyncRunStatusRunning   DirectorySyncRunStatus = "RUNNING"
	DirectorySyncRunStatusCompleted DirectorySyncRunStatus = "COMPLETED"
	DirectorySyncRunStatusFailed    DirectorySyncRunStatus = "FAILED"
)

var directorySyncRunStatusValues = []DirectorySyncRunStatus{
	DirectorySyncRunStatusPending, DirectorySyncRunStatusRunning, DirectorySyncRunStatusCompleted, DirectorySyncRunStatusFailed,
}

// Values returns all values as strings.
func (DirectorySyncRunStatus) Values() []string { return stringValues(directorySyncRunStatusValues) }

// String returns the string value.
func (r DirectorySyncRunStatus) String() string { return string(r) }

// ToDirectorySyncRunStatus converts a string to DirectorySyncRunStatus.
func ToDirectorySyncRunStatus(v string) *DirectorySyncRunStatus {
	return parse(v, directorySyncRunStatusValues, nil)
}

// MarshalGQL implements gqlgen Marshaler.
func (r DirectorySyncRunStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements gqlgen Unmarshaler.
func (r *DirectorySyncRunStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
