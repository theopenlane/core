package objectstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"

	"github.com/theopenlane/iam/tokens"

	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/objects/resolver"
	"github.com/theopenlane/core/v2/pkg/objects/storage"
	"github.com/theopenlane/core/v2/pkg/objects/storage/providers/gcs"
)

const (
	// googleSTSEndpoint is the Google Security Token Service RFC 8693 exchange endpoint
	googleSTSEndpoint = "https://sts.googleapis.com/v1/token"
	// workloadIdentityName is the fixed pool and provider ID customers create for Openlane federation
	workloadIdentityName = "openlane"
	// workloadIdentityAudienceFormat is the provider resource name Google STS expects as the exchange audience
	workloadIdentityAudienceFormat = "//iam.googleapis.com/projects/%s/locations/global/workloadIdentityPools/%s/providers/%s"
)

// workloadIdentityAudience builds the STS exchange audience for the pool-hosting project number
func workloadIdentityAudience(projectNumber string) string {
	return fmt.Sprintf(workloadIdentityAudienceFormat, projectNumber, workloadIdentityName, workloadIdentityName)
}

// newProvider builds the storage provider for one bucket from the provider set
func newProvider(ctx context.Context, providerType storage.ProviderType, providers storage.Providers, opts ...resolver.BuilderOption) (storage.Provider, error) {
	provider, err := resolver.NewProvider(ctx, providerType, providers, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	return provider, nil
}

// runtimeClientBuilder returns a build function that constructs the storage client for the platform-owned bucket on the runtime path
func runtimeClientBuilder() func(context.Context, json.RawMessage) (any, error) {
	return func(ctx context.Context, config json.RawMessage) (any, error) {
		var cfg RuntimeConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
		}

		providerType, _, err := cfg.enabledProvider()
		if err != nil {
			return nil, err
		}

		if !cfg.Provisioned() {
			return nil, ErrRuntimeConfigInvalid
		}

		opts, err := runtimeBuilderOptions(ctx, providerType, cfg)
		if err != nil {
			return nil, err
		}

		provider, err := newProvider(ctx, providerType, cfg.Providers, opts...)
		if err != nil {
			return nil, err
		}

		return &Client{
			Provider: provider,
			Import:   cfg.Import,
		}, nil
	}
}

// r2Client builds the R2 storage client from the installation's R2 credential
func r2Client(ctx context.Context, req types.ConnectionRequest[R2CredentialSchema]) (*Client, error) {
	cred := req.Credential

	provider, err := newProvider(ctx, storage.R2Provider, storage.Providers{R2: storage.R2Config{
		ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: cred.Bucket},
		Credentials: storage.R2Credentials{
			AccessKeyCredentials: storage.AccessKeyCredentials{AccessKeyID: cred.AccessKeyID, SecretAccessKey: cred.SecretAccessKey},
			AccountID:            cred.AccountID,
		},
	}})
	if err != nil {
		return nil, err
	}

	return &Client{Provider: provider}, nil
}

// serviceAccountClient builds the GCS storage client authenticating with the installation's service account key
func serviceAccountClient(ctx context.Context, req types.ConnectionRequest[ServiceAccountCredentialSchema]) (*Client, error) {
	cred := req.Credential

	opt, err := gcs.WithServiceAccountKey(ctx, []byte(normalizeServiceAccountKey(cred.ServiceAccountKey)))
	if err != nil {
		return nil, ErrServiceAccountKeyInvalid
	}

	return gcsClient(ctx, cred.Bucket, cred.ProjectID, opt)
}

// gcsClient builds the GCS storage client for one bucket with the supplied provider options
func gcsClient(ctx context.Context, bucket, projectID string, opts ...gcs.Option) (*Client, error) {
	provider, err := newProvider(ctx, storage.GCSProvider, storage.Providers{GCS: storage.GCSConfig{
		ProviderCommon: storage.ProviderCommon{Enabled: true, Bucket: bucket},
		ProjectID:      projectID,
	}}, resolver.WithGCSOptions(opts...))
	if err != nil {
		return nil, err
	}

	return &Client{Provider: provider}, nil
}

// normalizeServiceAccountKey trims and unwraps JSON-encoded service account keys
func normalizeServiceAccountKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	var decoded string
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return strings.TrimSpace(decoded)
	}

	return trimmed
}

// workloadIdentityClient builds the GCS storage client authenticating through workload identity federation
func workloadIdentityClient(ctx context.Context, req types.ConnectionRequest[WorkloadIdentityCredentialSchema]) (*Client, error) {
	source, err := federationSource(ctx, req.TokenManager, req.Integration.OwnerID, req.Credential)
	if err != nil {
		return nil, err
	}

	return gcsClient(ctx, req.Credential.Bucket, req.Credential.ProjectID, gcs.WithClientOptions(option.WithTokenSource(source)))
}

// federationSource builds the token source, impersonating a service account when configured
func federationSource(ctx context.Context, manager *tokens.TokenManager, organizationID string, cred WorkloadIdentityCredentialSchema) (oauth2.TokenSource, error) {
	if cred.ProjectNumber == "" {
		return nil, ErrProjectNumberRequired
	}

	federated, err := auth.FederatedTokenSource(ctx, manager, organizationID, auth.FederationSpec{
		Audience: workloadIdentityAudience(cred.ProjectNumber),
		Scopes:   []string{gcs.ReadWriteScope},
		Endpoint: googleSTSEndpoint,
	})
	if err != nil {
		return nil, err
	}

	if cred.ServiceAccountEmail == "" {
		return federated, nil
	}

	return impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: cred.ServiceAccountEmail,
		Scopes:          []string{gcs.ReadWriteScope},
	}, option.WithTokenSource(federated))
}
