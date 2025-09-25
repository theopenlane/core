package enums

import (
	"fmt"
	"io"
	"strings"
)

type WatermarkStatus string

var (
	WatermarkStatusDisabled WatermarkStatus = "DISABLED"
	// WatermarkStatusPending indicates that the watermarking job is pending
	WatermarkStatusPending WatermarkStatus = "PENDING"
	// WatermarkStatusInProgress indicates that the watermarking job is in progress
	WatermarkStatusInProgress WatermarkStatus = "IN_PROGRESS"
	// WatermarkStatusSuccess indicates that the watermarking job has completed successfully
	WatermarkStatusSuccess WatermarkStatus = "SUCCESS"
	// WatermarkStatusFailed indicates that the watermarking job has failed
	WatermarkStatusFailed  WatermarkStatus = "FAILED"
	WatermarkStatusInvalid WatermarkStatus = "INVALID"
)

func (WatermarkStatus) Values() []string {
	return []string{
		string(WatermarkStatusPending),
		string(WatermarkStatusInProgress),
		string(WatermarkStatusSuccess),
		string(WatermarkStatusFailed),
		string(WatermarkStatusDisabled),
	}
}

func (w WatermarkStatus) String() string { return string(w) }

func ToWatermarkStatus(str string) *WatermarkStatus {
	switch strings.ToUpper(str) {
	case WatermarkStatusPending.String():
		return &WatermarkStatusPending
	case WatermarkStatusInProgress.String():
		return &WatermarkStatusInProgress
	case WatermarkStatusSuccess.String():
		return &WatermarkStatusSuccess
	case WatermarkStatusFailed.String():
		return &WatermarkStatusFailed
	case WatermarkStatusDisabled.String():
		return &WatermarkStatusDisabled
	default:
		return &WatermarkStatusInvalid
	}
}

func (w WatermarkStatus) MarshalGQL(i io.Writer) { _, _ = i.Write([]byte(`"` + w.String() + `"`)) }

func (w *WatermarkStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for WatermarkStatus, got: %T", v) //nolint:err113
	}

	*w = WatermarkStatus(str)

	return nil
}
