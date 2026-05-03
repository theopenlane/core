package slack

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
)

//go:embed templates/*.tmpl
var systemMessageTemplatesFS embed.FS

// parseSystemTemplate parses a Slack system message template from the embedded FS by file name
func parseSystemTemplate(name string) *template.Template {
	return template.Must(template.ParseFS(systemMessageTemplatesFS, "templates/"+name))
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

// GitHubAppInstalledMessage is the input for the GitHub App installation Slack notification
type GitHubAppInstalledMessage struct {
	// GitHubOrganization is the GitHub organization name where the app was installed
	GitHubOrganization string `json:"githubOrganization" jsonschema:"required,description=GitHub organization name"`
	// GitHubAccountType is the GitHub account type (User, Organization)
	GitHubAccountType string `json:"githubAccountType,omitempty" jsonschema:"description=GitHub account type"`
	// OpenlaneOrganization is the Openlane organization display name
	OpenlaneOrganization string `json:"openlaneOrganization" jsonschema:"required,description=Openlane organization display name"`
	// OpenlaneOrganizationID is the Openlane organization identifier
	OpenlaneOrganizationID string `json:"openlaneOrganizationId,omitempty" jsonschema:"description=Openlane organization id"`
	// ShowOpenlaneOrganizationID reports whether the id should be rendered alongside the display name
	ShowOpenlaneOrganizationID bool `json:"showOpenlaneOrganizationId,omitempty" jsonschema:"description=Render the Openlane organization id alongside the display name"`
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
	CompanyDetails map[string]string `json:"companyDetails,omitempty" jsonschema:"description=Company attributes surfaced in the notification"`
	// UserDetails is a map of user attributes surfaced in the notification
	UserDetails map[string]string `json:"userDetails,omitempty" jsonschema:"description=User attributes surfaced in the notification"`
	// Compliance is a map of compliance interests surfaced in the notification
	Compliance map[string]string `json:"compliance,omitempty" jsonschema:"description=Compliance interests surfaced in the notification"`
	// DemoRequested marks whether the requester asked for a personalized demo
	DemoRequested bool `json:"demoRequested,omitempty" jsonschema:"description=Requester asked for a personalized demo"`
}

// System message operation schemas and refs
var (
	newSubscriberSchema, NewSubscriberOp         = providerkit.OperationSchema[NewSubscriberMessage]()      //nolint:revive
	newUserSchema, NewUserOp                     = providerkit.OperationSchema[NewUserMessage]()            //nolint:revive
	githubAppInstalledSchema, GitHubAppInstallOp = providerkit.OperationSchema[GitHubAppInstalledMessage]() //nolint:revive
	demoRequestSchema, DemoRequestOp             = providerkit.OperationSchema[DemoRequestMessage]()        //nolint:revive
)

// Parsed system message templates; exposed at package scope for unit-test reuse
var (
	newSubscriberTemplate      = parseSystemTemplate("new_subscriber.tmpl")
	newUserTemplate            = parseSystemTemplate("new_user.tmpl")
	githubAppInstalledTemplate = parseSystemTemplate("github_app_installation.tmpl")
	demoRequestTemplate        = parseSystemTemplate("demo_request.tmpl")
)

// systemMessageRegistration builds an OperationRegistration for a fire-and-forget Slack system
// message: the input is rendered through tmpl and posted via the SlackClient's active transport
func systemMessageRegistration[T any](op types.OperationRef[T], schema json.RawMessage, description string, tmpl *template.Template) types.OperationRegistration {
	return types.OperationRegistration{
		Name:         op.Name(),
		Description:  description,
		Topic:        DefinitionID.OperationTopic(op.Name()),
		ClientRef:    slackClient.ID(),
		ConfigSchema: schema,
		Policy:       types.ExecutionPolicy{SkipRunRecord: true},
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
		systemMessageRegistration(GitHubAppInstallOp, githubAppInstalledSchema, "Notify the platform Slack workspace that the Openlane GitHub App was installed", githubAppInstalledTemplate),
		systemMessageRegistration(DemoRequestOp, demoRequestSchema, "Notify the platform Slack workspace of an inbound demo request", demoRequestTemplate),
	}
}
