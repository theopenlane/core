package enums

import "io"

type ScanStatus string

var (
	ScanStatusPending    ScanStatus = "PENDING"
	ScanStatusProcessing ScanStatus = "PROCESSING"
	ScanStatusCompleted  ScanStatus = "COMPLETED"
	ScanStatusFailed     ScanStatus = "FAILED"
	ScanStatusInvalid    ScanStatus = "INVALID"
)

var scanStatusValues = []ScanStatus{ScanStatusPending, ScanStatusProcessing, ScanStatusCompleted, ScanStatusFailed}

// Values returns a slice of strings that represents all the possible values of the ScanStatus enum.
func (ScanStatus) Values() []string { return stringValues(scanStatusValues) }

// String returns the ScanStatus as a string
func (s ScanStatus) String() string { return string(s) }

// ToScanStatus returns the ScanStatus based on string input
func ToScanStatus(str string) *ScanStatus { return parse(str, scanStatusValues, &ScanStatusInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (s ScanStatus) MarshalGQL(w io.Writer) { marshalGQL(s, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (s *ScanStatus) UnmarshalGQL(v any) error { return unmarshalGQL(s, v) }
