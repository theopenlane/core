package enums

import (
	"fmt"
	"io"
	"strings"
)

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

// IntegrationWebhookStatuses is a list of all valid IntegrationWebhookStatus values.
var IntegrationWebhookStatuses = []string{
	string(IntegrationWebhookStatusActive),
	string(IntegrationWebhookStatusInactive),
	string(IntegrationWebhookStatusFailed),
	string(IntegrationWebhookStatusPending),
}

// Values returns a slice of strings that represents all the possible values of the IntegrationWebhookStatus enum.
func (IntegrationWebhookStatus) Values() (kinds []string) {
	return IntegrationWebhookStatuses
}

// String returns the IntegrationWebhookStatus as a string.
func (r IntegrationWebhookStatus) String() string {
	return string(r)
}

// ToIntegrationWebhookStatus returns the IntegrationWebhookStatus based on string input.
func ToIntegrationWebhookStatus(r string) *IntegrationWebhookStatus {
	switch strings.ToUpper(r) {
	case IntegrationWebhookStatusActive.String():
		return &IntegrationWebhookStatusActive
	case IntegrationWebhookStatusInactive.String():
		return &IntegrationWebhookStatusInactive
	case IntegrationWebhookStatusFailed.String():
		return &IntegrationWebhookStatusFailed
	case IntegrationWebhookStatusPending.String():
		return &IntegrationWebhookStatusPending
	default:
		return nil
	}
}

// MarshalGQL implement the Marshaler interface for gqlgen.
func (r IntegrationWebhookStatus) MarshalGQL(w io.Writer) {
	_, _ = w.Write([]byte(`"` + r.String() + `"`))
}

// UnmarshalGQL implement the Unmarshaler interface for gqlgen.
func (r *IntegrationWebhookStatus) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for IntegrationWebhookStatus, got: %T", v) //nolint:err113
	}

	*r = IntegrationWebhookStatus(str)

	return nil
}
