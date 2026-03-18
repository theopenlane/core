package enums

import "io"

// IntegrationWebhookStatus represents the status of an integration webhook.
type IntegrationWebhookStatus string

var (
	// IntegrationWebhookStatusActive indicates the webhook is active and receiving events.
	IntegrationWebhookStatusActive IntegrationWebhookStatus = "ACTIVE"
	// IntegrationWebhookStatusInactive indicates the webhook is inactive.
	IntegrationWebhookStatusInactive IntegrationWebhookStatus = "INACTIVE"
	// IntegrationWebhookStatusFailed indicates the webhook has failed delivery attempts.
	IntegrationWebhookStatusFailed IntegrationWebhookStatus = "FAILED"
	// IntegrationWebhookStatusPending indicates the webhook is pending verification.
	IntegrationWebhookStatusPending IntegrationWebhookStatus = "PENDING"
)

var integrationWebhookStatusValues = []IntegrationWebhookStatus{
	IntegrationWebhookStatusActive,
	IntegrationWebhookStatusInactive,
	IntegrationWebhookStatusFailed,
	IntegrationWebhookStatusPending,
}

// IntegrationWebhookStatuses is a list of all valid IntegrationWebhookStatus values.
var IntegrationWebhookStatuses = stringValues(integrationWebhookStatusValues)

// Values returns a slice of strings that represents all the possible values of the IntegrationWebhookStatus enum.
func (IntegrationWebhookStatus) Values() []string { return IntegrationWebhookStatuses }

// String returns the IntegrationWebhookStatus as a string.
func (r IntegrationWebhookStatus) String() string { return string(r) }

// ToIntegrationWebhookStatus returns the IntegrationWebhookStatus based on string input.
func ToIntegrationWebhookStatus(r string) *IntegrationWebhookStatus {
	return parse(r, integrationWebhookStatusValues, nil)
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationWebhookStatus) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationWebhookStatus) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
