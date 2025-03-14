package enums

import (
	"fmt"
	"io"
	"strings"
)

type EvidenceStatus string

var (
	EvidenceReady           EvidenceStatus = "READY"
	EvidenceApproved        EvidenceStatus = "APPROVED"
	EvidenceMissingArtifact EvidenceStatus = "MISSING_ARTIFACT"
	EvidenceNeedsRenewal    EvidenceStatus = "NEEDS_RENEWAL"
	EvidenceRejected        EvidenceStatus = "REJECTED"
	EvidenceInvalid         EvidenceStatus = "EVIDENCE_STATUS_INVALID"
)

// Values returns a slice of strings that represents all the possible values of the EvidenceStatus enum.
// Possible default values are "READY", "APPROVED", "MISSING_ARTIFACT", "REJECTED", and "NEEDS_RENEWAL"
func (EvidenceStatus) Values() (kinds []string) {
	for _, s := range []EvidenceStatus{EvidenceApproved, EvidenceReady, EvidenceMissingArtifact, EvidenceRejected, EvidenceNeedsRenewal} {
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
