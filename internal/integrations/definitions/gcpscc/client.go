package gcpscc

import (
	"context"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// defaultScope is the GCP OAuth scope requested for every SCC credential
const defaultScope = "https://www.googleapis.com/auth/cloud-platform"

// workloadIdentityClient builds the SCC client authenticating through workload identity federation
func workloadIdentityClient(ctx context.Context, req types.ConnectionRequest[WorkloadIdentityCredentialSchema]) (Client, error) {
	source, err := federationSource(ctx, req.TokenManager, req.Integration.OwnerID, req.Credential)
	if err != nil {
		return Client{}, err
	}

	return newClient(ctx, req.Credential.CollectionScope, option.WithTokenSource(source))
}

// serviceAccountClient builds the SCC client authenticating with a service account key
func serviceAccountClient(ctx context.Context, req types.ConnectionRequest[CredentialSchema]) (Client, error) {
	creds, err := serviceAccountCredentials(ctx, req.Credential.ServiceAccountKey)
	if err != nil {
		return Client{}, err
	}

	return newClient(ctx, req.Credential.CollectionScope, option.WithCredentials(creds))
}

// newClient creates the SCC client for a collection scope, setting the quota project when the scope names one
func newClient(ctx context.Context, scope CollectionScope, opts ...option.ClientOption) (Client, error) {
	if scope.ProjectID != "" {
		opts = append(opts, option.WithQuotaProject(scope.ProjectID))
	}

	scc, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return Client{}, ErrSecurityCenterClientCreate
	}

	return Client{Client: scc, Scope: scope}, nil
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
