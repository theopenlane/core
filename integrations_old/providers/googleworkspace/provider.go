package googleworkspace

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/samber/lo"
	"golang.org/x/oauth2"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/internal/integrations/providers/oauth"
	"github.com/theopenlane/core/internal/integrations/spec"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// TypeGoogleWorkspace identifies the Google Workspace provider
const TypeGoogleWorkspace = types.ProviderType("googleworkspace")

// googleWorkspaceCredentialsSchema is the JSON Schema for Google Workspace tenant credentials.
var googleWorkspaceCredentialsSchema = []byte(`{"type":"object","additionalProperties":false,"properties":{"alias":{"type":"string","title":"Credential Alias","description":"Friendly identifier for this Workspace tenant."},"adminEmail":{"type":"string","title":"Admin Email","description":"Workspace administrator to impersonate when performing directory sync."},"customerId":{"type":"string","title":"Workspace Customer ID","description":"Optional customer ID used to scope Admin SDK queries."},"subjectUser":{"type":"string","title":"Subject User","description":"Optional subject user for domain-wide delegation scenarios."},"organizationalUnitPath":{"type":"string","title":"Organizational Unit Path","description":"Restrict sync operations to a specific OU path."},"includeSuspendedUsers":{"type":"boolean","title":"Include Suspended Users","description":"Toggle to include suspended accounts in compliance exports.","default":false},"syncInterval":{"type":"string","title":"Sync Interval","description":"Optional duration string that hints how frequently users should be synchronized."},"enableGroupSync":{"type":"boolean","title":"Sync Groups","description":"Enable group synchronization in addition to users.","default":true}}}`)

const (
	// ClientGoogleWorkspaceAPI identifies the Google Workspace HTTP API client
	ClientGoogleWorkspaceAPI types.ClientName = "api"
)

const (
	googleWorkspaceHealthOp           types.OperationName = types.OperationHealthDefault
	googleWorkspaceDirectorySyncOp    types.OperationName = types.OperationDirectorySync
	googleWorkspaceDirectoryAlertType string              = "directory_account"
	googleWorkspaceDirectoryFields    string              = "nextPageToken,users(id,primaryEmail,name/fullName,name/givenName,name/familyName,thumbnailPhotoUrl,organizations/title,organizations/department,orgUnitPath,suspended,archived,isEnforcedIn2Sv,isEnrolledIn2Sv,lastLoginTime)"

	googleWorkspaceDirectoryDefaultPageSize = int64(200)
	googleWorkspaceHealthMaxResults         = int64(1)
)

// oauthProvider wraps oauth.Provider and implements types.MappingProvider
type oauthProvider struct {
	*oauth.Provider
}

// DefaultMappings returns the built-in ingest mapping registrations for Google Workspace
func (p *oauthProvider) DefaultMappings() []types.MappingRegistration {
	return googleWorkspaceDefaultMappings()
}

// Builder returns the Google Workspace provider builder with the supplied operator config applied.
func Builder(cfg Config) providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGoogleWorkspace,
		SpecFunc:     googleWorkspaceSpec,
		BuildFunc: func(_ context.Context, s spec.ProviderSpec) (types.Provider, error) {
			if s.OAuth != nil && cfg.ClientID != "" {
				s.OAuth.ClientID = cfg.ClientID
				s.OAuth.ClientSecret = cfg.ClientSecret
			}

			base, err := oauth.New(s,
				oauth.WithOperations(googleWorkspaceOperations()),
				oauth.WithClientDescriptors(googleWorkspaceClientDescriptors()),
			)
			if err != nil {
				return nil, err
			}

			return &oauthProvider{Provider: base}, nil
		},
	}
}

// googleWorkspaceSpec returns the static provider specification for the Google Workspace provider.
func googleWorkspaceSpec() spec.ProviderSpec {
	return spec.ProviderSpec{
		Name:             "googleworkspace",
		DisplayName:      "Google Workspace",
		Category:         "identity",
		AuthType:         types.AuthKindOAuth2,
		AuthStartPath:    "/v1/integrations/oauth/start",
		AuthCallbackPath: "/v1/integrations/oauth/callback",
		Active:           lo.ToPtr(true),
		Visible:          lo.ToPtr(true),
		LogoURL:          "",
		DocsURL:          "https://docs.theopenlane.io/docs/platform/integrations/google_workspace/overview",
		OAuth: &spec.OAuthSpec{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
			Scopes: []string{
				"https://www.googleapis.com/auth/admin.directory.user.readonly",
				"https://www.googleapis.com/auth/admin.directory.group.readonly",
				"https://www.googleapis.com/auth/apps.groups.migration",
			},
			AuthParams: map[string]string{
				"access_type": "offline",
				"prompt":      "consent",
			},
			RedirectURI: "https://api.theopenlane.io/v1/integrations/oauth/callback",
		},
		UserInfo: &spec.UserInfoSpec{
			URL:       "https://openidconnect.googleapis.com/v1/userinfo",
			Method:    "GET",
			AuthStyle: "Bearer",
			IDPath:    "sub",
			EmailPath: "email",
			LoginPath: "name",
		},
		Persistence: &spec.PersistenceSpec{
			StoreRefreshToken: true,
		},
		Labels: map[string]string{
			"vendor":  "google",
			"product": "workspace",
		},
		CredentialsSchema: googleWorkspaceCredentialsSchema,
		Description:       "Collect Google Workspace directory and identity metadata to support account hygiene and compliance posture checks.",
	}
}

// googleWorkspaceClientDescriptors returns the client descriptors published by Google Workspace
func googleWorkspaceClientDescriptors() []types.ClientDescriptor {
	return providerkit.DefaultClientDescriptors(TypeGoogleWorkspace, ClientGoogleWorkspaceAPI, "Google Workspace REST API client", buildGoogleWorkspaceClient)
}

// buildGoogleWorkspaceClient constructs an Admin SDK client from a credential set
func buildGoogleWorkspaceClient(ctx context.Context, credential types.CredentialSet, _ json.RawMessage) (types.ClientInstance, error) {
	token := credential.OAuthAccessToken
	if token == "" {
		return types.EmptyClientInstance(), providerkit.ErrOAuthTokenMissing
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	svc, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return types.EmptyClientInstance(), err
	}

	return types.NewClientInstance(svc), nil
}

// resolveGoogleWorkspaceClient returns a pooled Admin SDK client or builds one from the credential set
func resolveGoogleWorkspaceClient(ctx context.Context, input types.OperationInput) (*admin.Service, error) {
	if c, ok := types.ClientInstanceAs[*admin.Service](input.Client); ok {
		return c, nil
	}

	instance, err := buildGoogleWorkspaceClient(ctx, input.Credential, json.RawMessage(nil))
	if err != nil {
		return nil, err
	}

	client, ok := types.ClientInstanceAs[*admin.Service](instance)
	if !ok || client == nil {
		return nil, errGoogleWorkspaceAdminServiceClientBuild
	}

	return client, nil
}

type googleWorkspaceDirectoryConfig struct {
	// Customer overrides the default customer selector
	Customer string `json:"customer"`
	// Domain scopes the user listing to a domain
	Domain string `json:"domain"`
	// Query applies an Admin SDK user query filter
	Query string `json:"query"`
}

// googleWorkspaceHealthDetails captures health check details from the Admin SDK users list call
type googleWorkspaceHealthDetails struct {
	// UserCount is the number of users returned by the health probe
	UserCount int `json:"userCount"`
}

// googleWorkspaceDirectorySyncDetails captures directory sync operation details
type googleWorkspaceDirectorySyncDetails struct {
	// UsersTotal is the number of users returned by Google APIs
	UsersTotal int `json:"users_total"`
	// AlertsTotal is the number of emitted ingest envelopes
	AlertsTotal int `json:"alerts_total"`
	// Alerts is the emitted ingest envelopes
	Alerts []types.AlertEnvelope `json:"alerts"`
}

// googleWorkspaceUserPayload wraps an Admin SDK user with an observed timestamp for envelope encoding
type googleWorkspaceUserPayload struct {
	*admin.User
	ObservedAt string `json:"observedAt"`
}

// googleWorkspaceOperations returns the Google Workspace operations supported by this provider
func googleWorkspaceOperations() []types.OperationDescriptor {
	return []types.OperationDescriptor{
		providerkit.HealthOperation(googleWorkspaceHealthOp, "Call Google Admin SDK users.list to verify the workspace token", ClientGoogleWorkspaceAPI, runGoogleWorkspaceHealth),
		{
			Name:        googleWorkspaceDirectorySyncOp,
			Kind:        types.OperationKindCollectFindings,
			Description: "Collect Google Workspace directory users and emit directory account envelopes",
			Client:      ClientGoogleWorkspaceAPI,
			Run:         runGoogleWorkspaceDirectorySync,
			Ingest: []types.IngestContract{
				{
					Schema: mappingSchemaDirectoryAccount,
				},
			},
		},
	}
}

// runGoogleWorkspaceHealth verifies Admin SDK reachability with a small users.list probe
func runGoogleWorkspaceHealth(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	svc, err := resolveGoogleWorkspaceClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	resp, err := svc.Users.List().
		Customer("my_customer").
		MaxResults(googleWorkspaceHealthMaxResults).
		Projection("basic").
		ViewType("admin_view").
		Fields(googleapi.Field("users(id),nextPageToken")).
		Do()
	if err != nil {
		return providerkit.OperationFailure("Google Workspace health check failed", err, nil)
	}

	return providerkit.OperationSuccess(
		fmt.Sprintf("Google Workspace directory API reachable (%d sample users)", len(resp.Users)),
		googleWorkspaceHealthDetails{UserCount: len(resp.Users)},
	), nil
}

// runGoogleWorkspaceDirectorySync collects directory users and emits directory account envelopes
func runGoogleWorkspaceDirectorySync(ctx context.Context, input types.OperationInput) (types.OperationResult, error) {
	svc, err := resolveGoogleWorkspaceClient(ctx, input)
	if err != nil {
		return types.OperationResult{}, err
	}

	var config googleWorkspaceDirectoryConfig
	if err := jsonx.UnmarshalIfPresent(input.Config, &config); err != nil {
		return types.OperationResult{}, err
	}

	customer := "my_customer"
	if config.Customer != "" {
		customer = config.Customer
	}

	fetch := func(_ context.Context, pageToken string) (providerkit.PageResult[*admin.User], error) {
		call := svc.Users.List().
			MaxResults(googleWorkspaceDirectoryDefaultPageSize).
			Projection("full").
			ViewType("admin_view").
			Fields(googleapi.Field(googleWorkspaceDirectoryFields))

		if config.Domain != "" {
			call = call.Domain(config.Domain)
		} else {
			call = call.Customer(customer)
		}

		if config.Query != "" {
			call = call.Query(config.Query)
		}

		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return providerkit.PageResult[*admin.User]{}, err
		}

		return providerkit.PageResult[*admin.User]{
			Items:     resp.Users,
			NextToken: resp.NextPageToken,
		}, nil
	}

	allUsers, err := providerkit.CollectAll(ctx, fetch, 0)
	if err != nil {
		return providerkit.OperationFailure("Directory users fetch failed", err, nil)
	}

	observedAt := time.Now().UTC().Format(time.RFC3339)
	envelopes := make([]types.AlertEnvelope, 0, len(allUsers))

	for _, u := range allUsers {
		payload, err := json.Marshal(googleWorkspaceUserPayload{User: u, ObservedAt: observedAt})
		if err != nil {
			return providerkit.OperationFailure("Directory user payload encoding failed", err, nil)
		}

		envelopes = append(envelopes, types.AlertEnvelope{
			AlertType: googleWorkspaceDirectoryAlertType,
			Resource:  googleWorkspaceDirectoryResource(u),
			Payload:   payload,
		})
	}

	return providerkit.OperationSuccess(fmt.Sprintf("Collected %d directory accounts", len(envelopes)), googleWorkspaceDirectorySyncDetails{
		UsersTotal:  len(allUsers),
		AlertsTotal: len(envelopes),
		Alerts:      envelopes,
	}), nil
}

// googleWorkspaceDirectoryResource selects the envelope resource identity for a user
func googleWorkspaceDirectoryResource(u *admin.User) string {
	if u.PrimaryEmail != "" {
		return u.PrimaryEmail
	}

	return u.Id
}
