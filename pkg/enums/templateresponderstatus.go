package enums

import (
	"fmt"
	"io"
	"strings"
)

// TemplateResponderStatus is a custom type representing the various states of TemplateResponderStatus.
type TemplateResponderStatus string

var (
	// TemplateResponderStatusPending indicates the pending.
	TemplateResponderStatusPending TemplateResponderStatus = "PENDING"
	// TemplateResponderStatusSent indicates the sent.
	TemplateResponderStatusSent TemplateResponderStatus = "SENT"
	// TemplateResponderStatusViewed indicates the viewed.
	TemplateResponderStatusViewed TemplateResponderStatus = "VIEWED"
	// TemplateResponderStatusInProgress indicates the in progress.
	TemplateResponderStatusInProgress TemplateResponderStatus = "IN_PROGRESS"
	// TemplateResponderStatusCompleted indicates the completed.
	TemplateResponderStatusCompleted TemplateResponderStatus = "COMPLETED"
	// TemplateResponderStatusInvalid is used when an unknown or unsupported value is provided.
	TemplateResponderStatusInvalid TemplateResponderStatus = "TEMPLATERESPONDERSTATUS_INVALID"
)

// Values returns a slice of strings representing all valid TemplateResponderStatus values.
func (TemplateResponderStatus) Values() []string {
	return []string{
		string(TemplateResponderStatusPending),
		string(TemplateResponderStatusSent),
		string(TemplateResponderStatusViewed),
		string(TemplateResponderStatusInProgress),
		string(TemplateResponderStatusCompleted),
	}
}

// String returns the string representation of the TemplateResponderStatus value.
func (r TemplateResponderStatus) String() string {
	return string(r)
}

// ToTemplateResponderStatus converts a string to its corresponding TemplateResponderStatus enum value.
func ToTemplateResponderStatus(r string) *TemplateResponderStatus {
	switch strings.ToUpper(r) {
	case TemplateResponderStatusPending.String():
		return &TemplateResponderStatusPending
	case TemplateResponderStatusSent.String():
		return &TemplateResponderStatusSent
	case TemplateResponderStatusViewed.String():
		return &TemplateResponderStatusViewed
	case TemplateResponderStatusInProgress.String():
		return &TemplateResponderStatusInProgress
	case TemplateResponderStatusCompleted.String():
		return &TemplateResponderStatusCompleted
	default:
		return &TemplateResponderStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateResponderStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateResponderStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for TemplateResponderStatus, got: %T", v)  //nolint:err113
	}

	*r = TemplateResponderStatus(str)

	return nil
}
