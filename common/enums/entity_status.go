package enums

import "io"

// EntityStatus is a custom type for the status of an entity
type EntityStatus string

var (
	// EntityStatusDraft is when the entity has been added but not yet reviewed
	EntityStatusDraft EntityStatus = "DRAFT"
	// EntityStatusUnderReview is when the entity is undergoing security/legal review
	EntityStatusUnderReview EntityStatus = "UNDER_REVIEW"
	// EntityStatusApproved is when the entity has been cleared for use
	EntityStatusApproved EntityStatus = "APPROVED"
	// EntityStatusRestricted is when the entity is approved with conditions
	EntityStatusRestricted EntityStatus = "RESTRICTED"
	// EntityStatusRejected is when the entity is not allowed
	EntityStatusRejected EntityStatus = "REJECTED"
	// EntityStatusActive is when the entity is currently in use
	EntityStatusActive EntityStatus = "ACTIVE"
	// EntityStatusSuspended is when the entity is temporarily blocked
	EntityStatusSuspended EntityStatus = "SUSPENDED"
	// EntityStatusOffboarding is when the entity is terminating its relationship
	EntityStatusOffboarding EntityStatus = "OFFBOARDING"
	// EntityStatusTerminated is when the entity has been fully offboarded
	EntityStatusTerminated EntityStatus = "TERMINATED"
)

var entityStatusValues = []EntityStatus{
	EntityStatusDraft,
	EntityStatusUnderReview,
	EntityStatusApproved,
	EntityStatusRestricted,
	EntityStatusRejected,
	EntityStatusActive,
	EntityStatusSuspended,
	EntityStatusOffboarding,
	EntityStatusTerminated,
}

// Values returns a slice of strings that represents all the possible values of the EntityStatus enum.
func (EntityStatus) Values() []string { return stringValues(entityStatusValues) }

// String returns the EntityStatus as a string
func (r EntityStatus) String() string { return string(r) }

// ToEntityStatus returns the EntityStatus based on string input
func ToEntityStatus(r string) *EntityStatus {
	return parse(r, entityStatusValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r EntityStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *EntityStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
