package enums

import "io"

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

var evidenceStatusValues = []EvidenceStatus{
	EvidenceStatusSubmitted,
	EvidenceStatusReadyForAuditor,
	EvidenceStatusAuditorApproved,
	EvidenceStatusInReview,
	EvidenceStatusMissingArtifact,
	EvidenceStatusNeedsRenewal,
	EvidenceStatusRejected,
}

// Values returns a slice of strings representing all valid EvidenceStatus values.
func (EvidenceStatus) Values() []string { return stringValues(evidenceStatusValues) }

// String returns the string representation of the EvidenceStatus value.
func (r EvidenceStatus) String() string { return string(r) }

// ToEvidenceStatus converts a string to its corresponding EvidenceStatus enum value.
func ToEvidenceStatus(r string) *EvidenceStatus {
	return parse(r, evidenceStatusValues, &EvidenceStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r EvidenceStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *EvidenceStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
