package emailruntime

import (
	"github.com/theopenlane/newman/compose"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/integrations/operations"
	"github.com/theopenlane/core/common/models"
)

// ContextData is the base template data shape shared by all contexts.
// It embeds compose.Config (company, sender, URL fields) and adds a Recipient field.
type ContextData struct {
	compose.Config
	// Recipient holds per-recipient data injected at send time.
	Recipient compose.Recipient `json:"Recipient" jsonschema:"required,description=Recipient information"`
}

// CampaignData holds the template data shape for the CampaignRecipient context.
// It extends ContextData with campaign-specific metadata.
type CampaignData struct {
	ContextData
	// Campaign holds campaign metadata available in campaign recipient templates.
	Campaign struct {
		Name        string `json:"name"        jsonschema:"description=Name of the campaign"`
		Description string `json:"description" jsonschema:"description=Campaign description"`
	} `json:"Campaign" jsonschema:"description=Campaign metadata"`
}

// contextCatalog is the authoritative list of template data contexts.
// JSON schemas are reflected from typed data structs once at init time.
var contextCatalog = []models.TemplateContextEntry{
	{
		Context:     enums.TemplateContextCampaignRecipient,
		Label:       "Campaign Recipient",
		Description: "Used for campaign emails sent to individual recipients. Data is built per-recipient from the campaign target list.",
		Schema:      operations.SchemaFrom[CampaignData](),
	},
	{
		Context:     enums.TemplateContextTransactional,
		Label:       "Transactional",
		Description: "Used for system-triggered transactional emails such as email verification, password reset, invites, and billing notifications.",
		Schema:      operations.SchemaFrom[ContextData](),
	},
	{
		Context:     enums.TemplateContextWorkflowAction,
		Label:       "Workflow Action",
		Description: "Used for emails triggered by workflow action execution. The workflow definition may inject additional variables beyond the base set.",
		Schema:      operations.SchemaFrom[ContextData](),
	},
}

// TemplateContextSchema returns the reflected JSON Schema for the given TemplateContext,
// suitable for storing in the jsonconfig field of a template record.
func TemplateContextSchema(ctx enums.TemplateContext) map[string]any {
	for _, entry := range contextCatalog {
		if entry.Context == ctx {
			return entry.Schema
		}
	}

	return nil
}

// TemplateContextEntries returns all registered context entries.
func TemplateContextEntries() []models.TemplateContextEntry {
	return contextCatalog
}
