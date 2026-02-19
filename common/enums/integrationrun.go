package enums

import "io"

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

var integrationRunStatusValues = []IntegrationRunStatus{
	IntegrationRunStatusPending,
	IntegrationRunStatusRunning,
	IntegrationRunStatusSuccess,
	IntegrationRunStatusFailed,
	IntegrationRunStatusCancelled,
}

// IntegrationRunStatuses is a list of all valid IntegrationRunStatus values.
var IntegrationRunStatuses = stringValues(integrationRunStatusValues)

// Values returns a slice of strings that represents all the possible values of the IntegrationRunStatus enum.
func (IntegrationRunStatus) Values() []string { return IntegrationRunStatuses }

// String returns the IntegrationRunStatus as a string.
func (r IntegrationRunStatus) String() string { return string(r) }

// ToIntegrationRunStatus returns the IntegrationRunStatus based on string input.
func ToIntegrationRunStatus(r string) *IntegrationRunStatus {
	return parse(r, integrationRunStatusValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationRunStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationRunStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

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

var integrationOperationKindValues = []IntegrationOperationKind{
	IntegrationOperationKindSync,
	IntegrationOperationKindPush,
	IntegrationOperationKindPull,
	IntegrationOperationKindWebhook,
	IntegrationOperationKindScheduled,
}

// IntegrationOperationKinds is a list of all valid IntegrationOperationKind values.
var IntegrationOperationKinds = stringValues(integrationOperationKindValues)

// Values returns a slice of strings that represents all the possible values of the IntegrationOperationKind enum.
func (IntegrationOperationKind) Values() []string { return IntegrationOperationKinds }

// String returns the IntegrationOperationKind as a string.
func (r IntegrationOperationKind) String() string { return string(r) }

// ToIntegrationOperationKind returns the IntegrationOperationKind based on string input.
func ToIntegrationOperationKind(r string) *IntegrationOperationKind {
	return parse(r, integrationOperationKindValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationOperationKind) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationOperationKind) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }

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

var integrationRunTypeValues = []IntegrationRunType{
	IntegrationRunTypeManual,
	IntegrationRunTypeScheduled,
	IntegrationRunTypeWebhook,
	IntegrationRunTypeEvent,
}

// IntegrationRunTypes is a list of all valid IntegrationRunType values.
var IntegrationRunTypes = stringValues(integrationRunTypeValues)

// Values returns a slice of strings that represents all the possible values of the IntegrationRunType enum.
func (IntegrationRunType) Values() []string { return IntegrationRunTypes }

// String returns the IntegrationRunType as a string.
func (r IntegrationRunType) String() string { return string(r) }

// ToIntegrationRunType returns the IntegrationRunType based on string input.
func ToIntegrationRunType(r string) *IntegrationRunType {
	return parse(r, integrationRunTypeValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationRunType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationRunType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
