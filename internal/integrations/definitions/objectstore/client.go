package objectstore

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samber/lo"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"

	"github.com/theopenlane/core/v2/internal/integrations/auth"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/internal/objects/resolver"
	"github.com/theopenlane/core/v2/pkg/jsonx"
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

// Client is the storage provider for one bucket with the import spec it reads
type Client struct {
	// Provider is the storage provider scoped to the bucket
	Provider storage.Provider
	// Import is the prefix and schema the record import reads unless the trigger overrides them
	Import ImportRecords
}

// providerSpec is the provider selection derived from one credential slot
type providerSpec struct {
	// provider is the storage provider type the slot maps to
	provider storage.ProviderType
	// config is the provider configuration for the installation's bucket
	config storage.ProviderConfigs
	// options carry provider builder options such as a credential token source
	options []resolver.BuilderOption
}

// clientBuilder builds storage clients for one installation, selecting the provider from the bound credential slot
type clientBuilder struct {
	// aws is the platform's AWS source identity used to assume customer roles
	aws Config
}

// Build constructs the storage client for whichever credential slot the installation bound
func (c clientBuilder) Build(ctx context.Context, req types.ClientBuildRequest) (any, error) {
	slot, ok := boundSlot(req.Credentials)
	if !ok {
		return nil, ErrCredentialMetadataRequired
	}

	var (
		spec providerSpec
		err  error
	)

	switch slot {
	case workloadIdentityCredential.ID():
		spec, err = workloadIdentitySpec(ctx, req)
	case serviceAccountCredential.ID():
		spec, err = serviceAccountSpec(ctx, req.Credentials)
	case awsAssumeRoleCredential.ID():
		spec, err = c.assumeRoleSpec(ctx, req.Credentials)
	case awsAccessKeyCredential.ID():
		spec, err = accessKeySpec(req.Credentials)
	case r2Credential.ID():
		spec, err = r2Spec(req.Credentials)
	}

	if err != nil {
		return nil, err
	}

	var input UserInput
	if req.Integration != nil {
		if err := jsonx.UnmarshalIfPresent(req.Integration.Config.ClientConfig, &input); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
		}
	}

	provider, err := resolver.NewProvider(ctx, spec.provider, spec.config, spec.options...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
	}

	return &Client{
		Provider: provider,
		Import:   input.Import.ImportRecords,
	}, nil
}

// boundSlot returns the first credential slot the installation bound in provider selection order
func boundSlot(bindings types.CredentialBindings) (types.CredentialSlotID, bool) {
	return lo.Find(credentialSlots, func(slot types.CredentialSlotID) bool {
		_, ok := bindings.Resolve(slot)

		return ok
	})
}

// decodeCredential decodes the credential bound to the slot, failing when the slot is unbound or undecodable
func decodeCredential[T any](bindings types.CredentialBindings, slot types.CredentialRef[T]) (T, error) {
	cred, ok, err := slot.Resolve(bindings)
	if err != nil {
		var zero T

		return zero, ErrMetadataDecode
	}

	if !ok {
		var zero T

		return zero, ErrCredentialMetadataRequired
	}

	return cred, nil
}

// runtimeClientBuilder returns a build function that constructs the storage client for the platform-owned
// bucket on the runtime path
func runtimeClientBuilder() func(context.Context, json.RawMessage) (any, error) {
	return func(ctx context.Context, config json.RawMessage) (any, error) {
		var cfg RuntimeConfig
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrRuntimeConfigDecode, err)
		}

		providerType, providerCfg, err := cfg.enabledProvider()
		if err != nil {
			return nil, err
		}

		if !cfg.Provisioned() {
			return nil, ErrRuntimeConfigInvalid
		}

		opts, err := runtimeBuilderOptions(ctx, providerType, providerCfg)
		if err != nil {
			return nil, err
		}

		provider, err := resolver.NewProvider(ctx, providerType, providerCfg, opts...)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrClientCreate, err)
		}

		return &Client{
			Provider: provider,
			Import:   cfg.Import,
		}, nil
	}
}

// r2Spec selects R2 with the installation's R2 credential
func r2Spec(bindings types.CredentialBindings) (providerSpec, error) {
	cred, err := decodeCredential(bindings, r2Credential)
	if err != nil {
		return providerSpec{}, err
	}

	return providerSpec{
		provider: storage.R2Provider,
		config: storage.ProviderConfigs{
			Enabled: true,
			Bucket:  cred.Bucket,
			Credentials: storage.ProviderCredentials{
				AccountID:       cred.AccountID,
				AccessKeyID:     cred.AccessKeyID,
				SecretAccessKey: cred.SecretAccessKey,
			},
		},
	}, nil
}

// serviceAccountSpec selects GCS authenticating with the installation's service account key
func serviceAccountSpec(ctx context.Context, bindings types.CredentialBindings) (providerSpec, error) {
	cred, err := decodeCredential(bindings, serviceAccountCredential)
	if err != nil {
		return providerSpec{}, err
	}

	creds, err := serviceAccountCredentials(ctx, cred.ServiceAccountKey)
	if err != nil {
		return providerSpec{}, err
	}

	return gcsSpec(cred.Bucket, cred.ProjectID, option.WithCredentials(creds)), nil
}

// gcsSpec selects GCS for one bucket with the supplied Google API client options
func gcsSpec(bucket, projectID string, opts ...option.ClientOption) providerSpec {
	return providerSpec{
		provider: storage.GCSProvider,
		config: storage.ProviderConfigs{
			Enabled:     true,
			Bucket:      bucket,
			Credentials: storage.ProviderCredentials{ProjectID: projectID},
		},
		options: []resolver.BuilderOption{resolver.WithGCSOptions(gcs.WithClientOptions(opts...))},
	}
}

// serviceAccountCredentials parses and validates a service account key
func serviceAccountCredentials(ctx context.Context, rawKey string) (*google.Credentials, error) {
	key := normalizeServiceAccountKey(rawKey)
	if key == "" {
		return nil, ErrServiceAccountKeyInvalid
	}

	creds, err := google.CredentialsFromJSONWithType(ctx, []byte(key), google.ServiceAccount, gcsScope)
	if err != nil {
		return nil, ErrServiceAccountKeyInvalid
	}

	return creds, nil
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

// workloadIdentitySpec selects GCS authenticating through workload identity federation
func workloadIdentitySpec(ctx context.Context, req types.ClientBuildRequest) (providerSpec, error) {
	cred, err := decodeCredential(req.Credentials, workloadIdentityCredential)
	if err != nil {
		return providerSpec{}, err
	}

	source, err := federationSource(ctx, req, cred)
	if err != nil {
		return providerSpec{}, err
	}

	return gcsSpec(cred.Bucket, cred.ProjectID, option.WithTokenSource(source)), nil
}

// federationSource builds the token source, impersonating a service account when configured
func federationSource(ctx context.Context, req types.ClientBuildRequest, cred WorkloadIdentityCredentialSchema) (oauth2.TokenSource, error) {
	if cred.ProjectNumber == "" {
		return nil, ErrProjectNumberRequired
	}

	federated, err := auth.FederatedTokenSource(ctx, req, auth.FederationSpec{
		Audience: workloadIdentityAudience(cred.ProjectNumber),
		Scopes:   []string{gcsScope},
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
		Scopes:          []string{gcsScope},
	}, option.WithTokenSource(federated))
}
