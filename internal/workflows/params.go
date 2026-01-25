package workflows

import "github.com/theopenlane/core/common/enums"

// TargetedActionParams captures workflow action params that target recipients
type TargetedActionParams struct {
	// Targets identifies users, groups, roles, or resolvers to receive the action
	Targets []TargetConfig `json:"targets"`
}

// FieldUpdateActionParams defines params for FIELD_UPDATE actions
type FieldUpdateActionParams struct {
	// Updates maps field names to new values
	Updates map[string]any `json:"updates"`
}

// ApprovalActionParams defines params for APPROVAL actions
type ApprovalActionParams struct {
	// TargetedActionParams identifies the approval recipients
	TargetedActionParams
	// Required defaults to true when omitted
	Required *bool `json:"required"`
	// Label is an optional display label for the approval action
	Label string `json:"label"`
	// RequiredCount sets a quorum threshold (number of approvals needed) for this action
	RequiredCount int `json:"required_count"`
	// Fields lists the approval-gated fields for domain derivation
	Fields []string `json:"fields,omitempty"`
}

// NotificationActionParams defines params for NOTIFICATION actions
type NotificationActionParams struct {
	// TargetedActionParams identifies the notification recipients
	TargetedActionParams
	// Channels selects notification delivery channels
	Channels []enums.Channel `json:"channels"`
	// Topic sets an optional notification topic
	Topic string `json:"topic"`
	// Title is the notification title
	Title string `json:"title"`
	// Body is the notification body
	Body string `json:"body"`
	// Data is an optional payload merged into the notification data
	Data map[string]any `json:"data"`
}

// WebhookActionParams defines params for WEBHOOK actions
type WebhookActionParams struct {
	// URL is the webhook endpoint
	URL string `json:"url"`
	// Method is the HTTP method for the webhook request
	Method string `json:"method"`
	// Headers are additional HTTP headers for the webhook request
	Headers map[string]string `json:"headers"`
	// Payload is merged into the base webhook payload
	Payload map[string]any `json:"payload"`
	// TimeoutMS overrides the webhook timeout in milliseconds
	TimeoutMS int `json:"timeout_ms"`
	// Secret signs the webhook payload if provided
	Secret string `json:"secret"`
	// Retries overrides the retry count when non-zero
	Retries int `json:"retries"`

	// Optional override for the idempotency key header
	IdempotencyKey string `json:"idempotency_key"`
}

// IntegrationActionParams defines params for INTEGRATION actions
type IntegrationActionParams struct {
	// Integration is the integration identifier for the operation
	Integration string `json:"integration"`
	// Provider overrides the integration identifier when set
	Provider string `json:"provider"`
	// Operation is the integration operation name
	Operation string `json:"operation"`
	// Config holds the integration-specific configuration payload
	Config map[string]any `json:"config"`

	// TimeoutMS overrides the operation timeout in milliseconds
	TimeoutMS int `json:"timeout_ms"`
	// Retries overrides the retry count when non-zero
	Retries int `json:"retries"`
	// Force requests a refresh for the provider
	Force bool `json:"force_refresh"`
	// ClientForce requests a client-side refresh for the provider
	ClientForce bool `json:"client_force"`
}
