package enums

import "io"

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

var programTypeValues = []ProgramType{ProgramTypeFramework, ProgramTypeGapAnalysis, ProgramTypeRiskAssessment, ProgramTypeOther}

// Values returns a slice of strings representing all valid ProgramType values.
func (ProgramType) Values() []string { return stringValues(programTypeValues) }

// String returns the string representation of the ProgramType value.
func (r ProgramType) String() string { return string(r) }

// ToProgramType converts a string to its corresponding ProgramType enum value.
func ToProgramType(r string) *ProgramType { return parse(r, programTypeValues, &ProgramTypeInvalid) }

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ProgramType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ProgramType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
