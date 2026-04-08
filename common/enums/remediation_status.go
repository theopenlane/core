package enums

import "io"

type RemediationStatus string

var (
	RemediationStatusOpen       RemediationStatus = "OPEN"
	RemediationStatusInProgress RemediationStatus = "IN_PROGRESS"
	RemediationStatusInReview   RemediationStatus = "IN_REVIEW"
	RemediationStatusCompleted  RemediationStatus = "COMPLETED"
	RemediationStatusWontDo     RemediationStatus = "WONT_DO"
	RemediationStatusInvalid    RemediationStatus = "INVALID"
)

var remediationStatusValues = []RemediationStatus{RemediationStatusOpen, RemediationStatusInProgress, RemediationStatusInReview, RemediationStatusCompleted, RemediationStatusWontDo}

// Values returns a slice of strings that represents all the possible values of the RemediationStatus enum.
// Possible default values are "OPEN", "IN_PROGRESS", "IN_REVIEW", "COMPLETED", and "WONT_DO".
func (RemediationStatus) Values() []string { return stringValues(remediationStatusValues) }

// String returns the RemediationStatus as a string
func (r RemediationStatus) String() string { return string(r) }

// ToRemediationStatus returns the remediation status enum based on string input
func ToRemediationStatus(r string) *RemediationStatus {
	return parse(r, remediationStatusValues, &RemediationStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r RemediationStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *RemediationStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
