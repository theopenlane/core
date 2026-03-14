package enums

import "io"

// TemplateContext identifies the runtime data context available when a template is rendered.
// It controls which variable keys are valid for substitution and is used by the UI to
// populate the variable picker.
type TemplateContext string

var (
	// TemplateContextCampaignRecipient is used for campaign emails sent to individual recipients.
	// Provides: Recipient.{Email, FirstName, LastName}, Campaign.{Name, ID}, Company.{Name, LogoURL, SupportEmail}, URLS.*
	TemplateContextCampaignRecipient TemplateContext = "CAMPAIGN_RECIPIENT"
	// TemplateContextTransactional is used for system-triggered transactional emails.
	// Provides: URLS.*, CompanyName, SupportEmail, LogoURL, Recipient.{Email, FirstName, LastName}, plus per-template extras
	TemplateContextTransactional TemplateContext = "TRANSACTIONAL"
	// TemplateContextWorkflowAction is used for emails triggered by workflow action execution.
	// Provides: variables declared by the triggering workflow action's output schema
	TemplateContextWorkflowAction TemplateContext = "WORKFLOW_ACTION"
	// TemplateContextInvalid represents an invalid or unset template context.
	TemplateContextInvalid TemplateContext = "INVALID"
)

var templateContextValues = []TemplateContext{
	TemplateContextCampaignRecipient,
	TemplateContextTransactional,
	TemplateContextWorkflowAction,
}

// Values returns a slice of strings representing all valid TemplateContext enum values.
func (TemplateContext) Values() []string { return stringValues(templateContextValues) }

// String returns the TemplateContext as a string.
func (r TemplateContext) String() string { return string(r) }

// ToTemplateContext returns the TemplateContext enum based on string input.
func ToTemplateContext(r string) *TemplateContext {
	return parse(r, templateContextValues, &TemplateContextInvalid)
}

// MarshalGQL implements the gqlgen Marshaler interface.
func (r TemplateContext) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implements the gqlgen Unmarshaler interface.
func (r *TemplateContext) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
