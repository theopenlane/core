package slack

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/samber/lo"
	slackgo "github.com/slack-go/slack"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TypeSlack identifies the Slack provider
const TypeSlack = types.ProviderType("slack")

const (
	// ClientSlackAPI identifies the Slack HTTP API client
	ClientSlackAPI types.ClientName = "api"
)

const (
	slackOperationHealth      types.OperationName = types.OperationHealthDefault
	slackOperationTeam        types.OperationName = "team.inspect"
	slackOperationMessageSend types.OperationName = "message.send"
)

// Builder returns the Slack provider builder with the supplied operator config applied.
func Builder(cfg Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeSlack,
		SpecFunc:     slackSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if s.OAuth != nil && cfg.ClientID != "" {
				s.OAuth.ClientID = cfg.ClientID
				s.OAuth.ClientSecret = cfg.ClientSecret
			}

			return oauth.New(s,
				oauth.WithOperations(slackOperations()),
				oauth.WithClientDescriptors(slackClientDescriptors()),
			)
		},
	}
}

// slackSpec returns the static provider specification for the Slack provider.
func slackSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:             "slack",
		DisplayName:      "Slack",
		Category:         "collab",
		AuthType:         types.AuthKindOAuth2,
		AuthStartPath:    "/v1/integrations/oauth/start",
		AuthCallbackPath: "/v1/integrations/oauth/callback",
		Active:           lo.ToPtr(true),
		Visible:          lo.ToPtr(true),
		LogoURL:          "",
		DocsURL:          "https://docs.theopenlane.io/docs/platform/integrations/slack/overview",
		OAuth: &spec.OAuthSpec{
			AuthURL:     "https://slack.com/oauth/v2/authorize",
			TokenURL:    "https://slack.com/api/oauth.v2.access",
			Scopes:      []string{"chat:write", "chat:write.public", "team:read", "chat:write.customize", "users:read"},
			RedirectURI: "https://api.theopenlane.io/v1/integrations/oauth/callback",
		},
		UserInfo: &spec.UserInfoSpec{
			URL:       "https://slack.com/api/users.identity",
			Method:    "GET",
			AuthStyle: "Bearer",
			IDPath:    "user.id",
			EmailPath: "user.email",
			LoginPath: "user.name",
		},
		Persistence: &spec.PersistenceSpec{
			StoreRefreshToken: true,
		},
		Description: "Integrate with Slack to verify workspace posture and send operational or compliance notifications to channels.",
	}
}

// slackClientDescriptors returns the client descriptors published by Slack
func slackClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeSlack, ClientSlackAPI, "Slack Web API client", buildSlackClient)
}

// buildSlackClient constructs a Slack SDK client from a credential set
func buildSlackClient(_ context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token := credential.OAuthAccessToken
	if token == "" {
		return types.EmptyClientInstance(), providerkit.ErrOAuthTokenMissing
	}

	return types.NewClientInstance(slackgo.New(token)), nil
}

type slackMessageOperationConfig struct {
	// Channel identifies the Slack channel or user to receive the message
	Channel string `json:"channel" jsonschema:"required,description=Slack channel ID or user ID to receive the message."`
	// Text is the message text when blocks are not supplied
	Text string `json:"text,omitempty" jsonschema:"description=Message text (required unless blocks are supplied)."`
	// Blocks carries optional Block Kit payloads
	Blocks []json.RawMessage `json:"blocks,omitempty" jsonschema:"description=Optional Slack Block Kit payload."`
	// Attachments carries optional attachments payloads
	Attachments []json.RawMessage `json:"attachments,omitempty" jsonschema:"description=Optional attachments payload."`
	// ThreadTS identifies the thread timestamp for replies
	ThreadTS string `json:"thread_ts,omitempty" jsonschema:"description=Optional thread timestamp to reply within an existing thread."`
	// UnfurlLinks controls link unfurling in messages
	UnfurlLinks *bool `json:"unfurl_links,omitempty" jsonschema:"description=Whether to unfurl links in the message."`
	// UnfurlMedia controls media unfurling in messages
	UnfurlMedia *bool `json:"unfurl_media,omitempty" jsonschema:"description=Whether to unfurl media in the message."`
}

var slackMessageConfigSchema = providerkit.SchemaFrom[slackMessageOperationConfig]()

// slackOperations returns the Slack operations supported by this provider
func slackOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(slackOperationHealth, "Call auth.test to ensure the Slack token is valid and scoped correctly.", ClientSlackAPI, runSlackHealthOperation),
		{
			Name:        slackOperationTeam,
			Kind:        types.OperationKindScanSettings,
			Description: "Collect workspace metadata via team.info for posture analysis.",
			Client:      ClientSlackAPI,
			Run:         runSlackTeamOperation,
		},
		{
			Name:         slackOperationMessageSend,
			Kind:         types.OperationKindNotify,
			Description:  "Send a Slack message via chat.postMessage.",
			Client:       ClientSlackAPI,
			Run:          runSlackMessagePostOperation,
			ConfigSchema: slackMessageConfigSchema,
		},
	}
}

type slackHealthDetails struct {
	Team string `json:"team"`
	URL  string `json:"url"`
	User string `json:"user"`
}

type slackTeamDetails struct {
	TeamID      string `json:"teamId"`
	Name        string `json:"name"`
	Domain      string `json:"domain"`
	EmailDomain string `json:"emailDomain"`
}

type slackMessageDetails struct {
	Channel string `json:"channel"`
	TS      string `json:"ts"`
}

// resolveSlackClient returns a pooled Slack client or builds one from the credential set
func resolveSlackClient(input types.OperationInput) (*slackgo.Client, error) {
	if c, ok := types.ClientInstanceAs[*slackgo.Client](input.Client); ok {
		return c, nil
	}

	token := input.Credential.OAuthAccessToken
	if token == "" {
		return nil, providerkit.ErrOAuthTokenMissing
	}

	return slackgo.New(token), nil
}

// runSlackHealthOperation verifies the Slack OAuth token via auth.test
func runSlackHealthOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveSlackClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := client.AuthTestContext(ctx)
	if err != nil {
		return providerkit.OperationFailure("Slack auth.test failed", err, nil)
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Slack token valid for workspace %s", resp.Team), slackHealthDetails{
		Team: resp.Team,
		URL:  resp.URL,
		User: resp.User,
	}), nil
}

// runSlackTeamOperation fetches workspace metadata for posture analysis
func runSlackTeamOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveSlackClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	team, err := client.GetTeamInfoContext(ctx)
	if err != nil {
		return providerkit.OperationFailure("Slack team.info failed", err, nil)
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Workspace %s (%s) settings retrieved", team.Name, team.ID), slackTeamDetails{
		TeamID:      team.ID,
		Name:        team.Name,
		Domain:      team.Domain,
		EmailDomain: team.EmailDomain,
	}), nil
}

// runSlackMessagePostOperation sends a message to a Slack channel or user
func runSlackMessagePostOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := resolveSlackClient(input)
	if err != nil {
		return types.OperationResult{}, err
	}

	var cfg slackMessageOperationConfig
	if err := jsonx.UnmarshalIfPresent(input.Config, &cfg); err != nil {
		return types.OperationResult{}, err
	}

	channel := cfg.Channel
	if channel == "" {
		return types.OperationResult{}, ErrSlackChannelMissing
	}

	hasText := cfg.Text != ""
	hasBlocks := len(cfg.Blocks) > 0
	hasAttachments := len(cfg.Attachments) > 0

	if !hasText && !hasBlocks && !hasAttachments {
		return types.OperationResult{}, ErrSlackMessageEmpty
	}

	opts := []slackgo.MsgOption{
		slackgo.MsgOptionAsUser(true),
	}

	if hasText {
		opts = append(opts, slackgo.MsgOptionText(cfg.Text, false))
	}

	if cfg.ThreadTS != "" {
		opts = append(opts, slackgo.MsgOptionTS(cfg.ThreadTS))
	}

	if cfg.UnfurlLinks != nil && !*cfg.UnfurlLinks {
		opts = append(opts, slackgo.MsgOptionDisableLinkUnfurl())
	}

	respChannel, ts, err := client.PostMessageContext(ctx, channel, opts...)
	if err != nil {
		return providerkit.OperationFailure("Slack chat.postMessage failed", err, nil)
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Slack message sent to %s", respChannel), slackMessageDetails{
		Channel: respChannel,
		TS:      ts,
	}), nil
}
