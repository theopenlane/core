package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// Values returns a slice of strings representing all valid ImpersonationAction values.
func (ImpersonationAction) Values() []string {
	return []string{
		string(ImpersonationActionStart),
		string(ImpersonationActionStop),
	}
}

// String returns the string representation of the ImpersonationAction value.
func (r ImpersonationAction) String() string {
	return string(r)
}

// ToImpersonationAction converts a string to its corresponding ImpersonationAction enum value.
func ToImpersonationAction(r string) *ImpersonationAction {
	switch strings.ToUpper(r) {
	case ImpersonationActionStart.String():
		return &ImpersonationActionStart
	case ImpersonationActionStop.String():
		return &ImpersonationActionStop
	default:
		return &ImpersonationActionInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ImpersonationAction) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ImpersonationAction) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ImpersonationAction, got: %T", v) //nolint:err113
	}

	*r = ImpersonationAction(str)

	return nil
}
