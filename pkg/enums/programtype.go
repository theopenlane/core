package enums

import (
	"fmt"
	"io"
	"strings"
)

// ProgramType is a custom type representing the various states of ProgramType.
type ProgramType string

var (
	// ProgramTypeFramework indicates the framework.
	ProgramTypeFramework ProgramType = "FRAMEWORK"
	// ProgramTypeGapAnalysis indicates the gap analysis.
	ProgramTypeGapAnalysis ProgramType = "GAP_ANALYSIS"
	// ProgramTypeRiskAssessment indicates the risk assessment.
	ProgramTypeRiskAssessment ProgramType = "RISK_ASSESSMENT"
	// ProgramTypeOther indicates the other.
	ProgramTypeOther ProgramType = "OTHER"
	// ProgramTypeInvalid is used when an unknown or unsupported value is provided.
	ProgramTypeInvalid ProgramType = "PROGRAMTYPE_INVALID"
)

// Values returns a slice of strings representing all valid ProgramType values.
func (ProgramType) Values() []string {
	return []string{
		string(ProgramTypeFramework),
		string(ProgramTypeGapAnalysis),
		string(ProgramTypeRiskAssessment),
		string(ProgramTypeOther),
	}
}

// String returns the string representation of the ProgramType value.
func (r ProgramType) String() string {
	return string(r)
}

// ToProgramType converts a string to its corresponding ProgramType enum value.
func ToProgramType(r string) *ProgramType {
	switch strings.ToUpper(r) {
	case ProgramTypeFramework.String():
		return &ProgramTypeFramework
	case ProgramTypeGapAnalysis.String():
		return &ProgramTypeGapAnalysis
	case ProgramTypeRiskAssessment.String():
		return &ProgramTypeRiskAssessment
	case ProgramTypeOther.String():
		return &ProgramTypeOther
	default:
		return &ProgramTypeInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ProgramType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ProgramType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ProgramType, got: %T", v) //nolint:err113
	}

	*r = ProgramType(str)

	return nil
}
