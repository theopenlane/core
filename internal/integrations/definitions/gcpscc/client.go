package gcpscc

import (
	"context"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/jsonx"
)

// Client builds GCP Security Command Center clients for one installation
type Client struct{}

// Build constructs the GCP Security Command Center client for one installation
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	meta, err := metadataFromCredential(req.Credential)
	if err != nil {
		return nil, err
	}

	clientOpts, err := clientOptions(ctx, meta, req.Credential.OAuthAccessToken)
	if err != nil {
		return nil, err
	}

	opts := append([]option.ClientOption{}, clientOpts...)
	if meta.ProjectID != "" {
		opts = append(opts, option.WithQuotaProject(meta.ProjectID))
	}

	client, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return nil, ErrSecurityCenterClientCreate
	}

	return client, nil
}

// FromAny casts a registered client instance to the SCC client type
func (Client) FromAny(value any) (*cloudscc.Client, error) {
	c, ok := value.(*cloudscc.Client)
	if !ok {
		return nil, ErrClientType
	}

	return c, nil
}

// metadataFromCredential decodes SCC credential metadata from the credential set
func metadataFromCredential(credential types.CredentialSet) (credentialMetadata, error) {
	if len(credential.ProviderData) == 0 {
		return credentialMetadata{}, ErrCredentialMetadataRequired
	}

	var meta credentialMetadata
	if err := jsonx.UnmarshalIfPresent(credential.ProviderData, &meta); err != nil {
		return credentialMetadata{}, ErrMetadataDecode
	}

	return meta.applyDefaults(), nil
}

// clientOptions builds client options based on available credentials
func clientOptions(ctx context.Context, meta credentialMetadata, oauthToken string) ([]option.ClientOption, error) {
	if meta.ServiceAccountKey != "" {
		creds, err := serviceAccountCredentials(ctx, meta.ServiceAccountKey, meta.Scopes)
		if err != nil {
			return nil, err
		}

		return []option.ClientOption{option.WithCredentials(creds)}, nil
	}

	if oauthToken != "" {
		token := &oauth2.Token{AccessToken: oauthToken}

		return []option.ClientOption{option.WithTokenSource(oauth2.StaticTokenSource(token))}, nil
	}

	return nil, ErrAccessTokenMissing
}

// serviceAccountCredentials parses and validates a service account key
func serviceAccountCredentials(ctx context.Context, rawKey string, scopes []string) (*google.Credentials, error) {
	key := normalizeServiceAccountKey(rawKey)
	if key == "" {
		return nil, ErrServiceAccountKeyInvalid
	}

	scopeList := scopes
	if len(scopeList) == 0 {
		scopeList = []string{defaultScope}
	}

	creds, err := google.CredentialsFromJSONWithType(ctx, []byte(key), google.ServiceAccount, scopeList...)
	if err != nil {
		return nil, ErrServiceAccountKeyInvalid
	}

	return creds, nil
}
