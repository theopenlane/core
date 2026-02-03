package enums

import (
	"fmt"
	"io"
	"strings"
)

// IntegrationRunStatus represents the status of an integration run.
type IntegrationRunStatus string

var (
	// IntegrationRunStatusPending indicates the run is pending execution.
	IntegrationRunStatusPending IntegrationRunStatus = "PENDING"
	// IntegrationRunStatusRunning indicates the run is currently executing.
	IntegrationRunStatusRunning IntegrationRunStatus = "RUNNING"
	// IntegrationRunStatusSuccess indicates the run completed successfully.
	IntegrationRunStatusSuccess IntegrationRunStatus = "SUCCESS"
	// IntegrationRunStatusFailed indicates the run failed.
	IntegrationRunStatusFailed IntegrationRunStatus = "FAILED"
	// IntegrationRunStatusCancelled indicates the run was cancelled.
	IntegrationRunStatusCancelled IntegrationRunStatus = "CANCELLED"
)

// IntegrationRunStatuses is a list of all valid IntegrationRunStatus values.
var IntegrationRunStatuses = []string{
	string(IntegrationRunStatusPending),
	string(IntegrationRunStatusRunning),
	string(IntegrationRunStatusSuccess),
	string(IntegrationRunStatusFailed),
	string(IntegrationRunStatusCancelled),
}

// Values returns a slice of strings that represents all the possible values of the IntegrationRunStatus enum.
func (IntegrationRunStatus) Values() (kinds []string) {
	return IntegrationRunStatuses
}

// String returns the IntegrationRunStatus as a string.
func (r IntegrationRunStatus) String() string {
	return string(r)
}

// ToIntegrationRunStatus returns the IntegrationRunStatus based on string input.
func ToIntegrationRunStatus(r string) *IntegrationRunStatus {
	switch strings.ToUpper(r) {
	case IntegrationRunStatusPending.String():
		return &IntegrationRunStatusPending
	case IntegrationRunStatusRunning.String():
		return &IntegrationRunStatusRunning
	case IntegrationRunStatusSuccess.String():
		return &IntegrationRunStatusSuccess
	case IntegrationRunStatusFailed.String():
		return &IntegrationRunStatusFailed
	case IntegrationRunStatusCancelled.String():
		return &IntegrationRunStatusCancelled
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationRunStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationRunStatus) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for IntegrationRunStatus, got: %T", v) //nolint:err113
	}

	*r = IntegrationRunStatus(str)

	return nil
}

// IntegrationOperationKind represents the kind of operation for an integration run.
type IntegrationOperationKind string

var (
	// IntegrationOperationKindSync indicates a sync operation.
	IntegrationOperationKindSync IntegrationOperationKind = "SYNC"
	// IntegrationOperationKindPush indicates a push operation.
	IntegrationOperationKindPush IntegrationOperationKind = "PUSH"
	// IntegrationOperationKindPull indicates a pull operation.
	IntegrationOperationKindPull IntegrationOperationKind = "PULL"
	// IntegrationOperationKindWebhook indicates a webhook operation.
	IntegrationOperationKindWebhook IntegrationOperationKind = "WEBHOOK"
	// IntegrationOperationKindScheduled indicates a scheduled operation.
	IntegrationOperationKindScheduled IntegrationOperationKind = "SCHEDULED"
)

// IntegrationOperationKinds is a list of all valid IntegrationOperationKind values.
var IntegrationOperationKinds = []string{
	string(IntegrationOperationKindSync),
	string(IntegrationOperationKindPush),
	string(IntegrationOperationKindPull),
	string(IntegrationOperationKindWebhook),
	string(IntegrationOperationKindScheduled),
}

// Values returns a slice of strings that represents all the possible values of the IntegrationOperationKind enum.
func (IntegrationOperationKind) Values() (kinds []string) {
	return IntegrationOperationKinds
}

// String returns the IntegrationOperationKind as a string.
func (r IntegrationOperationKind) String() string {
	return string(r)
}

// ToIntegrationOperationKind returns the IntegrationOperationKind based on string input.
func ToIntegrationOperationKind(r string) *IntegrationOperationKind {
	switch strings.ToUpper(r) {
	case IntegrationOperationKindSync.String():
		return &IntegrationOperationKindSync
	case IntegrationOperationKindPush.String():
		return &IntegrationOperationKindPush
	case IntegrationOperationKindPull.String():
		return &IntegrationOperationKindPull
	case IntegrationOperationKindWebhook.String():
		return &IntegrationOperationKindWebhook
	case IntegrationOperationKindScheduled.String():
		return &IntegrationOperationKindScheduled
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationOperationKind) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationOperationKind) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for IntegrationOperationKind, got: %T", v) //nolint:err113
	}

	*r = IntegrationOperationKind(str)

	return nil
}

// IntegrationRunType represents the type of integration run.
type IntegrationRunType string

var (
	// IntegrationRunTypeManual indicates a manually triggered run.
	IntegrationRunTypeManual IntegrationRunType = "MANUAL"
	// IntegrationRunTypeScheduled indicates a scheduled run.
	IntegrationRunTypeScheduled IntegrationRunType = "SCHEDULED"
	// IntegrationRunTypeWebhook indicates a webhook-triggered run.
	IntegrationRunTypeWebhook IntegrationRunType = "WEBHOOK"
	// IntegrationRunTypeEvent indicates an event-triggered run.
	IntegrationRunTypeEvent IntegrationRunType = "EVENT"
)

// IntegrationRunTypes is a list of all valid IntegrationRunType values.
var IntegrationRunTypes = []string{
	string(IntegrationRunTypeManual),
	string(IntegrationRunTypeScheduled),
	string(IntegrationRunTypeWebhook),
	string(IntegrationRunTypeEvent),
}

// Values returns a slice of strings that represents all the possible values of the IntegrationRunType enum.
func (IntegrationRunType) Values() (kinds []string) {
	return IntegrationRunTypes
}

// String returns the IntegrationRunType as a string.
func (r IntegrationRunType) String() string {
	return string(r)
}

// ToIntegrationRunType returns the IntegrationRunType based on string input.
func ToIntegrationRunType(r string) *IntegrationRunType {
	switch strings.ToUpper(r) {
	case IntegrationRunTypeManual.String():
		return &IntegrationRunTypeManual
	case IntegrationRunTypeScheduled.String():
		return &IntegrationRunTypeScheduled
	case IntegrationRunTypeWebhook.String():
		return &IntegrationRunTypeWebhook
	case IntegrationRunTypeEvent.String():
		return &IntegrationRunTypeEvent
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationRunType) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationRunType) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for IntegrationRunType, got: %T", v) //nolint:err113
	}

	*r = IntegrationRunType(str)

	return nil
}
