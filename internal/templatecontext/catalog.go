package templatecontext

import (
	"encoding/json"
	"maps"
	"time"

	"github.com/samber/lo"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/pkg/jsonx"
)

// BrandingData holds email branding fields injected into template data when
// an EmailBranding record is associated with the template
type BrandingData struct {
	// Name is the branding configuration name
	Name string `json:"Name"`
	// BrandName is the customer-facing brand display name
	BrandName string `json:"BrandName"`
	// PrimaryColor is the primary brand color hex value
	PrimaryColor string `json:"PrimaryColor"`
	// SecondaryColor is the secondary brand color hex value
	SecondaryColor string `json:"SecondaryColor"`
	// BackgroundColor is the email background color hex value
	BackgroundColor string `json:"BackgroundColor"`
	// TextColor is the body text color hex value
	TextColor string `json:"TextColor"`
	// ButtonColor is the CTA button background color hex value
	ButtonColor string `json:"ButtonColor"`
	// ButtonTextColor is the CTA button text color hex value
	ButtonTextColor string `json:"ButtonTextColor"`
	// LinkColor is the hyperlink color hex value
	LinkColor string `json:"LinkColor"`
	// FontFamily is the resolved CSS font-family string
	FontFamily string `json:"FontFamily"`
	// IsDefault indicates whether this branding is the default for the owner
	IsDefault bool `json:"IsDefault"`
	// Logo is the branding logo URL, empty when no logo is configured
	Logo string `json:"Logo,omitempty"`
}

// RecipientData holds per-recipient data for template rendering.
// JSON tags use PascalCase to match template variable keys (e.g. {{ .Recipient.Email }})
type RecipientData struct {
	// Email is the recipient email address
	Email string `json:"Email" jsonschema:"required,description=Recipient email address"`
	// FirstName is the recipient first name
	FirstName string `json:"FirstName" jsonschema:"description=Recipient first name"`
	// LastName is the recipient last name
	LastName string `json:"LastName" jsonschema:"description=Recipient last name"`
}

// CampaignMeta holds campaign metadata available in campaign recipient templates
type CampaignMeta struct {
	// Name is the campaign name
	Name string `json:"Name" jsonschema:"description=Name of the campaign"`
	// Description is a longer form description of the campaign
	Description string `json:"Description" jsonschema:"description=Campaign description"`
}

// ContextData is the base template data shape shared by all contexts.
// Branding and Vars are excluded from JSON schema reflection since they are runtime-only
type ContextData struct {
	// CompanyName is the display name of the sending company
	CompanyName string `json:"CompanyName" jsonschema:"required,description=Company display name"`
	// CompanyAddress is the mailing address of the company
	CompanyAddress string `json:"CompanyAddress" jsonschema:"description=Company mailing address"`
	// Corporation is the legal corporation name
	Corporation string `json:"Corporation" jsonschema:"description=Legal corporation name"`
	// Year is the current year for copyright notices
	Year int `json:"Year" jsonschema:"description=Current year for copyright notices"`
	// FromEmail is the sender email address
	FromEmail string `json:"FromEmail" jsonschema:"required,description=Sender email address"`
	// SupportEmail is the support contact email address
	SupportEmail string `json:"SupportEmail" jsonschema:"description=Support contact email address"`
	// LogoURL is the company logo URL
	LogoURL string `json:"LogoURL" jsonschema:"description=Company logo URL"`
	// RootURL is the root application URL
	RootURL string `json:"RootURL" jsonschema:"description=Root application URL"`
	// ProductURL is the product home URL
	ProductURL string `json:"ProductURL" jsonschema:"description=Product home URL"`
	// DocsURL is the documentation URL
	DocsURL string `json:"DocsURL" jsonschema:"description=Documentation URL"`
	// Recipient holds per-recipient data injected at send time
	Recipient RecipientData `json:"Recipient" jsonschema:"required,description=Recipient information"`
	// Branding holds email branding data when an EmailBranding record is present
	Branding *BrandingData `json:"Branding,omitempty" jsonschema:"description=Email branding data"`
	// Vars holds user-defined template variables from defaults and per-send overrides
	Vars map[string]any `json:"-" jsonschema:"-"`
}

// Var returns a string variable from the user-defined Vars map
func (d *ContextData) Var(key string) string {
	v, _ := d.Vars[key].(string)
	return v
}

// ToTemplateData converts the typed struct into a flat map for Go template execution.
// Struct fields form the base, Vars are merged on top (highest precedence),
// and Branding is injected as a nested struct when present
func (d *ContextData) ToTemplateData() (map[string]any, error) {
	year := d.Year
	if year == 0 {
		year = time.Now().Year()
	}

	data, err := jsonx.ToMap(d)
	if err != nil {
		return nil, err
	}

	data["Year"] = year

	// Vars have highest precedence
	maps.Copy(data, d.Vars)

	if d.Branding != nil {
		data["Branding"] = d.Branding
	}

	return data, nil
}

// CampaignData holds the template data for the CampaignRecipient context
type CampaignData struct {
	ContextData
	// Campaign holds campaign metadata available in campaign recipient templates
	Campaign CampaignMeta `json:"Campaign" jsonschema:"description=Campaign metadata"`
}

// ToTemplateData converts campaign data into a flat map for Go template execution.
// Extends ContextData.ToTemplateData with the Campaign struct
func (d *CampaignData) ToTemplateData() (map[string]any, error) {
	data, err := d.ContextData.ToTemplateData()
	if err != nil {
		return nil, err
	}

	data["Campaign"] = d.Campaign

	return data, nil
}

// baseReservedFields is the set of top-level template variable names injected by the
// system for all template contexts, derived from ContextData's JSON schema properties
var baseReservedFields = providerkit.PropertyNames[ContextData]()

// contextCatalog is the authoritative list of template data contexts.
// JSON schemas are reflected from typed data structs once at init time
var contextCatalog = []models.TemplateContextEntry{
	{
		Context:        enums.TemplateContextCampaignRecipient,
		Label:          "Campaign Recipient",
		Description:    "Used for campaign emails sent to individual recipients. Data is built per-recipient from the campaign target list.",
		Schema:         providerkit.SchemaFrom[CampaignData](),
		ReservedFields: providerkit.PropertyNames[CampaignData](),
	},
	{
		Context:        enums.TemplateContextTransactional,
		Label:          "Transactional",
		Description:    "Used for system-triggered transactional emails such as email verification, password reset, invites, and billing notifications.",
		Schema:         providerkit.SchemaFrom[ContextData](),
		ReservedFields: baseReservedFields,
	},
	{
		Context:        enums.TemplateContextWorkflowAction,
		Label:          "Workflow Action",
		Description:    "Used for emails triggered by workflow action execution. The workflow definition may inject additional variables beyond the base set.",
		Schema:         providerkit.SchemaFrom[ContextData](),
		ReservedFields: baseReservedFields,
	},
}

// TemplateContextSchema returns the reflected JSON Schema for the given TemplateContext,
// suitable for storing in the jsonconfig field of a template record
func TemplateContextSchema(ctx enums.TemplateContext) json.RawMessage {
	entry, ok := lo.Find(contextCatalog, func(e models.TemplateContextEntry) bool {
		return e.Context == ctx
	})
	if !ok {
		return nil
	}

	return entry.Schema
}

// TemplateContextEntries returns all registered context entries
func TemplateContextEntries() []models.TemplateContextEntry {
	return contextCatalog
}

// reservedFieldNames is the union of all reserved field names across every template context,
// derived from the catalog entries at init time
var reservedFieldNames = func() map[string]struct{} {
	names := make(map[string]struct{})

	for _, entry := range contextCatalog {
		for _, f := range entry.ReservedFields {
			names[f] = struct{}{}
		}
	}

	return names
}()

// ReservedFieldNames returns the set of top-level template variable names that are
// injected by the system at render time and should be excluded from jsonconfig
func ReservedFieldNames() map[string]struct{} {
	return reservedFieldNames
}
