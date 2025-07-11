package enums

import (
	"fmt"
	"io"
	"strings"
)

// ExportType is a custom type representing the various states of ExportType.
type ExportType string

var (
	// ExportTypeControl indicates the control.
	ExportTypeControl ExportType = "CONTROL"
	// ExportTypeEvidence indicates the evidence.
	ExportTypeEvidence ExportType = "EVIDENCE"
	// ExportTypeInternalpolicy indicates the internalpolicy.
	ExportTypeInternalpolicy ExportType = "INTERNALPOLICY"
	// ExportTypeProcedure indicates the procedure.
	ExportTypeProcedure ExportType = "PROCEDURE"
	// ExportTypeRisk indicates the risk.
	ExportTypeRisk ExportType = "RISK"
	// ExportTypeSubscriber indicates the subscriber.
	ExportTypeSubscriber ExportType = "SUBSCRIBER"
	// ExportTypeTask indicates the task.
	ExportTypeTask ExportType = "TASK"
	// ExportTypeInvalid is used when an unknown or unsupported value is provided.
	ExportTypeInvalid ExportType = "EXPORTTYPE_INVALID"
)

// Values returns a slice of strings representing all valid ExportType values.
func (ExportType) Values() []string {
	return []string{
		string(ExportTypeControl),
		string(ExportTypeEvidence),
		string(ExportTypeInternalpolicy),
		string(ExportTypeProcedure),
		string(ExportTypeRisk),
		string(ExportTypeSubscriber),
		string(ExportTypeTask),
	}
}

// String returns the string representation of the ExportType value.
func (r ExportType) String() string {
	return string(r)
}

// ToExportType converts a string to its corresponding ExportType enum value.
func ToExportType(r string) *ExportType {
	switch strings.ToUpper(r) {
	case ExportTypeControl.String():
		return &ExportTypeControl
	case ExportTypeEvidence.String():
		return &ExportTypeEvidence
	case ExportTypeInternalpolicy.String():
		return &ExportTypeInternalpolicy
	case ExportTypeProcedure.String():
		return &ExportTypeProcedure
	case ExportTypeRisk.String():
		return &ExportTypeRisk
	case ExportTypeSubscriber.String():
		return &ExportTypeSubscriber
	case ExportTypeTask.String():
		return &ExportTypeTask
	default:
		return &ExportTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ExportType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ExportType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ExportType, got: %T", v) //nolint:err113
	}

	*r = ExportType(str)

	return nil
}
