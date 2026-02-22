package gcpscc

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	cloudscc "cloud.google.com/go/securitycenter/apiv2"
	"github.com/samber/lo"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/impersonate"
	"google.golang.org/api/option"
	stsv1 "google.golang.org/api/sts/v1"

	"github.com/theopenlane/core/common/integrations/auth"
	"github.com/theopenlane/core/common/integrations/config"
	"github.com/theopenlane/core/common/integrations/types"
	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/internal/integrations/providers"
	"github.com/theopenlane/core/pkg/logx"
)

// TypeGCPSCC identifies the GCP Security Command Center provider.
const TypeGCPSCC = types.ProviderType("gcp_scc")

const (
	workloadGrantType        = "urn:ietf:params:oauth:grant-type:token-exchange"
	requestedTokenType       = "urn:ietf:params:oauth:token-type:access_token" //nolint:gosec
	defaultScope             = "https://www.googleapis.com/auth/cloud-platform"
	defaultSubjectTokenType  = "urn:ietf:params:oauth:token-type:id_token" //nolint:gosec
	projectScopeAll          = "all"
	projectScopeSpecific     = "specific"
	subjectTokenAttr         = "subject_token"
	subjectTokenTypeAttr     = "subject_token_type"
	minImpersonationLifetime = time.Minute
	maxImpersonationLifetime = time.Hour
	defaultImpersonationLife = time.Hour
)

// ClientSecurityCenter identifies the SCC client descriptor.
const ClientSecurityCenter types.ClientName = "securitycenter.v2"

const securityCenterDescription = "Google Cloud Security Command Center v2 client"

var _ types.ClientProvider = (*Provider)(nil)

// Provider implements the GCP SCC integration using workload identity federation.
type Provider struct {
	// BaseProvider holds shared provider metadata
	providers.BaseProvider
	defaults  workloadDefaults
	stsClient func(ctx context.Context) (*stsv1.Service, error)
}

type workloadDefaults struct {
	scopes            []string
	audience          string
	targetServiceAcct string
	subjectTokenType  string
}

// Builder returns the GCP SCC provider builder.
func Builder() providers.Builder {
	return providers.BuilderFunc{
		ProviderType: TypeGCPSCC,
		BuildFunc: func(_ context.Context, spec config.ProviderSpec) (providers.Provider, error) {
			if spec.GoogleWorkloadIdentity == nil {
				return nil, providers.ErrSpecWorkloadIdentityRequired
			}

			return newProvider(spec), nil
		},
	}
}

// newProvider constructs the GCP SCC provider instance
func newProvider(spec config.ProviderSpec) *Provider {
	defaultScopes := append([]string(nil), spec.GoogleWorkloadIdentity.Scopes...)
	if len(defaultScopes) == 0 {
		defaultScopes = []string{defaultScope}
	}

	subjectTokenType := spec.GoogleWorkloadIdentity.SubjectTokenType
	if subjectTokenType == "" {
		subjectTokenType = defaultSubjectTokenType
	}
	audience := spec.GoogleWorkloadIdentity.Audience
	targetServiceAcct := spec.GoogleWorkloadIdentity.TargetServiceAccount

	defaults := workloadDefaults{
		scopes:            defaultScopes,
		audience:          audience,
		targetServiceAcct: targetServiceAcct,
		subjectTokenType:  subjectTokenType,
	}

	caps := types.ProviderCapabilities{
		SupportsRefreshTokens: false,
		SupportsClientPooling: true,
		SupportsMetadataForm:  len(spec.CredentialsSchema) > 0,
	}

	return &Provider{
		BaseProvider: providers.NewBaseProvider(TypeGCPSCC, caps, nil, nil),
		defaults:     defaults,
		stsClient: func(ctx context.Context) (*stsv1.Service, error) {
			return stsv1.NewService(ctx)
		},
	}
}

// BeginAuth is not applicable for workload identity flows; callers should rely on declarative metadata.
func (p *Provider) BeginAuth(_ context.Context, _ types.AuthContext) (types.AuthSession, error) {
	return nil, ErrBeginAuthNotSupported
}

// Mint exchanges stored workload identity metadata for short-lived Google credentials and persists the updated payload.
func (p *Provider) Mint(ctx context.Context, subject types.CredentialSubject) (types.CredentialPayload, error) {
	meta, err := metadataFromPayload(subject.Credential)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	meta = meta.withDefaults(p.defaults)

	if meta.ServiceAccountKey != "" {
		if _, err := serviceAccountCredentials(ctx, meta.ServiceAccountKey, meta.Scopes); err != nil {
			return types.CredentialPayload{}, err
		}

		providerData, err := auth.PersistMetadata(subject.Credential.Data.ProviderData, meta)
		if err != nil {
			return types.CredentialPayload{}, err
		}

		builder := types.NewCredentialBuilder(TypeGCPSCC).With(
			types.WithCredentialKind(types.CredentialKindMetadata),
			types.WithCredentialSet(models.CredentialSet{
				ProviderData: providerData,
			}),
		)

		return builder.Build()
	}

	var subjectToken, tokenType string
	subjectToken, tokenType, err = resolveSubjectToken(subject, meta, p.defaults)
	if err != nil {
		return types.CredentialPayload{}, err
	}
	meta.SubjectToken = types.TrimmedString(subjectToken)

	accessToken, err := p.mintWorkloadToken(ctx, meta, subjectToken, tokenType)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	providerData, err := auth.PersistMetadata(subject.Credential.Data.ProviderData, meta)
	if err != nil {
		return types.CredentialPayload{}, err
	}

	builder := types.NewCredentialBuilder(TypeGCPSCC).With(
		types.WithCredentialKind(types.CredentialKindWorkload),
		types.WithOAuthToken(accessToken),
		types.WithCredentialSet(models.CredentialSet{
			ProviderData: providerData,
		}),
	)

	return builder.Build()
}

// ClientDescriptors returns the client builders exposed by this provider.
func (p *Provider) ClientDescriptors() []types.ClientDescriptor {
	return []types.ClientDescriptor{
		{
			Provider:    TypeGCPSCC,
			Name:        ClientSecurityCenter,
			Description: securityCenterDescription,
			Build:       buildSecurityCenterClient,
			ConfigSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"findingFilter": map[string]any{
						"type":        "string",
						"description": "Optional SCC findings filter overriding the stored metadata.",
					},
					"sourceId": map[string]any{
						"type":        "string",
						"description": "Optional SCC source override (e.g. organizations/123/sources/456).",
					},
					"sourceIds": map[string]any{
						"type":        "array",
						"description": "Optional SCC source overrides for fan-out collection.",
						"items": map[string]any{
							"type": "string",
						},
					},
				},
			},
		},
	}
}

// mintWorkloadToken exchanges subject tokens for an impersonated access token
func (p *Provider) mintWorkloadToken(ctx context.Context, meta credentialMetadata, subjectToken, subjectTokenType string) (*oauth2.Token, error) {
	stsSvc, err := p.stsClient(ctx)
	if err != nil {
		return nil, ErrSTSInit
	}

	scope := strings.Join(meta.Scopes, " ")
	if scope == "" {
		scope = defaultScope
	}

	optionsValue, err := buildSTSOptions(meta)
	if err != nil {
		return nil, err
	}

	req := &stsv1.GoogleIdentityStsV1ExchangeTokenRequest{
		Audience:           meta.Audience.String(),
		GrantType:          workloadGrantType,
		RequestedTokenType: requestedTokenType,
		Scope:              scope,
		SubjectToken:       subjectToken,
		SubjectTokenType:   subjectTokenType,
		Options:            optionsValue,
	}

	resp, err := stsSvc.V1.Token(req).Context(ctx).Do()
	if err != nil {
		return nil, ErrWorkloadIdentityExchange
	}

	baseToken := &oauth2.Token{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
		Expiry:      time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second),
	}

	if resp.ExpiresIn == 0 {
		baseToken.Expiry = time.Time{}
	}

	staticSource := oauth2.StaticTokenSource(baseToken)

	lifetime := clampLifetime(meta.tokenLifetime())
	cfg := impersonate.CredentialsConfig{
		TargetPrincipal: meta.ServiceAccountEmail.String(),
		Scopes:          meta.Scopes,
		Lifetime:        lifetime,
	}

	ts, err := impersonate.CredentialsTokenSource(ctx, cfg, option.WithTokenSource(staticSource))
	if err != nil {
		return nil, ErrImpersonateServiceAccount
	}

	token, err := ts.Token()
	if err != nil {
		return nil, ErrImpersonatedTokenFetch
	}

	return token, nil
}

// buildSecurityCenterClient builds the SCC client using stored credentials
func buildSecurityCenterClient(ctx context.Context, payload types.CredentialPayload, _ map[string]any) (any, error) {
	meta, err := metadataFromPayload(payload)
	if err != nil {
		return nil, err
	}

	clientOpts, err := securityCenterClientOptions(ctx, meta, payload.Token)
	if err != nil {
		return nil, err
	}

	opts := append([]option.ClientOption{}, clientOpts...)
	if meta.ProjectID != "" {
		opts = append(opts, option.WithQuotaProject(meta.ProjectID.String()))
	}

	client, err := cloudscc.NewClient(ctx, opts...)
	if err != nil {
		return nil, ErrSecurityCenterClientCreate
	}

	return client, nil
}

// securityCenterClientOptions builds client options based on available credentials
func securityCenterClientOptions(ctx context.Context, meta credentialMetadata, token *oauth2.Token) ([]option.ClientOption, error) {
	logger := logx.FromContext(ctx).With().
		Str("provider", string(TypeGCPSCC)).
		Str("projectId", meta.ProjectID.String()).
		Logger()

	if meta.ServiceAccountKey != "" {
		creds, err := serviceAccountCredentials(ctx, meta.ServiceAccountKey, meta.Scopes)
		if err != nil {
			logger.Error().Err(err).Msg("gcpscc: failed to parse service account key for client")
			return nil, err
		}

		logger.Debug().Msg("gcpscc: building client with service account key")
		return []option.ClientOption{
			option.WithCredentials(creds),
		}, nil
	}

	if token != nil {
		accessEmpty := token.AccessToken == ""
		logger = logger.With().
			Bool("tokenEmpty", accessEmpty).
			Time("tokenExpiry", token.Expiry).
			Str("tokenType", token.Type()).
			Logger()
		logger.Debug().Msg("gcpscc: building client with minted oauth token")
		return []option.ClientOption{
			option.WithTokenSource(oauth2.StaticTokenSource(token)),
		}, nil
	}

	logger.Error().Msg("gcpscc: service account key and oauth token missing for client build")
	return nil, ErrAccessTokenMissing
}

// serviceAccountCredentials parses and validates a service account key
func serviceAccountCredentials[T ~string](ctx context.Context, rawKey T, scopes []string) (*google.Credentials, error) {
	key := auth.NormalizeServiceAccountKey(string(rawKey))
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

// credentialMetadata captures the persisted SCC metadata supplied during activation.
type credentialMetadata struct {
	// ProjectID is the GCP project identifier
	ProjectID types.TrimmedString `json:"projectId,omitempty"`
	// OrganizationID is the GCP organization identifier
	OrganizationID types.TrimmedString `json:"organizationId,omitempty"`
	// ProjectScope controls whether collection runs across all accessible projects or a specific set
	ProjectScope types.LowerString `json:"projectScope,omitempty"`
	// ProjectIDs lists explicit project IDs used when projectScope is specific
	ProjectIDs []string `json:"projectIds,omitempty"`
	// WorkloadIdentityProvider is the workload identity provider resource name
	WorkloadIdentityProvider types.TrimmedString `json:"workloadIdentityProvider,omitempty"`
	// Audience is the STS audience used for token exchange
	Audience types.TrimmedString `json:"audience,omitempty"`
	// ServiceAccountEmail is the target service account to impersonate
	ServiceAccountEmail types.TrimmedString `json:"serviceAccountEmail,omitempty"`
	// SourceID is the SCC source identifier used for findings
	SourceID types.TrimmedString `json:"sourceId,omitempty"`
	// SourceIDs optionally lists multiple SCC source identifiers for fan-out collection
	SourceIDs []string `json:"sourceIds,omitempty"`
	// Scopes lists OAuth scopes requested for access tokens
	Scopes []string `json:"scopes,omitempty"`
	// TokenLifetime configures the impersonation token lifetime
	TokenLifetime types.TrimmedString `json:"tokenLifetime,omitempty"`
	// AudienceHint provides an optional audience hint for metadata
	AudienceHint types.TrimmedString `json:"audienceHint,omitempty"`
	// WorkloadPoolProject identifies the billing project for STS exchanges
	WorkloadPoolProject types.TrimmedString `json:"workloadPoolProject,omitempty"`
	// FindingFilter holds the default SCC findings filter
	FindingFilter types.TrimmedString `json:"findingFilter,omitempty"`
	// SubjectToken stores the subject token for STS exchange
	SubjectToken types.TrimmedString `json:"subjectToken,omitempty"`
	// ServiceAccountKey stores the service account key JSON
	ServiceAccountKey types.TrimmedString `json:"serviceAccountKey,omitempty"`
}

// withDefaults applies provider defaults to missing metadata values
func (m credentialMetadata) withDefaults(defaults workloadDefaults) credentialMetadata {
	result := m
	if len(result.Scopes) == 0 {
		result.Scopes = append([]string(nil), defaults.scopes...)
	}
	if result.ServiceAccountEmail == "" {
		result.ServiceAccountEmail = types.TrimmedString(defaults.targetServiceAcct)
	}
	if result.Audience == "" {
		result.Audience = types.TrimmedString(defaults.audience)
	}
	return result
}

// tokenLifetime parses the configured token lifetime
func (m credentialMetadata) tokenLifetime() time.Duration {
	if m.TokenLifetime == "" {
		return 0
	}
	dur, err := time.ParseDuration(m.TokenLifetime.String())
	if err != nil {
		return 0
	}

	return dur
}

// metadataFromPayload decodes provider metadata from the credential payload
func metadataFromPayload(payload types.CredentialPayload) (credentialMetadata, error) {
	if len(payload.Data.ProviderData) == 0 {
		return credentialMetadata{}, ErrCredentialMetadataRequired
	}

	meta, err := auth.ExtractMetadata[credentialMetadata](payload)
	if err != nil {
		return credentialMetadata{}, ErrMetadataDecode
	}

	return meta.applyDefaults(), nil
}

// resolveSubjectToken selects the subject token and token type for STS exchange
func resolveSubjectToken(subject types.CredentialSubject, meta credentialMetadata, defaults workloadDefaults) (token string, tokenType string, err error) {
	if attr := subject.Attributes[subjectTokenAttr]; attr != "" {
		token = attr
	} else {
		token = meta.SubjectToken.String()
	}

	if token == "" {
		return "", "", ErrSubjectTokenRequired
	}

	if attrType := subject.Attributes[subjectTokenTypeAttr]; attrType != "" {
		tokenType = attrType
	} else {
		tokenType = defaults.subjectTokenType
	}

	return token, tokenType, nil
}

// buildSTSOptions encodes STS options for workload pool billing
func buildSTSOptions(meta credentialMetadata) (string, error) {
	if meta.WorkloadPoolProject == "" {
		return "", nil
	}

	payload := map[string]string{
		"userProject": meta.WorkloadPoolProject.String(),
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", ErrSTSOptionsEncode
	}

	return string(encoded), nil
}

// clampLifetime bounds token lifetimes to supported ranges
func clampLifetime(value time.Duration) time.Duration {
	if value <= 0 {
		return defaultImpersonationLife
	}
	return lo.Clamp(value, minImpersonationLifetime, maxImpersonationLifetime)
}

// applyDefaults fills in fallback values, deduplicates slice fields, and normalizes the service account key.
func (m credentialMetadata) applyDefaults() credentialMetadata {
	normalized := m
	if normalized.ProjectScope == "" {
		normalized.ProjectScope = types.LowerString(projectScopeAll)
	}
	normalized.ProjectIDs = types.NormalizeStringSlice(normalized.ProjectIDs)
	normalized.SourceIDs = types.NormalizeStringSlice(normalized.SourceIDs)
	normalized.ServiceAccountKey = types.TrimmedString(auth.NormalizeServiceAccountKey(normalized.ServiceAccountKey.String()))
	normalized.Scopes = types.NormalizeStringSlice(normalized.Scopes)
	return normalized
}

