package enums

import (
	"fmt"
	"io"
	"strings"
)

// ScheduledJobRunStatus is a custom type representing the various states of ScheduledJobRunStatus.
type ScheduledJobRunStatus string

var (
	// ScheduledJobRunStatusPending indicates the pending.
	ScheduledJobRunStatusPending ScheduledJobRunStatus = "PENDING"
	// ScheduledJobRunStatusAcquired indicates the acquired.
	ScheduledJobRunStatusAcquired ScheduledJobRunStatus = "ACQUIRED"
	// ScheduledJobRunStatusInvalid is used when an unknown or unsupported value is provided.
	ScheduledJobRunStatusInvalid ScheduledJobRunStatus = "SCHEDULEDJOBRUNSTATUS_INVALID"
)

// Values returns a slice of strings representing all valid ScheduledJobRunStatus values.
func (ScheduledJobRunStatus) Values() []string {
	return []string{
		string(ScheduledJobRunStatusPending),
		string(ScheduledJobRunStatusAcquired),
	}
}

// String returns the string representation of the ScheduledJobRunStatus value.
func (r ScheduledJobRunStatus) String() string {
	return strings.ToUpper(string(r))
}

// ToScheduledJobRunStatus converts a string to its corresponding ScheduledJobRunStatus enum value.
func ToScheduledJobRunStatus(r string) *ScheduledJobRunStatus {
	switch strings.ToUpper(r) {
	case ScheduledJobRunStatusPending.String():
		return &ScheduledJobRunStatusPending
	case ScheduledJobRunStatusAcquired.String():
		return &ScheduledJobRunStatusAcquired
	default:
		return &ScheduledJobRunStatusInvalid
	}
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ScheduledJobRunStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ScheduledJobRunStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for ScheduledJobRunStatus, got: %T", v) //nolint:err113
	}

	*r = ScheduledJobRunStatus(str)

	return nil
}
