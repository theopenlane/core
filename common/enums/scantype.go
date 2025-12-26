package enums

import (
	"fmt"
	"io"
	"strings"
)

type ScanType string

var (
	ScanTypeDomain        ScanType = "DOMAIN"
	ScanTypeVulnerability ScanType = "VULNERABILITY"
	ScanTypeVendor        ScanType = "VENDOR"
	ScanTypeProvider      ScanType = "PROVIDER"
	ScanTypeInvalid       ScanType = "INVALID"
)

func (ScanType) Values() []string {
	return []string{
		string(ScanTypeDomain),
		string(ScanTypeVulnerability),
		string(ScanTypeVendor),
		string(ScanTypeProvider),
	}
}

func (s ScanType) String() string { return string(s) }

func ToScanType(str string) *ScanType {
	switch strings.ToUpper(str) {
	case ScanTypeDomain.String():
		return &ScanTypeDomain
	case ScanTypeVulnerability.String():
		return &ScanTypeVulnerability
	case ScanTypeVendor.String():
		return &ScanTypeVendor
	case ScanTypeProvider.String():
		return &ScanTypeProvider
	default:
		return &ScanTypeInvalid
	}
}

func (s ScanType) MarshalGQL(w io.Writer) { _, _ = w.Write([]byte(`"` + s.String() + `"`)) }

func (s *ScanType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ScanType, got: %T", v) //nolint:err113
	}

	*s = ScanType(str)

	return nil
}
