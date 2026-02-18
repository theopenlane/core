package enums

import "io"

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

var scheduledJobRunStatusValues = []ScheduledJobRunStatus{ScheduledJobRunStatusPending, ScheduledJobRunStatusAcquired}

// Values returns a slice of strings representing all valid ScheduledJobRunStatus values.
func (ScheduledJobRunStatus) Values() []string { return stringValues(scheduledJobRunStatusValues) }

// String returns the string representation of the ScheduledJobRunStatus value.
func (r ScheduledJobRunStatus) String() string { return string(r) }

// ToScheduledJobRunStatus converts a string to its corresponding ScheduledJobRunStatus enum value.
func ToScheduledJobRunStatus(r string) *ScheduledJobRunStatus {
	return parse(r, scheduledJobRunStatusValues, &ScheduledJobRunStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r ScheduledJobRunStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *ScheduledJobRunStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
