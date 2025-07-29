package enums

import (
	"fmt"
	"io"
	"strings"
)

// EvidenceStatus is a custom type for evidence status
type EvidenceStatus string

var (
	// EvidenceSubmitted is the status to indicate that the evidence has been submitted and is ready for internal review
	EvidenceSubmitted EvidenceStatus = "SUBMITTED"
	// EvidenceInReview is the status to indicate that the evidence is currently under internal review
	EvidenceInReview EvidenceStatus = "IN_REVIEW"
	// EvidenceReady is the status to indicate that the evidence is ready for auditor review
	EvidenceReady EvidenceStatus = "READY"
	// EvidenceApproved is the status to indicate that the evidence has been approved by the auditor
	EvidenceApproved EvidenceStatus = "APPROVED"
	// EvidenceMissingArtifact is the status to indicate that the evidence is missing an artifact
	EvidenceMissingArtifact EvidenceStatus = "MISSING_ARTIFACT"
	// EvidenceNeedsRenewal is the status to indicate that the evidence needs to be renewed
	EvidenceNeedsRenewal EvidenceStatus = "NEEDS_RENEWAL"
	// EvidenceRejected is the status to indicate that the evidence has been rejected by the auditor
	EvidenceRejected EvidenceStatus = "REJECTED"
	// EvidenceInvalid is the status to indicate that the evidence is invalid
	EvidenceInvalid EvidenceStatus = "EVIDENCE_STATUS_INVALID"
)

// Values returns a slice of strings that represents all the possible values of the EvidenceStatus enum.
// Possible default values are "READY", "APPROVED", "MISSING_ARTIFACT", "REJECTED", and "NEEDS_RENEWAL"
func (EvidenceStatus) Values() (kinds []string) {
	for _, s := range []EvidenceStatus{EvidenceApproved, EvidenceReady, EvidenceMissingArtifact, EvidenceRejected, EvidenceNeedsRenewal, EvidenceSubmitted, EvidenceInReview} {
		kinds = append(kinds, string(s))
	}

	return
}

// String returns the evidence status as a string
func (r EvidenceStatus) String() string {
	return string(r)
}

// ToEvidenceStatus returns the evidence status enum based on string input
func ToEvidenceStatus(r string) *EvidenceStatus {
	switch r := strings.ToUpper(r); r {
	case EvidenceReady.String():
		return &EvidenceReady
	case EvidenceApproved.String():
		return &EvidenceApproved
	case EvidenceMissingArtifact.String():
		return &EvidenceMissingArtifact
	case EvidenceRejected.String():
		return &EvidenceRejected
	case EvidenceNeedsRenewal.String():
		return &EvidenceNeedsRenewal
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r EvidenceStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *EvidenceStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for EvidenceStatus, got: %T", v) //nolint:err113
	}

	*r = EvidenceStatus(str)

	return nil
}
