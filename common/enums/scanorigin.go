package enums

import "io"

// ScanOrigin records how a scan was created, derived from the caller rather than supplied as input
type ScanOrigin string

var (
	// ScanOriginUser marks a scan created by a logged-in user session, e.g. a console upload
	ScanOriginUser ScanOrigin = "USER"
	// ScanOriginSystem marks a scan created by an internal operation such as a scheduled or listener-driven scan
	ScanOriginSystem ScanOrigin = "SYSTEM"
	// ScanOriginIntegration marks a scan created by an installed integration
	ScanOriginIntegration ScanOrigin = "INTEGRATION"
	// ScanOriginAPI marks a scan created with a personal access or API token
	ScanOriginAPI ScanOrigin = "API"
	// ScanOriginInvalid is returned when parsing an unknown value
	ScanOriginInvalid ScanOrigin = "INVALID"
)

var scanOriginValues = []ScanOrigin{ScanOriginUser, ScanOriginSystem, ScanOriginIntegration, ScanOriginAPI}

// Values returns a slice of strings that represents all the possible values of the ScanOrigin enum
func (ScanOrigin) Values() []string { return stringValues(scanOriginValues) }

// String returns the ScanOrigin as a string
func (s ScanOrigin) String() string { return string(s) }

// ToScanOrigin returns the ScanOrigin based on string input
func ToScanOrigin(str string) *ScanOrigin { return parse(str, scanOriginValues, &ScanOriginInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (s ScanOrigin) MarshalGQL(w io.Writer) { marshalGQL(s, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (s *ScanOrigin) UnmarshalGQL(v any) error { return unmarshalGQL(s, v) }
