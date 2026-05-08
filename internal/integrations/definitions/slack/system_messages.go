package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

// joinStrings is a template helper that joins a string slice with the given separator
var templateFuncs = template.FuncMap{
	"joinStrings": strings.Join,
}

// newSystemTemplate parses an inline text/template with the shared helper funcs
func newSystemTemplate(name, src string) *template.Template {
	return template.Must(template.New(name).Funcs(templateFuncs).Parse(src))
}

// NewSubscriberMessage is the input for the new subscriber Slack notification
type NewSubscriberMessage struct {
	// Email is the subscriber email address
	Email string `json:"email" jsonschema:"required,description=Subscriber email address"`
}

// NewUserMessage is the input for the new user Slack notification
type NewUserMessage struct {
	// Email is the new user email address
	Email string `json:"email" jsonschema:"required,description=New user email address"`
}

// IntegrationInstalledMessage is the input for the integration installed Slack notification
type IntegrationInstalledMessage struct {
	// IntegrationName is the display name of the installed integration definition
	IntegrationName string `json:"integrationName" jsonschema:"required,description=Display name of the installed integration"`
	// OrganizationName is the Openlane organization display name
	OrganizationName string `json:"organizationName" jsonschema:"required,description=Openlane organization display name"`
	// OrganizationID is the Openlane organization identifier
	OrganizationID string `json:"organizationId" jsonschema:"required,description=Openlane organization id"`
}

// DemoRequestMessage is the input for the demo request Slack notification
type DemoRequestMessage struct {
	// CompanyName is the requesting company name
	CompanyName string `json:"companyName" jsonschema:"required,description=Requesting company name"`
	// Email is the requester email address
	Email string `json:"email" jsonschema:"required,description=Requester email address"`
	// Domains lists any additional domains associated with the request
	Domains []string `json:"domains,omitempty" jsonschema:"description=Additional domains associated with the request"`
	// CompanyDetails is a map of company attributes surfaced in the notification
	CompanyDetails map[string]any `json:"companyDetails,omitempty" jsonschema:"description=Company attributes surfaced in the notification"`
	// UserDetails is a map of user attributes surfaced in the notification
	UserDetails map[string]any `json:"userDetails,omitempty" jsonschema:"description=User attributes surfaced in the notification"`
	// Compliance is a map of compliance interests surfaced in the notification
	Compliance map[string]any `json:"compliance,omitempty" jsonschema:"description=Compliance interests surfaced in the notification"`
	// DemoRequested marks whether the requester asked for a personalized demo
	DemoRequested bool `json:"demoRequested,omitempty" jsonschema:"description=Requester asked for a personalized demo"`
}

// System message operation schemas and refs
var (
	newSubscriberSchema, NewSubscriberOp               = providerkit.OperationSchema[NewSubscriberMessage]()        //nolint:revive
	newUserSchema, NewUserOp                           = providerkit.OperationSchema[NewUserMessage]()              //nolint:revive
	integrationInstalledSchema, IntegrationInstalledOp = providerkit.OperationSchema[IntegrationInstalledMessage]() //nolint:revive
	demoRequestSchema, DemoRequestOp                   = providerkit.OperationSchema[DemoRequestMessage]()          //nolint:revive
)

// Inline system message templates
var (
	newSubscriberTemplate = newSystemTemplate("new_subscriber",
		`New waitlist subscriber: {{ .Email }}`)

	newUserTemplate = newSystemTemplate("new_user",
		`New user registered: {{ .Email }}

This message was sent using the integrations runtime framework.`)

	integrationInstalledTemplate = newSystemTemplate("integration_installed",
		`Integration installed: {{ .IntegrationName }}
Organization: {{ .OrganizationName }} ({{ .OrganizationID }})`)

	demoRequestTemplate = newSystemTemplate("demo_request",
		`New Company: {{ .CompanyName }}
Email: {{ .Email }}
{{- if .Domains }}
Domains: {{ joinStrings .Domains ", " }}
{{- end }}
{{- if .CompanyDetails }}
Company Details:
{{ range $k, $v := .CompanyDetails }}- {{ $k }}: {{ $v }}
{{ end }}
{{- end }}
{{- if .UserDetails }}
User Details:
{{ range $k, $v := .UserDetails }}- {{ $k }}: {{ $v }}
{{ end }}
{{- end }}
{{- if .Compliance }}
Compliance:
{{ range $k, $v := .Compliance }}- {{ $k }}: {{ $v }}
{{ end }}
{{- end }}
{{- if .DemoRequested }}

*Demo requested - user would like a personalized demo. Reach out to them at {{ .Email }}*
{{- end }}`)
)

// systemMessageRegistration builds an OperationRegistration for a fire-and-forget Slack system
// message: the input is rendered through tmpl and posted via the SlackClient's active transport
func systemMessageRegistration[T any](op types.OperationRef[T], schema json.RawMessage, description string, tmpl *template.Template) types.OperationRegistration {
	return types.OperationRegistration{
		Name:               op.Name(),
		Description:        description,
		Topic:              DefinitionID.OperationTopic(op.Name()),
		ClientRef:          slackClient.ID(),
		ConfigSchema:       schema,
		CustomerSelectable: lo.ToPtr(false),
		Policy:             types.ExecutionPolicy{SkipRunRecord: true},
		Handle: providerkit.WithClientConfig(slackClient, op, ErrOperationConfigInvalid,
			func(ctx context.Context, c *SlackClient, cfg T) (json.RawMessage, error) {
				return nil, renderAndSendSystemMessage(ctx, c, tmpl, cfg)
			},
		),
	}
}

// renderAndSendSystemMessage executes tmpl against input and posts the result through c's transport
func renderAndSendSystemMessage[T any](ctx context.Context, c *SlackClient, tmpl *template.Template, input T) error {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, input); err != nil {
		return fmt.Errorf("%w: %w", ErrTemplateRenderFailed, err)
	}

	return c.sendText(ctx, buf.String(), "")
}

// AllSlackSystemMessages returns all system Slack message operation registrations for wiring into the builder
func AllSlackSystemMessages() []types.OperationRegistration {
	return []types.OperationRegistration{
		systemMessageRegistration(NewSubscriberOp, newSubscriberSchema, "Notify the platform Slack workspace that a new waitlist subscriber signed up", newSubscriberTemplate),
		systemMessageRegistration(NewUserOp, newUserSchema, "Notify the platform Slack workspace that a new user registered", newUserTemplate),
		systemMessageRegistration(IntegrationInstalledOp, integrationInstalledSchema, "Notify the platform Slack workspace that an integration was installed", integrationInstalledTemplate),
		systemMessageRegistration(DemoRequestOp, demoRequestSchema, "Notify the platform Slack workspace of an inbound demo request", demoRequestTemplate),
	}
}
