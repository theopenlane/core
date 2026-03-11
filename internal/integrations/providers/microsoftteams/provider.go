package microsoftteams

import (
	"context"
	"fmt"
	"net/url"

	"github.com/samber/lo"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TypeMicrosoftTeams identifies the Microsoft Teams provider
const TypeMicrosoftTeams = types.ProviderType("microsoftteams")

const (
	// ClientMicrosoftTeamsAPI identifies the Microsoft Graph API client
	ClientMicrosoftTeamsAPI types.ClientName = "api"
)

const teamsGraphBaseURL = "https://graph.microsoft.com/v1.0/"

const (
	teamsHealthOp      types.OperationName = types.OperationHealthDefault
	teamsChannelsOp    types.OperationName = "teams.sample"
	teamsMessageSendOp types.OperationName = "message.send"
)

// teamsCredentialsSchema is the JSON Schema for Microsoft Teams tenant credentials.
var teamsCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"required":["tenantId"],"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly name for this Microsoft 365 tenant."},"tenantId":{"type":"string","title":"Tenant ID","description":"Azure AD tenant hosting the Teams instance."}}}`)

// Builder returns the Microsoft Teams provider builder with the supplied operator config applied.
func Builder(cfg Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeMicrosoftTeams,
		SpecFunc:     microsoftTeamsSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if s.OAuth != nil && cfg.ClientID != "" {
				s.OAuth.ClientID = cfg.ClientID
				s.OAuth.ClientSecret = cfg.ClientSecret
			}

			return oauth.New(s, oauth.WithOperations(teamsOperations()), oauth.WithClientDescriptors(teamsClientDescriptors()))
		},
	}
}

// microsoftTeamsSpec returns the static provider specification for the Microsoft Teams provider.
func microsoftTeamsSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:             "microsoftteams",
		DisplayName:      "Microsoft Teams",
		Category:         "collab",
		AuthType:         types.AuthKindOAuth2,
		AuthStartPath:    "/v1/integrations/oauth/start",
		AuthCallbackPath: "/v1/integrations/oauth/callback",
		Active:           lo.ToPtr(false),
		Visible:          lo.ToPtr(true),
		LogoURL:          "",
		DocsURL:          "https://docs.theopenlane.io/docs/platform/integrations/microsoft_teams/overview",
		OAuth: &spec.OAuthSpec{
			AuthURL:  "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL: "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			Scopes: []string{
				"https://graph.microsoft.com/User.Read",
				"https://graph.microsoft.com/Team.ReadBasic.All",
				"https://graph.microsoft.com/Channel.ReadBasic.All",
				"https://graph.microsoft.com/ChannelMessage.Send",
				"offline_access",
			},
			RedirectURI: "https://api.theopenlane.io/v1/integrations/oauth/callback",
		},
		UserInfo: &spec.UserInfoSpec{
			URL:       "https://graph.microsoft.com/v1.0/me",
			Method:    "GET",
			AuthStyle: "Bearer",
			IDPath:    "id",
			EmailPath: "mail",
			LoginPath: "displayName",
		},
		Persistence: &spec.PersistenceSpec{
			StoreRefreshToken: true,
		},
		Labels: map[string]string{
			"vendor":  "microsoft",
			"product": "teams",
		},
		CredentialsSchema: teamsCredentialsSchema,
		Description:       "Integrate with Microsoft Teams to collect collaboration metadata and send notification messages through Microsoft Graph.",
	}
}

// teamsClientDescriptors returns the client descriptors published by Microsoft Teams
func teamsClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeMicrosoftTeams, ClientMicrosoftTeamsAPI, "Microsoft Graph API client", providerkit.TokenClientBuilder(providerkit.OAuthTokenFromCredential, nil))
}

type teamsMessageOperationConfig struct {
	// TeamID identifies the Team to post into
	TeamID string `json:"team_id" jsonschema:"required,description=Microsoft Teams team ID to receive the message."`
	// ChannelID identifies the channel within the team
	ChannelID string `json:"channel_id" jsonschema:"required,description=Microsoft Teams channel ID to receive the message."`
	// Body is the message body content
	Body string `json:"body" jsonschema:"required,description=Message body content."`
	// BodyFormat is the message format (text or html)
	BodyFormat string `json:"body_format,omitempty" jsonschema:"description=Message body format (text or html)."`
	// Subject is an optional message subject
	Subject string `json:"subject,omitempty" jsonschema:"description=Optional message subject."`
}

var teamsMessageConfigSchema = providerkit.SchemaFrom[teamsMessageOperationConfig]()

// teamsOperations returns the Microsoft Teams operations supported by this provider
func teamsOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(teamsHealthOp, "Call Graph /me to verify Teams access.", ClientMicrosoftTeamsAPI,
			providerkit.HealthCheckRunner(providerkit.OAuthTokenFromCredential, "https://graph.microsoft.com/v1.0/me", "Graph /me failed",
				func(profile teamsProfileResponse) (string, any) {
					return fmt.Sprintf("Graph token valid for %s", profile.DisplayName), teamsHealthDetails{
						ID:   profile.ID,
						Mail: profile.Mail,
					}
				})),
		{
			Name:        teamsChannelsOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect a sample of joined teams for the user context.",
			Client:      ClientMicrosoftTeamsAPI,
			Run:         runTeamsSample,
		},
		{
			Name:         teamsMessageSendOp,
			Kind:         types.OperationKindNotify,
			Description:  "Send a Teams channel message via Microsoft Graph.",
			Client:       ClientMicrosoftTeamsAPI,
			Run:          runTeamsMessageSendOperation,
			ConfigSchema: teamsMessageConfigSchema,
		},
	}
}

type teamsProfileResponse struct {
	// ID is the user identifier
	ID string `json:"id"`
	// DisplayName is the user display name
	DisplayName string `json:"displayName"`
	// Mail is the primary email address
	Mail string `json:"mail"`
}

type teamsHealthDetails struct {
	ID   string `json:"id"`
	Mail string `json:"mail"`
}

type teamsSampleEntry struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

type teamsSampleDetails struct {
	Teams []teamsSampleEntry `json:"teams"`
}

type teamsMessageSendDetails struct {
	TeamID    string `json:"teamId"`
	ChannelID string `json:"channelId"`
	MessageID string `json:"messageId"`
}

type teamsChannelMessageBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type teamsChannelMessageRequest struct {
	Body    teamsChannelMessageBody `json:"body"`
	Subject string                  `json:"subject,omitempty"`
}

// runTeamsSample collects a sample of joined Teams for the authenticated user
func runTeamsSample(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := providerkit.ResolveAuthenticatedClient(input, providerkit.OAuthTokenFromCredential, teamsGraphBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	var resp struct {
		// Value lists the joined teams
		Value []struct {
			// ID is the team identifier
			ID string `json:"id"`
			// DisplayName is the team display name
			DisplayName string `json:"displayName"`
		} `json:"value"`
	}

	if err := client.GetJSON(ctx, "me/joinedTeams?$top=5", &resp); err != nil {
		return providerkit.OperationFailure("Graph joinedTeams failed", err, nil)
	}

	samples := make([]teamsSampleEntry, 0, len(resp.Value))
	for _, team := range resp.Value {
		samples = append(samples, teamsSampleEntry{
			ID:          team.ID,
			DisplayName: team.DisplayName,
		})
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Retrieved %d joined teams", len(samples)), teamsSampleDetails{Teams: samples}), nil
}

// runTeamsMessageSendOperation posts a message to a Teams channel
func runTeamsMessageSendOperation(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	client, err := providerkit.ResolveAuthenticatedClient(input, providerkit.OAuthTokenFromCredential, teamsGraphBaseURL, nil)
	if err != nil {
		return types.OperationResult{}, err
	}

	var cfg teamsMessageOperationConfig
	if err := jsonx.UnmarshalIfPresent(input.Config, &cfg); err != nil {
		return types.OperationResult{}, err
	}

	teamID := cfg.TeamID
	channelID := cfg.ChannelID
	if teamID == "" || channelID == "" {
		return types.OperationResult{}, ErrTeamsChannelMissing
	}

	body := cfg.Body
	if body == "" {
		return types.OperationResult{}, ErrTeamsMessageEmpty
	}

	contentType := cfg.BodyFormat
	if contentType == "" {
		contentType = "text"
	}

	switch contentType {
	case "text", "html":
	default:
		return types.OperationResult{}, ErrTeamsMessageFormatInvalid
	}

	payload := teamsChannelMessageRequest{
		Body: teamsChannelMessageBody{
			ContentType: contentType,
			Content:     body,
		},
		Subject: cfg.Subject,
	}

	path := fmt.Sprintf("teams/%s/channels/%s/messages", url.PathEscape(teamID), url.PathEscape(channelID))
	var resp struct {
		ID string `json:"id"`
	}
	if err := client.PostJSON(ctx, path, payload, &resp); err != nil {
		return providerkit.OperationFailure("Graph channel message failed", err, nil)
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Teams message sent to %s", channelID), teamsMessageSendDetails{
		TeamID:    teamID,
		ChannelID: channelID,
		MessageID: resp.ID,
	}), nil
}
