package enums

import "io"

type ScanType string

var (
	ScanTypeDomain        ScanType = "DOMAIN"
	ScanTypeVulnerability ScanType = "VULNERABILITY"
	ScanTypeVendor        ScanType = "VENDOR"
	ScanTypeProvider      ScanType = "PROVIDER"
	ScanTypeInvalid       ScanType = "INVALID"
)

var scanTypeValues = []ScanType{ScanTypeDomain, ScanTypeVulnerability, ScanTypeVendor, ScanTypeProvider}

// Values returns a slice of strings that represents all the possible values of the ScanType enum.
func (ScanType) Values() []string { return stringValues(scanTypeValues) }

// String returns the ScanType as a string
func (s ScanType) String() string { return string(s) }

// ToScanType returns the ScanType based on string input
func ToScanType(str string) *ScanType { return parse(str, scanTypeValues, &ScanTypeInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (s ScanType) MarshalGQL(w io.Writer) { marshalGQL(s, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (s *ScanType) UnmarshalGQL(v any) error { return unmarshalGQL(s, v) }
