package enums

import "io"

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

var watermarkStatusValues = []WatermarkStatus{
	WatermarkStatusPending,
	WatermarkStatusInProgress,
	WatermarkStatusSuccess,
	WatermarkStatusFailed,
	WatermarkStatusDisabled,
}

// Values returns a slice of strings representing all valid WatermarkStatus values.
func (WatermarkStatus) Values() []string { return stringValues(watermarkStatusValues) }

// String returns the WatermarkStatus as a string
func (r WatermarkStatus) String() string { return string(r) }

// ToWatermarkStatus returns the watermark status enum based on string input
func ToWatermarkStatus(r string) *WatermarkStatus {
	return parse(r, watermarkStatusValues, &WatermarkStatusInvalid)
}

// MarshalGQL implement the Marshaler interface for gqlgen
func (r WatermarkStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *WatermarkStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
