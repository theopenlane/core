package enums

import "io"

// ImpersonationAction is a custom type representing the various states of ImpersonationAction.
type ImpersonationAction string

var (
	// ImpersonationActionStart indicates the start.
	ImpersonationActionStart ImpersonationAction = "START"
	// ImpersonationActionStop indicates the stop.
	ImpersonationActionStop ImpersonationAction = "STOP"
	// ImpersonationActionInvalid is used when an unknown or unsupported value is provided.
	ImpersonationActionInvalid ImpersonationAction = "IMPERSONATIONACTION_INVALID"
)

var impersonationActionValues = []ImpersonationAction{ImpersonationActionStart, ImpersonationActionStop}

// Values returns a slice of strings representing all valid ImpersonationAction values.
func (ImpersonationAction) Values() []string { return stringValues(impersonationActionValues) }

// String returns the string representation of the ImpersonationAction value.
func (r ImpersonationAction) String() string { return string(r) }

// ToImpersonationAction converts a string to its corresponding ImpersonationAction enum value.
func ToImpersonationAction(r string) *ImpersonationAction {
	return parse(r, impersonationActionValues, &ImpersonationActionInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ImpersonationAction) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ImpersonationAction) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
