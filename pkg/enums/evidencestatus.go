package enums

import (
	"fmt"
	"io"
	"strings"
)

// EvidenceStatus is a custom type representing the various states of EvidenceStatus.
type EvidenceStatus string

var (
	// EvidenceStatusSubmitted indicates the submitted.
	EvidenceStatusSubmitted EvidenceStatus = "SUBMITTED"
	// EvidenceStatusReadyForAuditor indicates the ready for auditor.
	EvidenceStatusReadyForAuditor EvidenceStatus = "READY_FOR_AUDITOR"
	// EvidenceStatusAuditorApproved indicates the auditor approved.
	EvidenceStatusAuditorApproved EvidenceStatus = "AUDITOR_APPROVED"
	// EvidenceStatusInReview indicates the in review.
	EvidenceStatusInReview EvidenceStatus = "IN_REVIEW"
	// EvidenceStatusMissingArtifact indicates the missing artifact.
	EvidenceStatusMissingArtifact EvidenceStatus = "MISSING_ARTIFACT"
	// EvidenceStatusNeedsRenewal indicates the needs renewal.
	EvidenceStatusNeedsRenewal EvidenceStatus = "NEEDS_RENEWAL"
	// EvidenceStatusRejected indicates the rejected.
	EvidenceStatusRejected EvidenceStatus = "REJECTED"
	// EvidenceStatusInvalid is used when an unknown or unsupported value is provided.
	EvidenceStatusInvalid EvidenceStatus = "EVIDENCESTATUS_INVALID"
)

// Values returns a slice of strings representing all valid EvidenceStatus values.
func (EvidenceStatus) Values() []string {
	return []string{
		string(EvidenceStatusSubmitted),
		string(EvidenceStatusReadyForAuditor),
		string(EvidenceStatusAuditorApproved),
		string(EvidenceStatusInReview),
		string(EvidenceStatusMissingArtifact),
		string(EvidenceStatusNeedsRenewal),
		string(EvidenceStatusRejected),
	}
}

// String returns the string representation of the EvidenceStatus value.
func (r EvidenceStatus) String() string {
	return string(r)
}

// ToEvidenceStatus converts a string to its corresponding EvidenceStatus enum value.
func ToEvidenceStatus(r string) *EvidenceStatus {
	switch strings.ToUpper(r) {
	case EvidenceStatusSubmitted.String():
		return &EvidenceStatusSubmitted
	case EvidenceStatusReadyForAuditor.String():
		return &EvidenceStatusReadyForAuditor
	case EvidenceStatusAuditorApproved.String():
		return &EvidenceStatusAuditorApproved
	case EvidenceStatusInReview.String():
		return &EvidenceStatusInReview
	case EvidenceStatusMissingArtifact.String():
		return &EvidenceStatusMissingArtifact
	case EvidenceStatusNeedsRenewal.String():
		return &EvidenceStatusNeedsRenewal
	case EvidenceStatusRejected.String():
		return &EvidenceStatusRejected
	default:
		return &EvidenceStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r EvidenceStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *EvidenceStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for EvidenceStatus, got: %T", v) //nolint:err113
	}

	*r = EvidenceStatus(str)

	return nil
}
