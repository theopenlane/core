package enums

import (
	"fmt"
	"io"
	"strings"
)

type ScanStatus string

var (
	ScanStatusPending    ScanStatus = "PENDING"
	ScanStatusProcessing ScanStatus = "PROCESSING"
	ScanStatusCompleted  ScanStatus = "COMPLETED"
	ScanStatusFailed     ScanStatus = "FAILED"
	ScanStatusInvalid    ScanStatus = "INVALID"
)

func (ScanStatus) Values() (kinds []string) {
	for _, s := range []ScanStatus{ScanStatusPending, ScanStatusProcessing, ScanStatusCompleted, ScanStatusFailed} {
		kinds = append(kinds, string(s))
	}

	return
}

func (s ScanStatus) String() string { return string(s) }

func ToScanStatus(str string) *ScanStatus {
	switch strings.ToUpper(str) {
	case ScanStatusPending.String():
		return &ScanStatusPending
	case ScanStatusProcessing.String():
		return &ScanStatusProcessing
	case ScanStatusCompleted.String():
		return &ScanStatusCompleted
	case ScanStatusFailed.String():
		return &ScanStatusFailed
	default:
		return &ScanStatusInvalid
	}
}

func (s ScanStatus) MarshalGQL(w io.Writer) { _, _ = w.Write([]byte(`"` + s.String() + `"`)) }

func (s *ScanStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ScanStatus, got: %T", v) //nolint:err113
	}

	*s = ScanStatus(str)

	return nil
}
