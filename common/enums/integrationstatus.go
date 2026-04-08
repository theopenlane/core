package enums

import "io"

// IntegrationStatus represents the lifecycle state of an installed integration
type IntegrationStatus string

var (
	// IntegrationStatusPending indicates the installation has been created but is not yet fully connected
	IntegrationStatusPending IntegrationStatus = "PENDING"
	// IntegrationStatusConnected indicates the installation is configured and ready for use
	IntegrationStatusConnected IntegrationStatus = "CONNECTED"
	// IntegrationStatusErrored indicates the installation is present but currently unhealthy or misconfigured
	IntegrationStatusErrored IntegrationStatus = "ERRORED"
	// IntegrationStatusDisabled indicates the installation is intentionally disabled
	IntegrationStatusDisabled IntegrationStatus = "DISABLED"
	// IntegrationStatusDeleted indicates the installation has been deleted or is pending cleanup
	IntegrationStatusDeleted IntegrationStatus = "DELETED"
	// IntegrationStatusInvalid represents an invalid lifecycle status
	IntegrationStatusInvalid IntegrationStatus = "INVALID"
)

var integrationStatusValues = []IntegrationStatus{
	IntegrationStatusPending,
	IntegrationStatusConnected,
	IntegrationStatusErrored,
	IntegrationStatusDisabled,
	IntegrationStatusDeleted,
}

// Values returns the valid lifecycle states as strings
func (IntegrationStatus) Values() []string { return stringValues(integrationStatusValues) }

// String returns the enum value as a string
func (r IntegrationStatus) String() string { return string(r) }

// ToIntegrationStatus parses a string into an IntegrationStatus
func ToIntegrationStatus(r string) *IntegrationStatus {
	return parse(r, integrationStatusValues, &IntegrationStatusInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface
func (r IntegrationStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface
func (r *IntegrationStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
