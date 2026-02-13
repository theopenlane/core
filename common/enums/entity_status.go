package enums

import (
	"fmt"
	"io"
	"strings"
)

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

func (EntityStatus) Values() (kinds []string) {
	for _, s := range []EntityStatus{
		EntityStatusDraft,
		EntityStatusUnderReview,
		EntityStatusApproved,
		EntityStatusRestricted,
		EntityStatusRejected,
		EntityStatusActive,
		EntityStatusSuspended,
		EntityStatusOffboarding,
		EntityStatusTerminated,
	} {
		kinds = append(kinds, string(s))
	}

	return
}

func (r EntityStatus) String() string {
	return string(r)
}

func ToEntityStatus(r string) *EntityStatus {
	switch r := strings.ToUpper(r); r {
	case EntityStatusDraft.String():
		return &EntityStatusDraft
	case EntityStatusUnderReview.String():
		return &EntityStatusUnderReview
	case EntityStatusApproved.String():
		return &EntityStatusApproved
	case EntityStatusRestricted.String():
		return &EntityStatusRestricted
	case EntityStatusRejected.String():
		return &EntityStatusRejected
	case EntityStatusActive.String():
		return &EntityStatusActive
	case EntityStatusSuspended.String():
		return &EntityStatusSuspended
	case EntityStatusOffboarding.String():
		return &EntityStatusOffboarding
	case EntityStatusTerminated.String():
		return &EntityStatusTerminated
	default:
		return nil
	}
}

func (r EntityStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

func (r *EntityStatus) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for EntityStatus, got: %T", v)
	}

	*r = EntityStatus(str)

	return nil
}
