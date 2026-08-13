package gcpscc

import (
	"context"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/internal/integrations/types"
)

// defaultScope is the GCP OAuth scope requested for every SCC credential
const defaultScope = "https://www.googleapis.com/auth/cloud-platform"

// Client builds GCP Security Command Center clients for one installation
type Client struct{}

// Build constructs the GCP Security Command Center client for one installation
func (Client) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	scope, err := resolveScope(req.Credentials)
	if err != nil {
		return nil, err
	}

	clientOpts, err := clientOptions(ctx, req)
	if err != nil {
		return nil, err
	}

	opts := append([]option.ClientOption{}, clientOpts...)
	if scope.ProjectID != "" {
		opts = append(opts, option.WithQuotaProject(scope.ProjectID))
	}

	client, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return nil, ErrSecurityCenterClientCreate
	}

	return client, nil
}

// resolveCredential decodes SCC service account credential metadata from the credential bindings
func resolveCredential(bindings types.CredentialBindings) (CredentialSchema, error) {
	cred, ok, err := sccCredential.Resolve(bindings)
	if err != nil {
		return CredentialSchema{}, ErrMetadataDecode
	}

	if !ok {
		return CredentialSchema{}, ErrCredentialMetadataRequired
	}

	return cred, nil
}

// resolveScope decodes the collection scope from whichever credential slot the installation bound
func resolveScope(bindings types.CredentialBindings) (CollectionScope, error) {
	federated, ok, err := workloadIdentityCredential.Resolve(bindings)
	if err != nil {
		return CollectionScope{}, ErrMetadataDecode
	}

	if ok {
		return federated.CollectionScope, nil
	}

	meta, err := resolveCredential(bindings)
	if err != nil {
		return CollectionScope{}, err
	}

	return meta.CollectionScope, nil
}

// clientOptions builds client options for whichever credential slot the installation bound
func clientOptions(ctx context.Context, req types.ClientBuildRequest) ([]option.ClientOption, error) {
	if _, ok := req.Credentials.Resolve(workloadIdentityCredential.ID()); ok {
		return workloadIdentityOptions(ctx, req)
	}

	meta, err := resolveCredential(req.Credentials)
	if err != nil {
		return nil, err
	}

	if meta.ServiceAccountKey == "" {
		return nil, ErrServiceAccountKeyInvalid
	}

	creds, err := serviceAccountCredentials(ctx, meta.ServiceAccountKey)
	if err != nil {
		return nil, err
	}

	return []option.ClientOption{option.WithCredentials(creds)}, nil
}

// serviceAccountCredentials parses and validates a service account key
func serviceAccountCredentials(ctx context.Context, rawKey string) (*google.Credentials, error) {
	key := normalizeServiceAccountKey(rawKey)
	if key == "" {
		return nil, ErrServiceAccountKeyInvalid
	}

	creds, err := google.CredentialsFromJSONWithType(ctx, []byte(key), google.ServiceAccount, defaultScope)
	if err != nil {
		return nil, ErrServiceAccountKeyInvalid
	}

	return creds, nil
}
